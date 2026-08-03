// Command main demonstrates a Go-specific denial-of-service footgun: concurrent
// writes to a built-in map trigger `fatal error: concurrent map writes`, an
// UNRECOVERABLE runtime crash. The deferred recover() below never fires — which
// is the whole point: one racy map access anywhere takes down the whole process,
// and you cannot defend against it with recover().
//
//	go run ./05-concurrent-map-panic         # crashes with a fatal error
//	go run -race ./05-concurrent-map-panic   # the race detector pinpoints the line
package main

import (
	"fmt"
	"sync"
)

func main() {
	// A runtime throw is not a panic: recover() cannot catch it, so this
	// deferred handler never runs. It's here to prove exactly that.
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered:", r) // unreachable for a fatal error
		}
	}()

	m := make(map[int]int)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 0; j < 10000; j++ {
				m[id] = j // unsynchronised writes from many goroutines
			}
		}(i)
	}
	wg.Wait()
	fmt.Println("finished without crashing (len:", len(m), ") — got lucky; run it again")
}
