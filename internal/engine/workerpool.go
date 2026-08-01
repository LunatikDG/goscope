package engine

import "fmt"

// WorkerPool моделирует пул: диспетчер раздаёт N задач воркерам по одной.
// Последовательность специально «разрежена», чтобы фазы читались глазом:
//  1. поднимается диспетчер и воркеры (воркеры сразу блокируются на канале задач)
//  2. диспетчер по одной шлёт задачи → воркер получает (unblock) → работает → done
func WorkerPool(workers int) Scene {
	const dispatcher = 0
	const jobs = 1

	var steps []Step
	add := func(e EventType, g, ch int, label string) {
		steps = append(steps, Step{Event: e, Goroutine: g, Chan: ch, Label: label})
	}

	// фаза 1 — поднимаем пул
	add(Spawn, dispatcher, 0, "dispatcher")
	for w := 1; w <= workers; w++ {
		add(Spawn, w, 0, fmt.Sprintf("worker-%d", w))
		add(Block, w, jobs, "") // ждёт задачу на канале jobs
	}

	// фаза 2 — задачи прокатываются по воркерам по одной
	for w := 1; w <= workers; w++ {
		add(Send, dispatcher, jobs, "") // диспетчер отправил задачу
		add(Unblock, w, jobs, "")       // воркер получил и проснулся (Running)
		add(Done, w, 0, "")             // воркер отработал и завершился
	}

	add(Done, dispatcher, 0, "") // диспетчер закончил раздачу
	return Scene{Name: fmt.Sprintf("Worker Pool (%d)", workers), Steps: steps}
}
