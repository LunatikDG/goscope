// Command workerpool is a real worker pool: the dispatcher hands out tasks over
// a shared channel, and which worker picks up which task is decided by the Go
// scheduler, not a script. Instrumented with internal/instrument for streaming to the browser.
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/LunatikDG/goscope/internal/instrument"
)

const jobsChan = 1

const workerCount = 3

func main() {
	instrument.Spawn(0, "dispatcher")

	jobs := make(chan int)
	var wg sync.WaitGroup
	for w := 1; w <= workerCount; w++ {
		wg.Add(1)
		go worker(w, jobs, &wg)
	}

	time.Sleep(80 * time.Millisecond) // give the workers time to block on the channel before the first send
	for j := 1; j <= workerCount; j++ {
		jobs <- j
	}
	close(jobs)

	wg.Wait()
	instrument.Done(0)
}

func worker(id int, jobs <-chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	instrument.Spawn(id, fmt.Sprintf("worker-%d", id))
	instrument.Block(id, jobsChan)

	if _, ok := <-jobs; ok {
		instrument.Unblock(id, jobsChan)
		time.Sleep(150 * time.Millisecond) // simulate work
	}
	instrument.Done(id)
}
