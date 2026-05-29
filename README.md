# Redis Server (Go)

A Redis-like in-memory data store built from scratch in Go, implementing the Redis Serialization Protocol (RESP), raw TCP command execution, probabilistic data structures, and a multi-threaded server architecture with I/O multiplexing.

---

## Architecture

```
                        ┌─────────────────────────────────────┐
                        │           TCP Listeners             │
                        │  (SO_REUSEPORT multi-listener loop) │
                        └────────────────┬────────────────────┘
                                         │ round-robin
                        ┌────────────────▼────────────────────┐
                        │           IOHandlers                │
                        │  (kqueue / epoll multiplexing)      │
                        └────────────────┬────────────────────┘
                                         │ FNV key-hash dispatch
                        ┌────────────────▼────────────────────┐
                        │     Key-Partitioned Workers         │
                        │  (each owns its storage shard)      │
                        └────────────────┬────────────────────┘
                                         │
                        ┌────────────────▼────────────────────┐
                        │          In-Memory Store            │
                        │  Dict / TTL / Eviction / Structures │
                        └─────────────────────────────────────┘
```

### Execution Modes

**Single-thread** — one event loop using kqueue/epoll to multiplex all client connections, executing commands on a single command executor.

**Multi-thread** — multiple TCP listeners with `SO_REUSEPORT`, a pool of IOHandler goroutines assigned via round-robin, and key-partitioned workers that each own their own storage shard. Commands are routed to workers by FNV hashing the key, ensuring lock-free concurrent access.

---

## Supported Commands

| Category | Commands |
|---|---|
| String | `GET`, `SET`, `TTL`, `PING` |
| Set | `SADD`, `SREM`, `SMEMBERS`, `SISMEMBER` |
| Sorted Set | `ZADD`, `ZSCORE`, `ZRANK` |
| Bloom Filter | `BF.RESERVE`, `BF.MADD`, `BF.EXISTS` |
| Count-Min Sketch | `CMS.INITBYDIM`, `CMS.INITBYPROB`, `CMS.INCRBY`, `CMS.QUERY` |
| Server | `INFO` |

---

## Data Structures

### Dict (Hash Map)
Core key-value store backed by a Go map. Supports TTL via a separate expiry map, lazy deletion on read, and active expiry cleanup every 100ms sampling 20 keys per cycle.

### B+ Tree (Sorted Set)
Custom B+ Tree implementation backing sorted sets. Supports O(log n) insert/update with recursive leaf and internal node splitting, linked leaf nodes for ordered traversal, and O(n) rank queries via leaf scan.

### Bloom Filter
Space-efficient probabilistic membership structure using Murmur3 double-hashing. Optimal bit array size and hash function count computed from target error rate using standard formulas.

### Count-Min Sketch
Frequency estimation structure using a 2D counter array with independent Murmur3 hash functions per row. Supports `INITBYDIM` (explicit dimensions) and `INITBYPROB` (error rate + probability).

---

## Memory Management

- **TTL expiration** — lazy deletion on read; active cleanup loop runs every 100ms
- **Approximate LRU eviction** — samples N random keys into an eviction pool sorted by `LastAccessTime`, evicts the oldest candidates when key count hits the configured maximum
- **Random eviction** — evicts a configurable ratio of keys at random when the key limit is reached

---

## Getting Started

### Prerequisites
- Go 1.21+

### Run

```bash
# Clone the repo
git clone https://github.com/Anhtran0208/redis-server-intro.git
cd redis-server-intro

# Multi-thread mode (default)
go run cmd/main.go --mode=multi-thread --workers=4 --io-handlers=4 --listeners=3 --port=3000

# Single-thread mode
go run cmd/main.go --mode=single-thread --port=3000
```

### Connect with redis-cli

```bash
redis-cli -p 3000
```

### Configuration Flags

| Flag | Default | Description |
|---|---|---|
| `--mode` | `multi-thread` | Execution mode: `single-thread` or `multi-thread` |
| `--port` | `3000` | Server port |
| `--workers` | `4` | Number of worker goroutines (multi-thread only) |
| `--io-handlers` | `4` | Number of IOHandler goroutines (multi-thread only) |
| `--listeners` | `3` | Number of TCP listeners (multi-thread only) |
| `--max-connections` | `20000` | Maximum concurrent connections |
| `--max-keys` | `10` | Maximum keys before eviction triggers |
| `--eviction-policy` | `allkeys-lru` | `allkeys-lru`, `allkeys-random`, or `noeviction` |
| `--eviction-ratio` | `0.1` | Fraction of keys to evict per cycle |

---

## Example Usage

```bash
# Strings
SET foo bar
GET foo
SET foo bar EX 10
TTL foo

# Sets
SADD myset a b c
SMEMBERS myset
SISMEMBER myset a

# Sorted Sets
ZADD leaderboard 100 alice
ZADD leaderboard 200 bob
ZRANK leaderboard alice
ZSCORE leaderboard bob

# Bloom Filter
BF.RESERVE myfilter 0.01 1000
BF.MADD myfilter alice bob
BF.EXISTS myfilter alice

# Count-Min Sketch
CMS.INITBYPROB mysketch 0.001 0.01
CMS.INCRBY mysketch alice 5
CMS.QUERY mysketch alice
```

---

## Project Structure

```
.
├── cmd/
│   └── main.go                  # Entry point, mode dispatch
├── internal/
│   ├── config/                  # CLI flag parsing and config struct
│   ├── constant/                # RESP constants, server status, expiry config
│   ├── core/
│   │   ├── executor.go          # Command dispatch and execution
│   │   ├── resp.go              # RESP encoder/decoder
│   │   ├── storage.go           # Store wrapper
│   │   ├── worker.go            # Worker goroutine with task channel
│   │   ├── command*.go          # Per-type command implementations
│   │   └── io_multiplexing/     # kqueue (macOS) / epoll (Linux) abstraction
│   ├── data_structure/
│   │   ├── dict.go              # Core hash map with TTL and eviction
│   │   ├── bplustree.go         # B+ Tree for sorted sets
│   │   ├── bloom_filter.go      # Bloom Filter with Murmur3 double-hashing
│   │   ├── cms.go               # Count-Min Sketch
│   │   ├── eviction_pool.go     # LRU eviction candidate pool
│   │   └── simple_set.go        # Hash set
│   └── server/
│       ├── single_thread_server.go
│       ├── multi_thread_server.go
│       ├── io_handler.go        # Per-handler epoll/kqueue event loop
│       └── signal.go            # Graceful shutdown
```
