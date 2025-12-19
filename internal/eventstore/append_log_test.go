package eventstore

import (
	"context"
	"encoding/binary"
	"errors"
	"hash/crc32"
	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"

	"github.com/chuxorg/chux-agent-mesh/meshapi/meshpb"
)

func TestAppendWaitsForSync(t *testing.T) {
	path := filepath.Join(t.TempDir(), "append.log")
	log, err := OpenAppendLog(path)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	t.Cleanup(func() {
		_ = log.Close()
	})

	block := make(chan struct{})
	syncCalled := make(chan struct{})
	log.syncFunc = func() error {
		close(syncCalled)
		<-block
		return nil
	}

	env := &meshpb.EventEnvelope{EventId: "sync-test"}
	done := make(chan error, 1)
	go func() {
		_, err := log.Append(context.Background(), env)
		done <- err
	}()

	select {
	case <-syncCalled:
	case <-time.After(time.Second):
		t.Fatal("append did not invoke fsync")
	}

	select {
	case err := <-done:
		t.Fatalf("append returned before fsync finished: %v", err)
	case <-time.After(50 * time.Millisecond):
	}

	close(block)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("append failed: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("append did not finish after fsync released")
	}
}

func TestAppendOrdering(t *testing.T) {
	path := filepath.Join(t.TempDir(), "ordering.log")
	log, err := OpenAppendLog(path)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	t.Cleanup(func() { _ = log.Close() })

	want := []*meshpb.EventEnvelope{
		{EventId: "1", EventType: "alpha"},
		{EventId: "2", EventType: "beta"},
		{EventId: "3", EventType: "gamma"},
	}

	var offsets []int64
	for _, env := range want {
		off, err := log.Append(context.Background(), env)
		if err != nil {
			t.Fatalf("append %s: %v", env.EventId, err)
		}
		offsets = append(offsets, off)
	}

	if !isStrictlyIncreasing(offsets) {
		t.Fatalf("offsets not strictly increasing: %v", offsets)
	}

	got := readAllEnvelopes(t, path)
	if len(got) != len(want) {
		t.Fatalf("got %d records, want %d", len(got), len(want))
	}

	for i := range want {
		if !proto.Equal(got[i], want[i]) {
			t.Fatalf("record %d mismatch: got %v want %v", i, got[i], want[i])
		}
	}
}

func TestDetectsCorruption(t *testing.T) {
	path := filepath.Join(t.TempDir(), "corrupt.log")
	log, err := OpenAppendLog(path)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	if _, err := log.Append(context.Background(), &meshpb.EventEnvelope{EventId: "bad"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() == 0 {
		t.Fatal("log is empty")
	}
	if _, err := f.WriteAt([]byte{0xFF}, info.Size()-1); err != nil {
		t.Fatalf("corrupt: %v", err)
	}
	_ = f.Close()

	_, err = OpenAppendLog(path)
	if !errors.Is(err, ErrCorruptRecord) {
		t.Fatalf("expected ErrCorruptRecord, got %v", err)
	}
}

func TestDetectsPartialRecord(t *testing.T) {
	path := filepath.Join(t.TempDir(), "partial.log")
	log, err := OpenAppendLog(path)
	if err != nil {
		t.Fatalf("open log: %v", err)
	}
	if _, err := log.Append(context.Background(), &meshpb.EventEnvelope{EventId: "partial"}); err != nil {
		t.Fatalf("append: %v", err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("close: %v", err)
	}

	f, err := os.OpenFile(path, os.O_RDWR, 0)
	if err != nil {
		t.Fatalf("open raw: %v", err)
	}
	info, err := f.Stat()
	if err != nil {
		t.Fatalf("stat: %v", err)
	}
	if info.Size() <= checksumSize {
		t.Fatal("log too small to truncate")
	}
	if err := f.Truncate(info.Size() - 2); err != nil {
		t.Fatalf("truncate: %v", err)
	}
	_ = f.Close()

	_, err = OpenAppendLog(path)
	if !errors.Is(err, ErrPartialRecord) {
		t.Fatalf("expected ErrPartialRecord, got %v", err)
	}
}

func isStrictlyIncreasing(v []int64) bool {
	for i := 1; i < len(v); i++ {
		if v[i] <= v[i-1] {
			return false
		}
	}
	return true
}

func readAllEnvelopes(t *testing.T, path string) []*meshpb.EventEnvelope {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read log: %v", err)
	}

	var records []*meshpb.EventEnvelope
	offset := 0
	for offset < len(data) {
		if offset+headerSize > len(data) {
			t.Fatalf("truncated header at offset %d", offset)
		}

		header := data[offset : offset+headerSize]
		length := int(binary.LittleEndian.Uint32(header[8:12]))
		payloadStart := offset + headerSize
		payloadEnd := payloadStart + length
		checksumStart := payloadEnd
		checksumEnd := checksumStart + checksumSize

		if checksumEnd > len(data) {
			t.Fatalf("truncated record at offset %d", offset)
		}

		payload := data[payloadStart:payloadEnd]
		expected := crc32.ChecksumIEEE(payload)
		checksum := binary.LittleEndian.Uint32(data[checksumStart:checksumEnd])
		if checksum != expected {
			t.Fatalf("checksum mismatch at offset %d", offset)
		}

		var env meshpb.EventEnvelope
		if err := proto.Unmarshal(payload, &env); err != nil {
			t.Fatalf("unmarshal at offset %d: %v", offset, err)
		}
		records = append(records, proto.Clone(&env).(*meshpb.EventEnvelope))
		offset = checksumEnd
	}

	return records
}
