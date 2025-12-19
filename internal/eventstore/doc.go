// Package eventstore implements the durable append-only log that backs the runtime
// message bus. Each record uses a deterministic wire format:
//
//	0-3   : magic number 0x504C4558 ("PLEX")
//	4-5   : format version (uint16, currently 1)
//	6-7   : reserved flags (uint16, zero)
//	8-11  : serialized protobuf envelope length in bytes (uint32, little endian)
//	12..N : deterministically marshaled meshpb.EventEnvelope payload
//	N..N+4: CRC32 checksum of the payload (IEEE polynomial)
//
// Records are appended atomically, fsync'ed before acknowledging the write, and
// fully validated when the log is opened so corruption and partial tails are
// detected immediately.
package eventstore
