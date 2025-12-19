package eventstore

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"sync"

	"google.golang.org/protobuf/proto"

	"github.com/chuxorg/chux-agent-mesh/meshapi/meshpb"
)

const (
	recordMagic   uint32 = 0x504C4558 // "PLEX"
	recordVersion uint16 = 1
	headerSize           = 12
	checksumSize         = 4
)

var (
	// ErrClosed is returned when Append is invoked on a closed log.
	ErrClosed = errors.New("eventstore: append log closed")
	// ErrCorruptRecord indicates the log contains a record with an invalid magic,
	// version, or checksum.
	ErrCorruptRecord = errors.New("eventstore: corrupt record detected")
	// ErrPartialRecord indicates the log terminates mid-record, which usually
	// happens when a crash interrupts an append.
	ErrPartialRecord = errors.New("eventstore: partial record detected")
	// ErrUnsupportedVersion is returned if a record uses a newer, unsupported
	// format version.
	ErrUnsupportedVersion = errors.New("eventstore: unsupported record version")
	// ErrNilEnvelope guards against accidentally appending a nil protobuf.
	ErrNilEnvelope = errors.New("eventstore: cannot append nil envelope")
)

// AppendLog is a durable, append-only log for EventEnvelope records. Each
// append is acknowledged only after the payload is flushed to disk with fsync.
// The log validates all historical records when opened to ensure corruption is
// detected before accepting new writes.
type AppendLog struct {
	mu         sync.Mutex
	file       *os.File
	syncFunc   func() error
	size       int64
	closed     bool
	marshaller proto.MarshalOptions
}

// OpenAppendLog opens (or creates) the append log located at path. Existing
// records are validated before the log is made available for appends.
func OpenAppendLog(path string) (*AppendLog, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR, 0o600)
	if err != nil {
		return nil, err
	}

	l := &AppendLog{
		file:       f,
		syncFunc:   f.Sync,
		marshaller: proto.MarshalOptions{Deterministic: true},
	}

	if err := l.bootstrap(); err != nil {
		_ = f.Close()
		return nil, err
	}

	return l, nil
}

// Append writes env to the tail of the log. The returned offset points to the
// beginning of the newly written record.
func (l *AppendLog) Append(ctx context.Context, env *meshpb.EventEnvelope) (int64, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if env == nil {
		return 0, ErrNilEnvelope
	}

	payload, err := l.marshaller.Marshal(env)
	if err != nil {
		return 0, err
	}

	record := makeRecord(payload)

	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return 0, ErrClosed
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	offset := l.size

	if err := writeAll(l.file, record); err != nil {
		return 0, err
	}
	if err := l.syncFunc(); err != nil {
		return 0, err
	}

	l.size += int64(len(record))
	return offset, nil
}

// Close closes the underlying file. Future appends will fail with ErrClosed.
func (l *AppendLog) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return nil
	}

	l.closed = true
	return l.file.Close()
}

func (l *AppendLog) bootstrap() error {
	info, err := l.file.Stat()
	if err != nil {
		return err
	}

	size := info.Size()
	if err := validateExistingRecords(l.file, size); err != nil {
		return err
	}

	if _, err := l.file.Seek(size, io.SeekStart); err != nil {
		return err
	}

	l.size = size
	return nil
}

func validateExistingRecords(r io.ReaderAt, size int64) error {
	if size == 0 {
		return nil
	}

	var header [headerSize]byte
	offset := int64(0)

	for offset < size {
		if err := readFullAt(r, header[:], offset); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return fmt.Errorf("%w: truncated header at offset %d", ErrPartialRecord, offset)
			}
			return err
		}

		magic := binary.LittleEndian.Uint32(header[0:4])
		if magic != recordMagic {
			return fmt.Errorf("%w: unexpected magic 0x%x at offset %d", ErrCorruptRecord, magic, offset)
		}

		version := binary.LittleEndian.Uint16(header[4:6])
		if version != recordVersion {
			return fmt.Errorf("%w: got %d", ErrUnsupportedVersion, version)
		}

		length := binary.LittleEndian.Uint32(header[8:12])
		recordLen := int64(headerSize) + int64(length) + checksumSize
		if offset+recordLen > size {
			return fmt.Errorf("%w: truncated payload at offset %d", ErrPartialRecord, offset)
		}

		payload := make([]byte, length)
		if err := readFullAt(r, payload, offset+headerSize); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return fmt.Errorf("%w: truncated payload at offset %d", ErrPartialRecord, offset)
			}
			return err
		}

		var checksumBuf [checksumSize]byte
		checksumOffset := offset + headerSize + int64(length)
		if err := readFullAt(r, checksumBuf[:], checksumOffset); err != nil {
			if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
				return fmt.Errorf("%w: truncated checksum at offset %d", ErrPartialRecord, offset)
			}
			return err
		}

		checksum := binary.LittleEndian.Uint32(checksumBuf[:])
		expected := crc32.ChecksumIEEE(payload)
		if checksum != expected {
			return fmt.Errorf("%w: checksum mismatch at offset %d", ErrCorruptRecord, offset)
		}

		offset += recordLen
	}

	return nil
}

func makeRecord(payload []byte) []byte {
	recordLen := headerSize + len(payload) + checksumSize
	buf := make([]byte, recordLen)

	binary.LittleEndian.PutUint32(buf[0:4], recordMagic)
	binary.LittleEndian.PutUint16(buf[4:6], recordVersion)
	binary.LittleEndian.PutUint16(buf[6:8], 0)
	binary.LittleEndian.PutUint32(buf[8:12], uint32(len(payload)))
	copy(buf[headerSize:headerSize+len(payload)], payload)

	checksum := crc32.ChecksumIEEE(payload)
	binary.LittleEndian.PutUint32(buf[len(buf)-checksumSize:], checksum)

	return buf
}

func readFullAt(r io.ReaderAt, buf []byte, off int64) error {
	read := 0
	for read < len(buf) {
		n, err := r.ReadAt(buf[read:], off+int64(read))
		read += n
		if err != nil {
			if errors.Is(err, io.EOF) {
				if read == len(buf) {
					return nil
				}
				return io.EOF
			}
			return err
		}
	}
	return nil
}

func writeAll(w io.Writer, data []byte) error {
	for len(data) > 0 {
		n, err := w.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}
