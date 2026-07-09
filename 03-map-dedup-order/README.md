# 03 — dedup with a map, not nested loops — and don't trust its iteration order

## The trap

Two traps, back to back.

**Trap 1 — the O(n²) nested loop.** Asked to keep the *unique* values from a slice (duplicates thrown away), the reflex from many languages is: for each element, scan what you've already kept to see if it's there.

```go
// O(n²) — the trap
func unique(nums []int) []int {
    var out []int
    for _, n := range nums {
        found := false
        for _, seen := range out { // inner scan of everything kept so far
            if seen == n {
                found = true
                break
            }
        }
        if !found {
            out = append(out, n)
        }
    }
    return out
}
```

It works, but it's quadratic. The idiomatic Go answer is a **map as a set**: O(1) membership, O(n) overall.

**Trap 2 — building the result by ranging the map.** Once you have a map, it's tempting to skip the result slice and just range the keys. But **Go deliberately randomizes map iteration order**, so your unique values come out in a *different order on every run*.

## The correct shape

Map for **membership**, slice for **order**. Append to the slice the first time you see each key:

```go
func unique(nums []int) []int {
    seen := make(map[int]bool)
    out := make([]int, 0, len(nums))
    for _, n := range nums {
        if !seen[n] {              // first time we've seen this value?
            seen[n] = true         // mark as seen
            out = append(out, n)   // record first-seen order
        }
    }
    return out
}
```

- `map[int]bool` is the set: the value just marks presence, so `if !seen[n]` reads cleanly. A missing key returns the zero value `false`, so you never need a separate "does it exist?" check.
- The zero-byte `map[int]struct{}` variant (`seen[n] = struct{}{}`, tested with comma-ok `_, ok := seen[n]`) saves the pointless bool byte — see the empty-struct entry for why `struct{}{}` looks the way it does. `map[int]bool` is just as correct and easier to read.

## What breaks if you range the map for output

```go
// BUG: order is nondeterministic
func uniqueBad(nums []int) []int {
    seen := make(map[int]bool)
    for _, n := range nums {
        seen[n] = true
    }

    out := make([]int, 0, len(seen))
    for n := range seen { // <-- randomized every run
        out = append(out, n)
    }
    return out
}
```

`uniqueBad([]int{3, 1, 3, 2, 1})` returns the set `{1, 2, 3}` — but as `[2 3 1]`, then `[1 3 2]`, then `[3 1 2]`… a different permutation each run. The Go runtime randomizes the starting offset of map iteration *on purpose*, precisely so code can't quietly grow a dependency on it. If first-seen order (or any order) matters, ranging the map is a bug — and it's the kind that passes locally, passes one CI run, and fails the next.

## Why the order wanders (a peek at the map's internals)

Range the *same* map ten times and you get something like `[3 1 2]` nine times, then `[1 2 3]` once — a **skew, not a coin flip**. The skew is the tell, and it comes straight from the data structure.

A Go map is a hash table whose entries live in **buckets of 8 slots**. Three keys fit in one bucket, filled in insertion order:

```
bucket 0:  [ slot0 | slot1 | slot2 | slot3 … slot7 ]
             3       1       2       └── empty ──┘
```

The entries sit still. What `for k := range m` randomizes is only the **starting slot**: `mapiterinit` rolls an offset 0–7, then walks from there, wrapping mod 8 and skipping empties. So the printed order is just "where did the walk start?":

| start offset | first filled slot hit | order |
|---|---|---|
| 0 | slot0 | `[3 1 2]` |
| 1 | slot1 | `[1 2 3]` |
| 2 | slot2 | `[2 3 1]` |
| 3–7 | empty → wraps to slot0 | `[3 1 2]` |

Six of the eight offsets land on (or wrap to) slot0, so `[3 1 2]` appears ~**75%** of the time, `[1 2 3]` and `[2 3 1]` ~12.5% each. The empty tail is what biases iteration toward the "natural" slot order.

**Two traps hide in that skew:**

- **It's skewed, not uniform — which is *worse* for testing.** A uniform 1-in-N would fail fast and get noticed; a 75%-natural distribution looks stable across a few manual runs and lulls you into shipping it.
- **"Natural order == insertion order" is a coincidence, not a rule.** It holds here only because the map is tiny (one bucket), freshly built, with no deletions. Exceed 8 keys (multiple buckets — whose order is *also* randomized on iteration), or delete/re-add keys, or trigger a grow+rehash, and the slot layout diverges from insertion order entirely, driven by key hashes.

And it's **deliberate**: Go randomizes both a per-map hash seed and the per-range start offset precisely so nobody can depend on map order — early Go code did, and broke when the implementation changed. The randomization is a guardrail, not a quirk.

### See it yourself

```go
func main() {
    m := map[int]bool{3: true, 1: true, 2: true} // built once, never mutated
    for i := 0; i < 10; i++ {
        var out []int
        for k := range m { // range the SAME map ten times
            out = append(out, k)
        }
        fmt.Println(out)
    }
}
```

A real run of exactly this printed:

```
[3 1 2]
[3 1 2]
[3 1 2]
[3 1 2]
[3 1 2]
[3 1 2]
[3 1 2]
[3 1 2]
[3 1 2]
[1 2 3]
```

Nine times the same, once different — same map, same code, different order. That single deviant line is the whole point: if map order were stable it could never appear. And notice it prints the "natural" `[3 1 2]` ~90% of the time — which is exactly how a couple of manual runs fool you into thinking order is preserved. (The 9-to-1 skew is explained in the section above.)

## When each is right

- **Set membership / dedup** → map. `map[T]struct{}` or `map[T]bool`.
- **Order matters** (output slice, printed list, API response, test assertion) → keep an explicit slice; never rely on map range order. If all you have is the map, `slices.Sort` the keys or track insertion order separately.
- **Nested loop** → only when n is tiny and known-small, or you genuinely have no hashable key. Otherwise it's the quadratic trap.

## How to spot it in review

- `for k := range someMap { out = append(out, k) }` feeding anything order-sensitive — nondeterministic. Sort it, or preserve order upstream.
- A test that sorts a result "just to be safe" before asserting — often a sign the producer leaked map order.
- An inner loop doing a linear `contains` check — replace with a set.

## The habit

> A map answers "have I seen this?" in O(1). It does **not** remember order — iteration is randomized by design. Need order? Carry a slice alongside the map.

## References

- Go spec — [For statements (range)](https://go.dev/ref/spec#For_statements) — "The iteration order over maps is not specified and is not guaranteed to be the same from one iteration to the next."
- Go blog — [Go maps in action](https://go.dev/blog/maps) — the `struct{}` set idiom and comma-ok
- [`slices.Sort`](https://pkg.go.dev/slices#Sort) (Go 1.21+) — the usual "I need the map's keys in a stable order" fix
