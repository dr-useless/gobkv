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
  "ShardCount": 10,
  "ShardDir": "shards"
}
```

## Play
1. Clone CLI tool, gobler
  `git clone https://github.com/dr-useless/gobler`
2. Bind to your gobkv instance
  `./gobler bind 127.0.0.1:8100 --auth [your_secret]`
3. Call RPCs
  - `./gobler set this isAwesome`
  - `./gobler get this`

# To do
- Replication (of master)
- Expiring keys
- Test membership using Bloom filter before GET

# Sharding
To reduce load on the file system & and decrease blocking, the dataset is split across the configured number of shards. When a key is written to or deleted, the target shard is flagged as changed.

Watchdog periodically writes all changed shards to the file system.

## Key:Shard mapping
Distance from key to shard is calculated as:
```go
d := hash(key) ^ shardID
```
The `^` represents XOR.

Target shard ID is the one with smallest distance.
