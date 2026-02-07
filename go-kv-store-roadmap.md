# Go KV Store — Learning Roadmap

A progressive curriculum for building a Redis-like KV store in Go, organized by the concurrency patterns you'll learn at each stage.

---

## Stage 1: The Foundation — Goroutines, Mutexes, and TCP

**Goal:** A working TCP server that handles GET/SET/DEL with concurrent clients.

**Concepts to learn first:**
- Goroutines and why they're cheap (stack growth, M:N scheduling)
- `sync.Mutex` vs `sync.RWMutex` — when to use which
- `sync.WaitGroup` for coordinating goroutine lifetimes
- `net.Listen` / `net.Accept` / `net.Conn` — Go's blocking I/O model
- `context.Context` for cancellation and graceful shutdown
- `bufio.Scanner` for line-oriented TCP reading

**What you build:**
- An in-memory `map[string]string` protected by `RWMutex`
- A TCP server that spawns one goroutine per client connection
- A simple text protocol: `SET key value`, `GET key`, `DEL key`, `PING`
- Signal-based graceful shutdown (SIGINT/SIGTERM -> context cancel -> drain connections)

**Concurrency patterns unlocked:**
- Goroutine-per-connection
- Shared state with read-write locks
- Context-based cancellation
- WaitGroup for connection draining

**Checklist:**
- [ ] Initialize Go module (`go mod init go-kv-store`)
- [ ] Create `store.go` — `Store` struct with `map[string]string` + `sync.RWMutex`
- [ ] Implement `Get(key)`, `Set(key, value)`, `Del(key)` methods
- [ ] Create `server.go` — TCP listener with `net.Listen("tcp", ":6379")`
- [ ] Accept loop: `listener.Accept()` in a loop, spawn `go handleConn(conn)`
- [ ] `handleConn`: use `bufio.Scanner` to read lines, parse commands, call store methods
- [ ] Protocol parsing: split line by spaces, dispatch to GET/SET/DEL/PING
- [ ] Response format: `+OK`, `$value`, `-ERR message`, `+PONG`
- [ ] Add `context.Context` — create with `signal.NotifyContext(ctx, SIGINT, SIGTERM)`
- [ ] Pass context to accept loop — break on `ctx.Done()`
- [ ] Add `sync.WaitGroup` — `wg.Add(1)` per connection, `wg.Done()` when handler exits
- [ ] On shutdown: cancel context, close listener, `wg.Wait()` for in-flight connections
- [ ] Write tests: concurrent GET/SET from multiple goroutines
- [ ] Run `go test -race ./...` — fix any races
- [ ] Manual test: open 2+ `netcat` sessions, SET/GET concurrently, Ctrl+C server

