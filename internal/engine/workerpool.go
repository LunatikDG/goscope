package engine

import "fmt"

// WorkerPool: 1 диспетчер раздаёт задачи N воркерам через канал.
func WorkerPool(workers int) Scene {
	const dispatcher = 0
	const jobs = 1 // id канала задач

	steps := []Step{{Event: Spawn, Goroutine: dispatcher, Label: "dispatcher"}}

	// воркеры рождаются и сразу блокируются в ожидании задачи
	for w := 1; w <= workers; w++ {
		steps = append(steps,
			Step{Event: Spawn, Goroutine: w, Label: fmt.Sprintf("worker-%d", w)},
			Step{Event: Block, Goroutine: w, Chan: jobs},
		)
	}
	// диспетчер шлёт задачи, разблокируя воркеров по очереди
	for w := 1; w <= workers; w++ {
		steps = append(steps,
			Step{Event: Send, Goroutine: dispatcher, Chan: jobs},
			Step{Event: Unblock, Goroutine: w, Chan: jobs},
			Step{Event: Done, Goroutine: w},
		)
	}
	steps = append(steps, Step{Event: Done, Goroutine: dispatcher})

	return Scene{Name: fmt.Sprintf("Worker Pool (%d)", workers), Steps: steps}
}