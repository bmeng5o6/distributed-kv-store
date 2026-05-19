# distributed-kv-store
## Single Server

Built:
- In-memory KV Store with concurrent read and writes. Safe with mutual exclusion to prevent overwriting or corruption
- Write-ahead log (WAL) that survives server crashes to maintain data quality and consistency
- HTTP server with simple GET, PUSH, DELETE functions

## Hash Ring

Implemented a consistent hash ring, complete with testing. The ring is comprised of many nodes which own a slice of the key space based on the hashed key's value in comparison to the nodes. 
This allows multiple nodes (servers) to exist, which can own different keys, depending on the key's value. 
Added 50 virtual nodes per server, which are duplications of original nodes to spread out distribution of keys on the nodes, so one node won't have too many values, and thus a disproportionate number of queries on a single server. Originally, I'd run into an issue of having too many keys on one node (70%, which I discovered through `ring_test.go`), but adding the virtual nodes allowed the keys to be more evenly spread. 
I used the FNV hash function for the ring, as it was designed for hashing. 