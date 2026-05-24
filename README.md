# distributed-kv-store

A distributed key-value store built in Go, supporting multi-node routing, replication, and failure detection. Deployed to AWS EC2 via Docker.

---

## Why Go?

I chose Go because of its concurrency primitives (`goroutines`, `sync.RWMutex`) and the industry's usage of Go in distributed systems, such as with CockroachDB and etcd. The concurrency model maps naturally to a KV store: `RWMutex` allows multiple concurrent readers while blocking writers, which is the right tradeoff for a read-heavy workload.

## Design Decisions

**Leader-based replication over leaderless**: leaderless replication requires vector clocks or last-write-wins to resolve conflicts when two nodes accept writes simultaneously. Leader-based replication avoids this, since one node is the single source of truth per shard, so there are no conflicts to resolve.

**WAL over a database**: a Write Ahead Log is the mechanism databases use internally for durability. I wanted to build it from scratch to better understand how this works, rather than using a database like Postgres or SQLite. 

**FNV over CRC32**: During testing, CRC32 was causing one node to receive 70% of keys. FNV was designed specifically for hash tables and distributes values more evenly across the key space.

**PUT throughput is intentionally bounded by fsync**: I chose to flush to disk before ACKing the client to maintain consistency in data. However, to improve throughput in production, I would implement group commits which would batch multiple writes into a single sync. 

**Immediate replica fallback over waiting for gossip**: Originally, when a leader node was unreachable, the server returned a 502 and waited for gossip to mark the node dead (up to 3 seconds) before routing to a replica. I changed this so that any forward error immediately falls through to the replica. This reduced failover recovery time from ~3s to ~10ms under continuous load.

**Gossip interval of 500ms with 2-strike rule**: Pinging every 500ms means a dead node is marked dead within ~1 second. Using 3 strikes instead would extend this to 1.5 seconds unnecessarily. The tradeoff is that a faster interval may cause more false positives, in that a slower node may be marked as dead, if not responsive within that time limit. 

**Shared HTTP client with 500ms timeout**: All inter-node requests now use a shared `http.Client` with a 500ms timeout rather than `http.DefaultClient`, which has no timeout. This prevents hung nodes from blocking requests indefinitely. A tighter timeout would detect slow nodes faster but risks false positives in skipping nodes.


## Single Server

Built:
- In-memory KV Store with concurrent read and writes. Safe with mutual exclusion to prevent overwriting or corruption
- Write-ahead log (WAL) that survives server crashes to maintain data quality and consistency
- HTTP server with simple GET, PUT, DELETE functions

## Hash Ring

Implemented a consistent hash ring, complete with testing. The ring is comprised of many nodes which own a slice of the key space based on the hashed key's value in comparison to the nodes. 
This allows multiple nodes (servers) to exist, which can own different keys, depending on the key's value. 
Added 50 virtual nodes per server, which are duplications of original nodes to spread out distribution of keys on the nodes, so one node won't have too many values, and thus a disproportionate number of queries on a single server. Originally, I'd run into an issue of having too many keys on one node (70%, which I discovered through `ring_test.go`), but adding the virtual nodes allowed the keys to be more evenly spread. 

## Distributedness
Added multi-node routing with hashing. Each node now forwards requests to the correct owner (`forward` function in `http.go`) based on the hash ring. Clients can query any node and still get the correct response. 

## Replication
Added `replicate`, which can write directly to replicas from a leader node, with the `X-Replication: true` header. This header tells the receiving node to store the value directly without re-routing or replicating again, preventing infinite loops.
Wrote `GetNodes` function to grab multiple nodes (replicas of leader). This is used to check replicas for GET, update replicas for PUT and DELETE.
Together, these allow replicas to hold values that original servers that fail originally held. This means that if servers fall out, values can still be preserved. 
NOTE: The `count` argument on the `GetNodes` function increases the number of replicas per node. 
- One of the issues I ran into was if the leader node that owns a specific key is removed, then the key couldn't be retrieved. To fix this issue, I decided to create a new function `tryForward` that could be used with `HandleGet` that would allow the next node in the nodes list to be checked with `HandleGet` to see if the key is present there. 
- Another issue that was present is the loop condition of `GetNodes` in `ring.go`. A conditional on the index being less than the number of nodes (virtual nodes included) allows all nodes to be visited once, preventing an infinite loop. Duplicate servers are prevented with a seen map. It breaks early once it has collected `count` unique nodes. 

## Failure Detection
Added `gossip.go` file, which introduces the Gossip struct that provides information on the availability of nodes. It runs a goroutine function in `Start()` that continuously checks on the availability of a node's peers, sending a ping to `/health` on each peer. Each peer responsds with `HandleHealth` if it is alive. If a node is dead, it is skipped during routing without waiting for a connection failure or error. This feature helps improve the efficiency of failure detection. 

## Failover Recovery Optimization

The original implementation returned a 502 when the leader was unreachable, relying on gossip to eventually mark the node dead before rerouting. This created a recovery window of up to 3 seconds.

The fix was that in the case of a forward error to immediately try the replica. Gossip still runs in the background to mark the node dead for future routing decisions, but in-flight requests won't wait for it.

| Version | Recovery Time |
|---|---|
| Original (wait for gossip) | ~3s |
| After immediate fallback | ~55ms at 20 req/sec |
| After immediate fallback | ~10ms under continuous load |

## Github Actions
Experimented with adding tests to every push or pull request to the main branch. Ensures quality of code and basic functionality of distributed store. 

## Docker and AWS
Containerized the project using Docker with a two-stage build. The first stage compiles the Go binary, the second stage copies just the binary into a minimal Alpine image. `docker-compose` starts all three nodes with a single command, with nodes communicating over Docker's internal network by service name.

Deployed the cluster to AWS EC2. I SSH'd into an Amazon Linux instance, installed Docker, cloned the repo, and ran `docker-compose up`. The cluster is accessible over the instance's public IP. Verified replication by reading the WAL file inside individual containers using `docker exec`.

## Running locally
```bash
docker-compose up --build
```

## Running on AWS
```bash
ssh -i kv-store-key.pem ec2-user@YOUR_EC2_IP
cd distributed-kv-store
docker-compose up -d
```

## Benchmarks

Run on Apple M5 (`go test -bench=. -benchmem ./internal/store/...`):

| Operation | Throughput | Latency | Memory |
|---|---|---|---|
| PUT (single) | ~250 ops/sec | ~4ms | 53 B/op |
| GET (single) | ~200M ops/sec | 5ns | 0 B/op |
| GET (concurrent, 10 goroutines) | ~10M ops/sec | 98ns | 0 B/op |
| Leader failover (continuous load) | — | ~10ms recovery | — |
| Leader failover (20 req/sec) | — | ~55ms recovery | — |

PUT latency is bounded by `fsync` — each write is flushed to disk before ACKing the client, which is the correct tradeoff for durability over throughput.