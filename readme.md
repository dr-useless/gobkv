# RocketKV

Minimalisic, highly performant key-value storage, written in Go.

<img src="https://goreportcard.com/badge/github.com/intob/rocketkv" />

# Usage
1. Install `go install github.com/intob/rocketkv`
2. (Optional) Define config.json (see configuration)
3. Start `./rocketkv -c cfg.json`

## Config
This project uses [Viper](https://github.com/spf13/viper).

Unless you explicitly provide a config file with `-c`, the config file should be named `config.`, with an extension. TOML, YAML & JSON are supported.

The following directories are searched for a config file:
- `/etc/rocketkv`
- `.` (pwd)

```toml
# /tmp/rocketkv/config.toml

# network & auth
network = "tcp"
address = ":8100"
auth = "supersecretsecret,wait,it'sinthereadme"

# general
segments = 16 # make 256 blocks (16 parts * 16 blocks)
buffersize = 2000000 # 2MB
scanperiod = 10

# persistence
persist = true
dir = "/etc/rocketkv"
writeperiod = 10

[tls]
  cert = "path/to/x509/cert.pem"
  key = "path/to/x509/key.pem"
```
For periords, unit of time is one second. I will add support for parsing time strings.

For each part, the number of blocks created is equal to the part count. So, 8 parts will result in 64 blocks.

## Play
1. Install CLI tool, rkteer
  `go install github.com/intob/rkteer`
2. Bind to your rocketkv instance
  `rkteer bind [NETWORK] [ADDRESS] --a [AUTHSECRET]`
3. Call set, get, del, list, or count
```
./rkteer set coffee beans
status: OK
./rkteer get coffee
beans
```

# In progress
- Support for horizontal scaling

# To do
- Re-partitioning
- Test membership using Bloom filter before GET

# Keys
Keys can be specified in two forms; bare & namespaced.

## Bare
A bare key has no namespace prefix. It is simply a string with no `/` path separator.

## Namespaced
Grouping keys is acheived by namespacing. This is done by prefixing the key with a path.
```
 randomnamespaced/examplekey
|---namespace----|---name---|
```
Note that the namespace includes the final `/`.

All keys for a given namespace will land in the same [block](#blocks). This greatly improves performance for collecting & listing multiple keys, because only a single block must be searched.

# Segmentation
To reduce load on the file system & and decrease blocking, the dataset is split into 2 layers. Each layer contains the configured number of segments.

## Partitions
This is the top layer.

When initialising a dataset, the number of partitions (parts) created will be equal to the configured `Parts.Count` property.

Identifying the partition corresponding to a key is the first step to locating (or placing) a key.

## Blocks
Each part is split into blocks. The number of blocks in each part is equal to the number of parts. So 8 parts will result in 64 blocks.

Each block has it's own mutex & map of keys.

When a key is written to or deleted, the parent block is flagged as changed.

If persistence is enabled in the config via `"Parts.Persist": true`, then each block is written to the file system periodically, when changed.

## Partition:Block:Key mapping
Distance from key to a partition or block is calculated using Hamming distance.
If the key contains a namespace, only the namespace is hashed.
```go
d := hash(key) ^ blockId // or partId
```
The lookup process goes as follows:
1. Find closest part
2. Find closest block in part

This 2-step approach scales well for large datasets where many blocks are desired to reduce blocking.

## Re-partitioning (to do)
Each time the partition list is loaded, it must be compared to the configured partition count. If they do not match, a re-partitioning process must occur before serving connections.

1. Create new manifest (partition:block list) in sub-directory
2. Create new Store
3. For each current part, re-map all keys to their new part
4. Write each part after all keys are re-mapped

# Key expiry
The expires time is evaluated periodically. The period between scans can be configured using `ExpiryScanPeriod`, giving a number of seconds.

# Protocol

## Msg
A normal operation is transmitted in the serialized form of `protocol.Msg`.
```go
type Msg struct {
	Op      byte
	Status  byte
	Key     string
	Value   []byte
	Expires int64
}
```

## Serialization
```
| 0             | 1             | 2             | 3             |
|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|0 1 2 3 4 5 6 7|
+---------------+---------------+---------------+---------------+
| < OP        > | < STATUS    > | < EXPIRES UNIX UINT64         |
|                                                               |
|                             > | < KEY LEN UINT16            > |
  KEY ...                                                       
  VALUE ...                                                     
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
| 0x5F | _    | OK           |
| 0x2F | /    | StreamEnd    |
| 0x2E | .    | NotFound     |
| 0x21 | !    | Error        |
| 0x23 | #    | Unauthorized |

## Endianness
Big endian

# Scaling (in progress)
The aim is to support horizontal scaling to increase availability & load capacity.

For now, we will assume that each node is aware of every other node by configuration. Dynamic service discovery will follow later. Therefore, adding a node involves updating the configuration of all other nodes.

The current solution involves both the client & server.

## Client
When a client wants to read/write a key, they will execute the following process.

### Read
1. Hash key
2. Calculate closest node using the Rendezvous hash (Hamming)
3. Request key from closest node
4. Fallback to next node, recurring until successful or end of node-list is reached

### Write
1. Hash key
2. Calculate closest 3 nodes using the Rendezvous hash (Hamming)
3. Send the request concurrently to each node
4. The operation can be considered complete when the desired number of nodes have acknowleged the request (sent an OK response). 1 node is not sufficient for consistency. 2 of 3 nodes is sufficient for eventual consistency.
5. If a node does not respond, it can optionally be marked by the client as 'down' for a defined period, to prevent future requests timing-out.

## RocketKV
The solution for eventual consistency is a little simpler for RocketKV. Eventual consistency is acheived by periodically replicating blocks to all other known nodes.

Low-latency & consistency is acheived because the mapping for key:node is deterministic.

For now, the modifed-date is used to determine causality. As long as nodes have somewhat syncronised clocks, this is perfectly adequate.

## Service discovery (later)
Each RocketKV node, and all clients would query a single service or cluster.

The single role of this service is to tell clients & RocketKV nodes which nodes currently exist, and their health.

### Updates
Adding a node to the RocketKV network is as simple as spinning it up & making sure that this service knows about it.

A simple solution for automating this would be to include a key-pair in the configuration of each node. A new node can then securely notify this service of it's presence.

An even simpler (but less secure) method would be the use of a shared secret.