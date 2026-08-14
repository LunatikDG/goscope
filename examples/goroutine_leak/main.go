// Command goroutineleak has an orchestrator spawn goroutines one after another;
// each blocks on a channel nobody will ever send to. It then declares its own
// work done and exits. The leaked goroutines stay hanging until the process
// itself ends — exactly how leaks look in real programs.
package main

import (
	"fmt"
	"time"

	"github.com/LunatikDG/goscope/internal/instrument"
)

const leakChan = 1

const leakCount = 5

func main() {
	instrument.Spawn(0, "orchestrator")

	never := make(chan struct{}) // nobody will ever send to or close this

	for i := 1; i <= leakCount; i++ {
		id := i
		go func() {
			instrument.Spawn(id, fmt.Sprintf("leaked-%d", id))
			instrument.Block(id, leakChan)
			<-never
		}()
		time.Sleep(60 * time.Millisecond) // spread the leaked goroutines' spawns out over the timeline
	}

	instrument.Done(0)

	// keep the process alive so the browser has time to see the hung goroutines
	// before returning from main kills them along with the process
	time.Sleep(3 * time.Second)
}
