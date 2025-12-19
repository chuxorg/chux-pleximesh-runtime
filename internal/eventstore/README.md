# Event Store Append Log

The `AppendLog` implements the local, durable event stream for the runtime
message bus. Its guarantees:

- Append-only writes serialized via a single file descriptor lock.
- Deterministic, protobuf-based record encoding (MarshalOptions.Deterministic).
- An fsync is issued before acknowledging each append.
- Log files are validated on open; corruption and partial tails fail fast.

## Record Format

Each record is laid out as follows (all values are little endian):

| Offset | Size | Description |
| --- | --- | --- |
| 0 | 4 | Magic number `0x504C4558` ("PLEX") |
| 4 | 2 | Format version (`1`) |
| 6 | 2 | Flags (currently zero) |
| 8 | 4 | Serialized `meshpb.EventEnvelope` length |
| 12 | _N_ | Deterministically marshaled envelope bytes |
| 12+_N_ | 4 | CRC32 checksum of the payload |

The CRC covers only the envelope payload so tail truncations or bit flips are
detected when the log is opened.
