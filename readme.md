# gobkv
## KV Storage over TCP, using Go's net/rpc package
Minimalisic, highly performant key-value storage, written in Go.

# Usage
1. Clone `git clone https://github.com/dr-useless/gobkv`
2. (Optional) Define config.json (see configuration)
3. Start `go build; ./gobkv -c config.json`

## Play
1. Clone CLI tool, gobler `git clone https://github.com/dr-useless/gobler`
2. Bind to your gobkv instance `./gobler bind 127.0.0.1:8100 --auth [your_secret]`
3. Call RPCs
  - `./gobler set this isAwesome`
  - `./gobler get this`


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

## 

# To do
- Paging of persistence gobs to reduce file IO load for large datasets
- Replication (of master)
- Expiring keys
- Transactions (BEGIN, SET, DEL, COMMIT)
