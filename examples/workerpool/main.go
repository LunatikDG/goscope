// Command workerpool — настоящий воркер-пул: диспетчер раздаёт задачи через
// общий канал, и какой воркер какую задачу подхватит — решает планировщик Go,
// а не сценарий. Инструментировано internal/instrument для стриминга в браузер.
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

	time.Sleep(80 * time.Millisecond) // дать воркерам заблокироваться на канале до первой отправки
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
		time.Sleep(150 * time.Millisecond) // имитация работы
	}
	instrument.Done(id)
}
