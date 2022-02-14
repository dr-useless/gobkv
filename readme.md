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
  "PersistFile": "path/to/persistence/file.gob"
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
- Paging of persistence gobs to reduce load on file system for large datasets
- Replication (of master)
- Expiring keys
- Test membership using Bloom filter before GET

# Paging
## Why
Currently, each time the Watchdog writes the dataset to the file system, the entire set is encoded & written. The same on launch, the entire dataset is read from a single file.

While being very simple, the performance of this design suffers for large datasets. It would be much better to segment/shard the data into pages.

## How
### Init
When we create a new dataset, we also create the configured number of pages. Each page has a unique ID, a random 32-bit hash. The filename of each page is `[BASE64URL_HASH].gob`.

### Key:Page mapping
Distance from key to page is calculated as:
```go
d := hash(key) ^ pageID
```
The `^` represents XOR.

Target page ID is the one with smallest distance.

### Write
When a key is written to or deleted, the target page must be flagged for writing by the Watchdog.

### Read
As MVP, all keys will be stored in memory.

An enhancement will be to add an option to cache recently used keys, and leave the rest in file storage. Similar to Redis' virtual memory.