**Resources:**
- [Go Wiki: LearnConcurrency](https://go.dev/wiki/LearnConcurrency)
- [Go by Example: Goroutines](https://gobyexample.com/goroutines), [Mutexes](https://gobyexample.com/mutexes)
- [go-concurrency-exercises](https://github.com/loong/go-concurrency-exercises)

---

## Stage 2: Background Workers — TTL & Expiration

**Goal:** Keys can expire after a time-to-live. A background goroutine cleans them up.

**Concepts to learn first:**
- `time.Ticker` and `time.After` — scheduling periodic work
- The "background worker goroutine" pattern (loop + select + ctx.Done)
- Lazy vs eager expiration (Redis does both — learn why)

**What you build:**
- `SET key value EX seconds` — stores an expiry timestamp alongside the value
- Lazy expiration: `GET` checks if the key is expired before returning it
- Eager expiration: a background goroutine wakes up every N seconds and sweeps expired keys
- The sweeper shuts down cleanly when context is canceled

**Concurrency patterns unlocked:**
- Long-lived worker goroutine with periodic wake-up (`time.Ticker`)
- `select` statement multiplexing (ticker channel vs context done channel)
- Coordinating background work with shared mutable state

**Checklist:**
- [ ] Change store value type: `map[string]string` -> `map[string]entry` where `entry{value string, expiresAt time.Time}`
- [ ] Update `Set()` to accept optional TTL: `Set(key, value string, ttl time.Duration)`
- [ ] Parse `SET key value EX seconds` in command handler — convert seconds to `time.Duration`
- [ ] Lazy expiration: in `Get()`, check `expiresAt`; if expired, delete key and return not-found
- [ ] Eager expiration: `StartSweeper(ctx context.Context, interval time.Duration)`
- [ ] Sweeper goroutine: `ticker := time.NewTicker(interval)`, loop with `select { case <-ticker.C: sweep(); case <-ctx.Done(): return }`
- [ ] `sweep()`: lock store, iterate map, delete expired entries, unlock
- [ ] Wire sweeper into server startup — pass server's context
- [ ] Test: SET with EX, wait, GET returns nil
- [ ] Test: verify sweeper goroutine exits (count goroutines before/after with `runtime.NumGoroutine()`)
- [ ] Run `go test -race ./...`

---

## Stage 3: Durability — Write-Ahead Log

**Goal:** Data survives restarts. Every mutation is logged to disk before being applied.

**Concepts to learn first:**
- Append-only file I/O (`os.OpenFile` with `O_APPEND`)
- `fsync` and the durability vs performance tradeoff
- Write-ahead logging: why you write the log *before* updating memory
- Replay: rebuilding state from the log on startup

**What you build:**
- A WAL file where each line is a serialized SET or DEL operation
- On startup, replay the WAL to rebuild the in-memory map
- Every write takes the mutex, appends to WAL, then updates the map (inside the same critical section)

**Concurrency patterns unlocked:**
- Critical sections that include I/O (the lock spans both disk write and map update)
- Understanding the tension between holding locks and doing slow operations
- This motivates Stage 4's optimization

**Checklist:**
- [ ] Create `wal.go` — `WAL` struct wrapping an `*os.File`
- [ ] `OpenWAL(path string)` — open file with `O_CREATE|O_RDWR|O_APPEND`
- [ ] WAL line format: `SET key value [EX seconds]\n` or `DEL key\n`
- [ ] `wal.Append(op string)` — write line + `file.Sync()` (fsync)
- [ ] `wal.Replay() []Operation` — read file line by line, parse back into operations
- [ ] On server startup: open WAL, replay entries into store (call `store.Set`/`store.Del`)
- [ ] Modify `store.Set()` and `store.Del()` — inside the lock, call `wal.Append()` *before* updating the map
- [ ] Handle replay of TTL entries: if `SET key value EX seconds` was logged, compute if still valid on replay
- [ ] Test: SET keys, kill server, restart, GET returns values
- [ ] Test: DEL a key, restart, key is gone
- [ ] Benchmark: compare write latency/throughput with Stage 1 (use `testing.B`)
- [ ] Run `go test -race ./...`

---

## Stage 4: Performance — Pipelining & Batching

**Goal:** Reduce lock contention and disk I/O by batching operations.

**Concepts to learn first:**
- Buffered channels as work queues
- The batching/coalescing pattern: collect N items or wait T time, then flush
- Pipelining: clients send multiple commands without waiting for each response
- `bufio.Writer` and explicit flushing

**What you build:**
- A write-batch channel: mutations are sent to a channel, a single goroutine drains it in batches
- The batch goroutine acquires the lock once per batch, applies all ops, writes one WAL entry
- Client-side pipelining: read all available commands, execute as batch, write all responses, flush

**Concurrency patterns unlocked:**
- Buffered channel as a work queue
- Batch/coalesce pattern (collect until buffer full OR timeout)
- Backpressure: what happens when the channel is full? (blocking vs dropping)
- Single-writer pattern: only one goroutine mutates state

**Checklist:**
- [ ] Define `WriteOp` struct: `{Op string, Key string, Value string, TTL time.Duration, Result chan error}`
- [ ] Create buffered channel: `writeCh chan WriteOp` (e.g. buffer size 256)
- [ ] Batch writer goroutine: loop draining channel into a slice (up to N ops or T timeout)
- [ ] Drain logic: first `<-writeCh` (blocking), then non-blocking reads until empty or max batch
- [ ] Use `time.After(batchTimeout)` in select for the flush deadline
- [ ] In batch flush: acquire lock once, append all ops to WAL in one write, apply all to map, release lock
- [ ] Send result back on each `WriteOp.Result` channel
- [ ] Modify `store.Set()`/`store.Del()` — send to `writeCh`, block on result channel
- [ ] Client pipelining: wrap conn in `bufio.Writer`, buffer responses, `Flush()` after processing all available commands
- [ ] Read commands in `handleConn`: use `scanner.Scan()` in a loop, collect commands while data is available
- [ ] Test: concurrent writers, verify all ops applied correctly
- [ ] Benchmark: compare throughput vs Stage 3 (should see improvement)
- [ ] Load test: N goroutines doing SET in parallel, measure ops/sec
- [ ] Run `go test -race ./...`

---

## Stage 5: Pub/Sub — Fan-Out & Select

**Goal:** Clients can subscribe to channels and receive messages in real-time.

**Concepts to learn first:**
- Go channels as message-passing primitives (`chan<-` vs `<-chan`)
- Fan-out: one producer, many consumers
- `select` with multiple channel cases
- Non-blocking sends (`select` with `default`) for slow consumers
- Channel closing semantics and `range` over channels

**What you build:**
- `SUBSCRIBE channel` — client enters subscription mode, receives messages
- `PUBLISH channel message` — sends to all subscribers of that channel
- A `PubSub` manager that maintains a map of channel -> subscriber channels
- Connection handler uses `select` to multiplex between incoming commands and subscription messages

**Concurrency patterns unlocked:**
- Fan-out with Go channels
- `select` multiplexing multiple channel sources
- Slow consumer problem: buffered channel fills up — drop message or block publisher?
- Goroutine lifecycle management for subscriber goroutines

**Checklist:**
- [ ] Create `pubsub.go` — `PubSub` struct with `sync.RWMutex` + `map[string]map[uint64]chan string`
- [ ] `Subscribe(channel string) (ch <-chan string, id uint64)` — create buffered chan (e.g. 64), add to map
- [ ] `Unsubscribe(channel string, id uint64)` — remove from map, close the chan
- [ ] `Publish(channel string, message string) int` — fan-out to all subscribers, return count
- [ ] Use non-blocking send: `select { case sub <- msg: default: /* drop for slow consumer */ }`
- [ ] Parse `SUBSCRIBE channel` command — enter subscription mode
- [ ] Subscription mode: spawn goroutine to read from sub channel and write to conn
- [ ] Use `select` in connection handler to multiplex: incoming commands + subscription messages
- [ ] Parse `PUBLISH channel message` — call `pubsub.Publish()`, respond with subscriber count
- [ ] Parse `UNSUBSCRIBE` — call `pubsub.Unsubscribe()`, exit subscription mode
- [ ] Handle client disconnect: unsubscribe from all channels, close sub channels
- [ ] Test: 2 subscribers + 1 publisher, both receive message
- [ ] Test: slow subscriber doesn't block publisher (verify with timeout)
- [ ] Test: unsubscribe + disconnect — no goroutine leaks
- [ ] Run `go test -race ./...`

---

## Stage 6: Distribution — Replication

**Goal:** One leader server replicates writes to follower servers in real-time.

**Concepts to learn first:**
- Leader/follower replication model (how Redis does it)
- Streaming data over TCP between servers
- Reconnection with exponential backoff
- Eventually consistent vs strongly consistent (and where your system falls)
- `io.Pipe` or channels for streaming WAL entries

**What you build:**
- Leader mode: accepts follower connections, streams WAL entries as they happen
- Follower mode: connects to leader, receives WAL stream, applies to local store
- Followers serve read-only GET requests (reject SET/DEL from clients)
- Follower reconnects automatically on disconnect with exponential backoff

**Concurrency patterns unlocked:**
- Producer-consumer across network boundaries
- Exponential backoff with `time.Timer`
- Concurrent readers + single remote writer (follower's store)
- Goroutine supervision: restarting failed goroutines

**Checklist:**
- [ ] Add CLI flags: `--role leader|follower`, `--leader-addr host:port`, `--repl-port port`
- [ ] Create `replication.go` — leader and follower logic
- [ ] Leader: `ReplicationListener` — accepts follower TCP connections on `repl-port`
- [ ] Leader: on follower connect, send full state snapshot (all current key-value pairs)
- [ ] Leader: after snapshot, stream new WAL entries as they happen (hook into WAL append)
- [ ] WAL hook: add `OnAppend(callback func(entry string))` to WAL — leader registers callback to broadcast
- [ ] Leader: maintain list of follower connections, fan-out WAL entries to all
- [ ] Follower: `ConnectToLeader(addr string)` — dial TCP, receive snapshot, apply to store
- [ ] Follower: after snapshot, read WAL stream in a loop, apply each entry to local store
- [ ] Follower: reject SET/DEL from clients with `-ERR READONLY` response
- [ ] Follower: on disconnect, reconnect with exponential backoff (1s, 2s, 4s, ... max 30s)
- [ ] Backoff: use `time.Timer` + jitter, reset on successful reconnect
- [ ] Follower: on reconnect, receive fresh snapshot + resume streaming
- [ ] Handle follower disconnect on leader side: remove from list, clean up goroutine
- [ ] Test: leader SET, follower GET returns value
- [ ] Test: kill follower, SET on leader, restart follower — catches up
- [ ] Test: kill leader — follower keeps serving reads
- [ ] Run `go test -race ./...`

---

## Project Structure

```
go-kv-store/
├── main.go           # CLI flags, wiring, signal handling
├── store.go          # In-memory store with RWMutex
├── server.go         # TCP server, accept loop, connection handler
├── protocol.go       # Command parsing and response formatting
├── wal.go            # Write-ahead log (append, replay, sync)
├── batcher.go        # Write batching goroutine
├── pubsub.go         # Pub/Sub manager
├── replication.go    # Leader/follower replication
├── store_test.go     # Store unit tests
├── server_test.go    # Integration tests (TCP client tests)
├── wal_test.go       # WAL tests
├── batcher_test.go   # Batcher tests
├── pubsub_test.go    # Pub/Sub tests
├── replication_test.go # Replication tests
├── go.mod
└── README.md
```

## Suggested Order of Attack

1. **Stage 1** — Get `store.go` + `server.go` + `protocol.go` + `main.go` working first
2. **Stage 2** — Modify `store.go` to add TTL, add sweeper
3. **Stage 3** — Add `wal.go`, wire into store mutations
4. **Stage 4** — Add `batcher.go`, refactor store writes to go through channel
5. **Stage 5** — Add `pubsub.go`, extend protocol and connection handler
6. **Stage 6** — Add `replication.go`, add CLI flags to `main.go`

## Quick Reference — Testing Commands

```bash
# Run all tests with race detector
go test -race ./...

# Run benchmarks
go test -bench=. -benchmem ./...

# Manual testing with netcat
echo "PING" | nc localhost 6379
echo "SET foo bar" | nc localhost 6379
echo "GET foo" | nc localhost 6379

# Interactive netcat session
nc localhost 6379
# then type commands one per line

# Count goroutines (in test code)
before := runtime.NumGoroutine()
// ... do stuff ...
after := runtime.NumGoroutine()
// assert after <= before
```

## Resources

- [Go Wiki: LearnConcurrency](https://go.dev/wiki/LearnConcurrency)
- [Go by Example: Goroutines](https://gobyexample.com/goroutines), [Mutexes](https://gobyexample.com/mutexes)
- [go-concurrency-exercises](https://github.com/loong/go-concurrency-exercises)
- [Redis internals](https://redis.io/docs/reference/internals/)
- [Bitcask paper](https://riak.com/assets/bitcask-intro.pdf)
- [Raft paper](https://raft.github.io/raft.pdf)
