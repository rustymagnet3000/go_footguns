# 02 — parallel assignment binds positionally (and evaluates the whole right side first)

## The trap

Go lets you assign several variables in one statement:

```go
i, j := 0, len(runes)-1
```

Two things about this catch people out:

1. **Values bind positionally, left to right.** `i` gets `0`; `j` gets `len(runes)-1`. It is *not* "both start at `0`." In a dense `for` header the `0` sits right next to `j`, and the eye reads `j := 0`. It isn't — `j` is bound to the *second* value on the right.
2. **The entire right-hand side is evaluated before any assignment happens.** That snapshot is what makes `a, b = b, a` swap without a temp, and what makes `i, j = i+1, j-1` step both variables from their *old* values in a single statement.

Neither is a bug; both are easy to misread if your mental model is "assign left to right, one at a time."

## Where it showed up

In-place string reversal — the loop is parallel assignment three times over:

```go
func Reverse(s string) string {
    runes := []rune(s)
    for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
        runes[i], runes[j] = runes[j], runes[i]
    }
    return string(runes)
}
```

- **init:** `i, j := 0, len(runes)-1` → `i=0`, `j=`last index. (Not both `0`.)
- **post:** `i, j = i+1, j-1` → both step inward, each computed from the *old* value.
- **body:** `runes[i], runes[j] = runes[j], runes[i]` → swap, no temp variable.

For `"hello"` (length 5, last index **4**), the misread is thinking `j` starts at `0`. It starts at `4` — `len-1`, positionally bound to the second value.

## The rule that makes it work

The [Go spec](https://go.dev/ref/spec#Assignment_statements) defines a multi-assignment in **two phases**:

> 1. The operands of index expressions and pointer indirections on the **left**, and **all the expressions on the right**, are evaluated in the usual order.
> 2. The assignments are carried out in left-to-right order.

So the right-hand side is a photograph taken *before* any write lands. That's why all of these do what you want:

```go
a, b = b, a        // swap — no temp
x, y = y, x+y      // fibonacci step — x+y uses the OLD x
i, j = i+1, j-1    // both computed from old i, old j
```

## Where it actually bites

Phase 1 also evaluates **index expressions on the left** using the *old* values — so a left-hand subscript does not see an assignment made later in the same statement:

```go
i := 0
a := []int{10, 20, 30}
a[i], i = 99, 2        // which slot gets 99?
```

Phase 1 captures the left index `i` (still `0`) and the right side (`99`, `2`). Phase 2 writes `a[0] = 99`, then `i = 2`. So **`a[0]`** becomes `99` — *not* `a[2]` — even though `i` is set to `2` on the same line. If you assumed `i = 2` "happened first," you're wrong. None of this errors; it just silently targets the wrong element.

## How to spot it in review

- A multi-assign `for` header (`for i, j := ...; ...; i, j = ...`): read the commas as **positional pairs**, not "everything on the left equals the first thing on the right."
- Any `x, y = <expr using x or y>, <expr using x or y>`: confirm the author *intended* old-value semantics (they usually did — that's the point).
- `a[i], i = ...` shapes: rare, almost always a bug or an unreadable trick. Split it into two lines.

## The fix / the habit

Nothing to fix in `Reverse` — it's idiomatic and correct. The fix is the mental model:

> Commas pair up **positionally**. The right side is a photograph taken **before** any assignment. Then the left side is written, **in order**.

Once that clicks, `a, b = b, a` and `i, j = i+1, j-1` read as intended, and the `a[i], i = ...` trap stops surprising you.

## References

- Go spec — [Assignment statements](https://go.dev/ref/spec#Assignment_statements) — the two-phase evaluation rule
- Go spec — [`for` statements](https://go.dev/ref/spec#For_statements) — init and post are ordinary simple statements, so they can be multi-assignments
- [`slices.Reverse`](https://pkg.go.dev/slices#Reverse) (Go 1.21+) — the stdlib version of the loop above, when you don't need to hand-roll it
