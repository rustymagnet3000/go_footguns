# 05 — a concurrent map write is a *fatal, unrecoverable* crash

## The trap

Two goroutines writing the same built-in `map` at the same time is not "undefined and probably fine" — the Go runtime **actively detects** it and calls `throw("concurrent map writes")`, which **kills the entire process**. It is not a `panic`; it is a `fatal error`, and **`recover()` cannot catch it**.

```go
m := make(map[int]int)
for i := 0; i < 8; i++ {
    go func(id int) {
        for j := 0; j < 10000; j++ {
            m[id] = j // unsynchronised writes → fatal error: concurrent map writes
        }
    }(i)
}
```

So a single unsynchronised map access — reachable by, say, two HTTP requests hitting a shared in-memory cache at once — is a **remote denial-of-service**: it doesn't corrupt a value, it takes the whole server down, and no `defer/recover` in your handler will save it.

## The correct shape

Guard the map with a lock, or use `sync.Map` for genuinely concurrent workloads.

```go
type cache struct {
    mu sync.RWMutex
    m  map[int]int
}

func (c *cache) set(k, v int) { c.mu.Lock(); c.m[k] = v; c.mu.Unlock() }
func (c *cache) get(k int) (int, bool) {
    c.mu.RLock(); defer c.mu.RUnlock()
    v, ok := c.m[k]
    return v, ok
}
```

- **`sync.RWMutex`** — the default answer; cheap, obvious, lets concurrent readers through.
- **`sync.Map`** — purpose-built for high-concurrency read-mostly / disjoint-key maps; avoids the lock but has a clunkier API.
- Run the reproducer both ways:

```
$ go run ./05-concurrent-map-panic
fatal error: concurrent map writes
...goroutine dump...
exit status 2

$ go run -race ./05-concurrent-map-panic
==================
WARNING: DATA RACE
Write at 0x... by goroutine N:
  main.main.func1()
      .../05-concurrent-map-panic/main.go:31 +0x...
==================
```

## Why Go differs from other languages

| | Concurrent modification of a plain hash map |
|---|---|
| **Java** (`HashMap`) | Undefined: may corrupt the structure, throw `ConcurrentModificationException`, or (older JVMs) spin in an **infinite loop** — but it does **not** kill the JVM. |
| **Python** (`dict`) | The GIL serialises bytecode, so individual dict ops are effectively atomic; concurrent writers don't crash the interpreter (logical races still exist). |
| **Go** (`map`) | The runtime **detects** concurrent access and issues a `fatal error` that **terminates the process** and **cannot be recovered**. |

Go deliberately turns "you raced a map" into an immediate, loud, whole-process crash — the designers chose a hard stop over silent corruption. The security consequence is that the failure mode is **availability** (DoS), not data integrity, and it's un-catchable.

## What breaks

- It's a `throw`, not a `panic`: **`recover()` does nothing** (the reproducer proves it — its deferred `recover` never prints).
- It's **non-deterministic**: the reproducer crashes on most runs but occasionally "gets lucky." That's the dangerous part — it can pass local testing and CI a few times, then crash under production concurrency.
- Concurrent **read + write** is caught too, not just write + write. A racing reader is enough.

## How to spot it in review

- A `map` field on a struct shared across goroutines/requests with **no adjacent `sync.Mutex`/`RWMutex`** (or not a `sync.Map`).
- An in-memory cache/registry/counter written from an HTTP handler or any `go func(...)`.
- Any map mutated inside a goroutine where the same map is read/written elsewhere.
- CI that doesn't run `go test -race` — the detector finds these deterministically; without it they hide.

## The fix / the habit

> A built-in `map` is **not** safe for concurrent use. Shared across goroutines → put a `sync.RWMutex` next to it (or use `sync.Map`). And run `go test -race` in CI, because the crash is a fatal, unrecoverable, whole-process DoS you can't guard with `recover()`.

## References

- Go blog — [Go maps in action § Concurrency](https://go.dev/blog/maps) — "Maps are not safe for concurrent use."
- [`sync.RWMutex`](https://pkg.go.dev/sync#RWMutex) · [`sync.Map`](https://pkg.go.dev/sync#Map)
- [Data Race Detector](https://go.dev/doc/articles/race_detector) — `go test -race` / `go run -race`
- Runtime source: `runtime.mapaccess`/`mapassign` set `hashWriting` and `throw("concurrent map writes")` — a `throw`, which is why `recover()` can't catch it
