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

It works, but it's quadratic. The idiomatic Go answer — and what an interviewer asking to "see map usage" wants — is a **map as a set**: O(1) membership, O(n) overall.

**Trap 2 — building the result by ranging the map.** Once you have a map, it's tempting to skip the result slice and just range the keys. But **Go deliberately randomizes map iteration order**, so your unique values come out in a *different order on every run*.

## The correct shape

Map for **membership**, slice for **order**. Append to the slice the first time you see each key:

```go
func unique(nums []int) []int {
    seen := make(map[int]struct{})
    out := make([]int, 0, len(nums))
    for _, n := range nums {
        if _, ok := seen[n]; !ok { // comma-ok membership test
            seen[n] = struct{}{}   // add to the set
            out = append(out, n)   // record first-seen order
        }
    }
    return out
}
```

- `map[int]struct{}` is the zero-byte **set** idiom: the value carries no information, so `struct{}{}` costs no memory per entry. (`map[int]bool` also works and reads a touch friendlier as `if seen[n]`; it just stores a pointless bool.)
- `_, ok := seen[n]` is the **comma-ok** test: `ok` is `true` iff the key is present.

## What breaks if you range the map for output

```go
// BUG: order is nondeterministic
func uniqueBad(nums []int) []int {
    seen := make(map[int]struct{})
    for _, n := range nums {
        seen[n] = struct{}{}
    }

    out := make([]int, 0, len(seen))
    for n := range seen { // <-- randomized every run
        out = append(out, n)
    }
    return out
}
```

`uniqueBad([]int{3, 1, 3, 2, 1})` returns the set `{1, 2, 3}` — but as `[2 3 1]`, then `[1 3 2]`, then `[3 1 2]`… a different permutation each run. The Go runtime randomizes the starting offset of map iteration *on purpose*, precisely so code can't quietly grow a dependency on it. If first-seen order (or any order) matters, ranging the map is a bug — and it's the kind that passes locally, passes one CI run, and fails the next.

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
