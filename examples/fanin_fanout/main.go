// Command faninfanout — fan-out: the dispatcher hands out tasks to workers over
// a shared channel; fan-in: the workers send results to a single collector.
// Real concurrency: the order of events on the timeline is decided by the Go scheduler.
package main

import (
	"fmt"
	"sync"
	"time"

	"github.com/LunatikDG/goscope/internal/instrument"
)

const (
	jobsChan    = 1
	resultsChan = 2
	workerCount = 3
)

func main() {
	instrument.Spawn(0, "dispatcher")
	instrument.Spawn(1, "collector")
	instrument.Block(1, resultsChan)

	jobs := make(chan int)
	results := make(chan int)

	var workers sync.WaitGroup
	for w := 0; w < workerCount; w++ {
		id := 2 + w
		workers.Add(1)
		go worker(id, jobs, results, &workers)
	}

	go func() {
		workers.Wait()
		close(results)
	}()

	time.Sleep(80 * time.Millisecond) // give the workers time to block on jobs before the first send
	for j := 1; j <= workerCount; j++ {
		jobs <- j
	}
	close(jobs)
	instrument.Done(0)

	collected := 0
	for range results {
		collected++
		instrument.Unblock(1, resultsChan)
		if collected < workerCount {
			instrument.Block(1, resultsChan) // waiting for the next result
		}
	}
	instrument.Done(1)
}

func worker(id int, jobs <-chan int, results chan<- int, wg *sync.WaitGroup) {
	defer wg.Done()
	instrument.Spawn(id, fmt.Sprintf("worker-%d", id-1))
	instrument.Block(id, jobsChan)

	job, ok := <-jobs
	if !ok {
		instrument.Done(id)
		return
	}
	instrument.Unblock(id, jobsChan)
	time.Sleep(120 * time.Millisecond) // simulate work

	instrument.Send(id, resultsChan)
	results <- job
	instrument.Done(id)
}
