// Command deadlock is a genuine circular wait: goroutine-a waits for something
// only goroutine-b could send, and vice versa. Neither Send ever happens, and
// the Go runtime detects the deadlock itself — the process exits with
// "fatal error: all goroutines are asleep - deadlock!". That's expected: the
// server runs the example as a separate subprocess exactly for this kind of isolation.
package main

import "github.com/LunatikDG/goscope/internal/instrument"

func main() {
	waitForA := make(chan struct{}) // this is where goroutine-b waits for what goroutine-a could have sent
	waitForB := make(chan struct{}) // this is where goroutine-a waits for what goroutine-b could have sent

	go goroutineA(waitForB)
	go goroutineB(waitForA)

	select {} // we block forever here too — together with a and b, that's a full deadlock of the whole program
}

func goroutineA(waitForB <-chan struct{}) {
	instrument.Spawn(1, "goroutine-a")
	instrument.Block(1, 2)
	<-waitForB
}

func goroutineB(waitForA <-chan struct{}) {
	instrument.Spawn(2, "goroutine-b")
	instrument.Block(2, 1)
	<-waitForA
}
