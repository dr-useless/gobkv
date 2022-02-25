gobkv
-----

# KV Storage over TCP, using Go's net/rpc package
Minimalisic, highly performant key-value storage, written in Go.

# Usage
1. Install `go install github.com/dr-useless/gobkv`
2. (Optional) Define config.json (see configuration)
3. Start `./gobkv -c config.json`

## Config
```json
{
  "Port": 8100,
  "AuthSecret": "arandomstring",
  "CertFile": "path/to/x509/cert.pem",
  "KeyFile": "path/to/x509/key.pem",
  "Persist": true,
  "PartCount": 10,
  "PartDir": "parts",
  "PartWritePeriod": 10,
  "ExpiryScanPeriod": 10
}
```
Unit of time is one second (for PartWritePeriod).

## Play
1. Install CLI tool, gobler
  `go install github.com/dr-useless/gobler`
2. Bind to your gobkv instance
  `gobler bind [NETWORK] [ADDRESS] --a [AUTHSECRET]`
3. Call set, get, del, or list
```bash
./gobler set coffee beans
status: OK
./gobler get coffee
beans
```

# In progress
- Replication

# To do
- Re-partitioning
- Test membership using Bloom filter before GET

# Partitions
To reduce load on the file system & and decrease blocking, the dataset is split across the configured number of partitions (parts). When a key is written to or deleted, the target partition is flagged as changed.

Watchdog periodically writes all changed parts to the file system.

## Key:Partition mapping
Distance from key to a partition is calculated as:
```go
d := hash(key) ^ partitionID
```
The `^` represents XOR.

Target partition ID is the one with smallest distance value.

## Re-partitioning (to do)
Each time the partition list is loaded, it must be compared to the configured partition count. If they do not match, a re-partitioning process must occur before serving connections.

1. Create new partition list in sub-directory
2. Create new Store
3. For each current part, re-map all keys to their new part
4. Write each part after all keys are re-mapped

# Key expiry
The expires time is evaluated periodically in a separate goroutine. The period between scans can be configured using `ExpiryScanPeriod`, giving a number of seconds.

The scan is done in a (mostly) non-blocking way. The partition's write lock is held only while deleting each expired key. Currently, this is done on a per-key basis for simplicity.

# Protocol
Framing & serialization is mostly handled by [chamux](https://github.com/intob/chamux).

## Frame
```
<GOB ENCODED MSG>+END
```

### Frame body
The frame body is a Gob-encoded struct; `protocol.Msg`.
```go
type Msg struct {
	Op      byte
	Status  byte
	Key     string
	Value   []byte
	Expires int64
	Keys    []string
}
```

## Op codes
| Byte | Meaning |
|------|---------|
| 0x01 | Close   |
| 0x02 | Auth    |
| 0x10 | Ping    |
| 0x11 | Pong    |
| 0x20 | Get     |
| 0x30 | Set     |
| 0x31 | SetAck  |
| 0x40 | Del     |
| 0x41 | DelAck  |
| 0x50 | List    |

## Status codes
| Byte | Rune | Meaning      |
|------|------|--------------|
| 0x21 | !    | Error        |
| 0x2F | /    | Unauthorized |
| 0x30 | 0    | NotFound     |
| 0x5F | _    | OK           |

## Endianness
Big endian.