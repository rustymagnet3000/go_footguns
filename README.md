# go_footguns

Subtle Go behaviors that are easy to miss and hard to debug. Working Go, doing exactly what the spec says — and biting you anyway.

One folder per finding: write-up, embedded code, and — where it helps — a runnable reproducer.

## Findings

| # | Finding | The trap |
|---|---------|----------|
| 01 | [`strings.Fields` vs `strings.Split`](./01-strings-fields-vs-split) | `strings.Split(s, " ")` returns empty tokens on adjacent spaces and doesn't split on tabs/newlines. `strings.Fields` splits on any Unicode whitespace and skips empties. |

## Format

Each numbered folder contains a `README.md` — the write-up: what happened, why it bit, how to spot it in review, how to fix it. Code samples are embedded inline; some findings also include a runnable `main.go` reproducer alongside.

## Why

A public log of the subtle things Go lets you do that will one day cost you a debugging evening. Written for readers who've shipped Go, not for people learning it.

Not a tutorial. Not a beginners' guide. Not "Effective Go". Assumes you've read that already and want your remaining assumptions punctured.

## Canon (further reading worth your time)

- Dave Cheney — [Go Language Design](https://dave.cheney.net) — decades of design commentary
- Teiva Harsanyi — [*100 Go Mistakes and How to Avoid Them*](https://100go.co/) — the definitive index
- The Go release notes — [tip.golang.org/doc/](https://tip.golang.org/doc/) — each minor release quietly changes something surprising
- `go doc -all` on the standard library — the footguns are documented, we just don't read them

## Not covered here

- Basics (`goroutine`, `chan`, error handling patterns)
- Style / lint issues
- Anything solved by reading the standard library docs once
