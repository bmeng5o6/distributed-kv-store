# distributed-kv-store
## Single Server

Built:
- In-memory KV Store with concurrent read and writes. Safe with mutual exclusion to prevent overwriting or corruption
- Write-ahead log (WAL) that survives server crashes to maintain data quality and consistency
- HTTP server with simple GET, PUT, DELETE functions

## Hash Ring

Implemented a consistent hash ring, complete with testing. The ring is comprised of many nodes which own a slice of the key space based on the hashed key's value in comparison to the nodes. 
This allows multiple nodes (servers) to exist, which can own different keys, depending on the key's value. 
Added 50 virtual nodes per server, which are duplications of original nodes to spread out distribution of keys on the nodes, so one node won't have too many values, and thus a disproportionate number of queries on a single server. Originally, I'd run into an issue of having too many keys on one node (70%, which I discovered through `ring_test.go`), but adding the virtual nodes allowed the keys to be more evenly spread. 
I used the FNV hash function for the ring, as it was designed for hashing. 

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