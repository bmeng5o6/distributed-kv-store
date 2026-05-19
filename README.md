# distributed-kv-store
## Single Server

Built:
- In-memory KV Store with concurrent read and writes. Safe with mutual exclusion to prevent overwriting or corruption
- Write-ahead log (WAL) that survives server crashes to maintain data quality and consistency
- HTTP server with simple GET, PUSH, DELETE functions