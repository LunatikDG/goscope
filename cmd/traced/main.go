// Command traced runs a small concurrent workload while recording a real
// runtime/trace — the input for a future trace-parsing/visualization phase.
// Run it, then inspect the result with `go tool trace <path>`.
package main

import (
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/LunatikDG/goscope/internal/tracer"
)

func main() {
	path := flag.String("trace", "trace.out", "path to write the runtime/trace to")
	flag.Parse()

	stop, err := tracer.StartToFile(*path)
	if err != nil {
		log.Fatal(err)
	}

	runWorkload()

	if err := stop(); err != nil {
		log.Fatal(err)
	}

	fmt.Printf("wrote trace to %s — inspect it with: go tool trace %s\n", *path, *path)
}

// runWorkload is a small worker pool: enough goroutine creation, channel
// blocking, and scheduling to produce a trace with something to look at.
func runWorkload() {
	const workers = 3
	const jobCount = 10

	jobs := make(chan int)
	var wg sync.WaitGroup

	for w := 1; w <= workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				time.Sleep(20 * time.Millisecond) // simulate work
			}
		}()
	}

	for j := 1; j <= jobCount; j++ {
		jobs <- j
	}
	close(jobs)
	wg.Wait()
}
