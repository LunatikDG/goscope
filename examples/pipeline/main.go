// Command pipeline sends a value through a chain of stages: producer -> stage2
// -> stage3, each stage connected to its neighbor by its own channel.
package main

import (
	"time"

	"github.com/LunatikDG/goscope/internal/instrument"
)

const (
	stage1to2 = 1
	stage2to3 = 2
)

func main() {
	ch1 := make(chan int)
	ch2 := make(chan int)

	done2 := make(chan struct{})
	done3 := make(chan struct{})

	go stage2(ch1, ch2, done2)
	go stage3(ch2, done3)

	time.Sleep(80 * time.Millisecond) // give the stages time to block on their input channels

	instrument.Spawn(0, "producer")
	instrument.Send(0, stage1to2)
	ch1 <- 1
	close(ch1)
	instrument.Done(0)

	<-done2
	<-done3
}

func stage2(in <-chan int, out chan<- int, done chan<- struct{}) {
	instrument.Spawn(1, "stage-2")
	instrument.Block(1, stage1to2)

	v, ok := <-in
	if !ok {
		instrument.Done(1)
		close(done)
		return
	}
	instrument.Unblock(1, stage1to2)

	instrument.Send(1, stage2to3)
	out <- v
	close(out)
	instrument.Done(1)
	close(done)
}

func stage3(in <-chan int, done chan<- struct{}) {
	instrument.Spawn(2, "stage-3")
	instrument.Block(2, stage2to3)

	if _, ok := <-in; ok {
		instrument.Unblock(2, stage2to3)
	}
	instrument.Done(2)
	close(done)
}
