package engine

import "fmt"

// WorkerPool models a pool: a dispatcher hands out N tasks to workers one at a time.
// The sequence is deliberately "spread out" so the phases read clearly:
//  1. the dispatcher and workers spin up (workers immediately block on the jobs channel)
//  2. the dispatcher sends tasks one by one → a worker receives (unblock) → works → done
func WorkerPool(workers int) Scene {
	const dispatcher = 0
	const jobs = 1

	var steps []Step
	add := func(e EventType, g, ch int, label string) {
		steps = append(steps, Step{Event: e, Goroutine: g, Chan: ch, Label: label})
	}

	// phase 1 — spin up the pool
	add(Spawn, dispatcher, 0, "dispatcher")
	for w := 1; w <= workers; w++ {
		add(Spawn, w, 0, fmt.Sprintf("worker-%d", w))
		add(Block, w, jobs, "") // waiting for a task on the jobs channel
	}

	// phase 2 — tasks roll out to workers one at a time
	for w := 1; w <= workers; w++ {
		add(Send, dispatcher, jobs, "") // dispatcher sent a task
		add(Unblock, w, jobs, "")       // worker received it and woke up (Running)
		add(Done, w, 0, "")             // worker finished and is done
	}

	add(Done, dispatcher, 0, "") // dispatcher finished handing out work
	return Scene{Name: fmt.Sprintf("Worker Pool (%d)", workers), Steps: steps}
}
