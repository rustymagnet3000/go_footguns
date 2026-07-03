# 01 — you probably want `strings.Fields`, not `strings.Split`

## The trap

`strings.Split(s, " ")` and `strings.Fields(s)` look interchangeable when you're splitting a sentence into words. They aren't:

- `strings.Split` splits on the **exact** separator you pass. Adjacent separators produce empty-string tokens. Non-space whitespace (`\t`, `\n`, `\r`) doesn't split anything.
- `strings.Fields` splits on any run of Unicode whitespace (per [`unicode.IsSpace`](https://pkg.go.dev/unicode#IsSpace)) and never returns empty tokens.

For real-world input — copy-pasted text, log lines, user-typed strings — `Split(" ")` will silently miscount.

## The correct shape

`WordCount` returns a map from each whitespace-separated word to how many times it appears:

```
WordCount("the cat sat on the mat the")
  == map[string]int{"the": 3, "cat": 1, "sat": 1, "on": 1, "mat": 1}
```

```go
func WordCount(str string) map[string]int {
    wordMap := make(map[string]int)

    // non-unique word slice
    words := strings.Fields(str)
    for _, word := range words {
        wordMap[word]++
    }

    return wordMap
}

func main() {
    output := WordCount("the cat sat on the mat the")
    fmt.Println(output)
}
```

## What breaks if you reach for `strings.Split`

Given `" the cat sat  on the\tmat the"` (leading space, double space, tab):

| Function | Tokens produced |
|---|---|
| `strings.Split(s, " ")` | `["", "the", "cat", "sat", "", "on", "the\tmat", "the"]` |
| `strings.Fields(s)` | `["the", "cat", "sat", "on", "the", "mat", "the"]` |

`Split` gives you: two empty-string keys in your map, "the\tmat" as one un-tokenised word, and an undercount for "mat". None of these throw — the bug ships silently.

## When each is right

- **`strings.Fields`** — natural-language text, log lines, whitespace-separated user input, anything where "whitespace" is a fuzzy concept.
- **`strings.Split`** — strict delimiters where empty tokens are meaningful and the separator character is precise. CSV parsing (usually via `encoding/csv`, but if you must), path components, protocol frames, `Cookie:` header parsing.

Rule of thumb: if you're about to write `strings.Split(s, " ")`, you almost certainly want `Fields` instead.

## References

- [`strings.Fields`](https://pkg.go.dev/strings#Fields)
- [`strings.Split`](https://pkg.go.dev/strings#Split)
- [`unicode.IsSpace`](https://pkg.go.dev/unicode#IsSpace) — the definition of whitespace `Fields` uses
