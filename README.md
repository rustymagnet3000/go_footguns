# go_footguns

Subtle Go behaviors that are easy to miss and hard to debug. Working Go, doing exactly what the spec says — and biting you anyway.

One folder per finding: write-up next to a runnable reproducer. `git clone`, `go run`, see the point.

## Findings

| # | Finding | The trap |
|---|---------|----------|
| — | *(added one at a time via PR)* | |

## Format

Each numbered folder contains:

- `README.md` — the write-up: what happened, why it bit, how to spot it in review, how to fix it
- `main.go` — minimal reproducer, no dependencies, runs standalone

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
