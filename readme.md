# gobkv
## KV Storage over TCP, using Go's net/rpc package
Minimalisic, highly performant key-value storage, written in Go.

# Usage
1. Clone `git clone https://github.com/dr-useless/gobkv`
2. (Optional) Define config.json (see configuration)
3. Start `go build; ./gobkv -c config.json`

## Config
```json
{
  "Port": 8100,
  "AuthSecret": "a random string",
  "CertFile": "path/to/x509/cert.pem",
  "KeyFile": "path/to/x509/key.pem",
  "Persist": true,
  "PartCount": 10,
  "PartDir": "parts",
  "PartWritePeriod": 10 // seconds
}
```

## Play
1. Clone CLI tool, gobler
  `git clone https://github.com/dr-useless/gobler`
2. Bind to your gobkv instance
  `./gobler bind 127.0.0.1:8100 --auth [your_secret]`
3. Call RPCs
```bash
./gobler set coffee life
_
./gobler get coffee
life
```

# To do
- Replication (of master)
- Expiring keys
- Test membership using Bloom filter before GET
- Re-partitioning

# Partitions
To reduce load on the file system & and decrease blocking, the dataset is split across the configured number of partitions (parts). When a key is written to or deleted, the target partition is flagged as changed.

Watchdog periodically writes all changed parts to the file system.

## Key:Partition mapping
Distance from key to partition is calculated as:
```go
d := hash(key) ^ partitionID
```
The `^` represents XOR.

Target partition ID is the one with smallest distance.

## Re-partitioning (to do)
Each time the partition list is loaded, it must be compared to the configured partition count. If they do not match, a re-partitioning process must occur before serving connections.

1. Create new partition list in sub-directory
2. Create new Store
3. For each current part, re-map all keys to their new part
4. Write each part after all keys are re-mapped
