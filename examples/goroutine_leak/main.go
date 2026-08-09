// Command goroutineleak — оркестратор один за другим плодит горутины, которые
// блокируются на канале, в который никто и никогда не отправит, объявляет
// свою работу выполненной и завершается. Утёкшие горутины остаются висеть до
// самого конца процесса — ровно так леки и выглядят в реальных программах.
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

	never := make(chan struct{}) // никто никогда не отправит и не закроет

	for i := 1; i <= leakCount; i++ {
		id := i
		go func() {
			instrument.Spawn(id, fmt.Sprintf("leaked-%d", id))
			instrument.Block(id, leakChan)
			<-never
		}()
		time.Sleep(60 * time.Millisecond) // разнести спавн лекнутых горутин по таймлайну
	}

	instrument.Done(0)

	// подержать процесс живым, чтобы браузер успел увидеть зависшие горутины,
	// прежде чем выход из main убьёт их вместе с процессом
	time.Sleep(3 * time.Second)
}
