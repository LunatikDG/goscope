// Package instrument gives the instrumented examples in examples/ a single way
// to report what's happening: one NDJSON line on stdout per event. The server
// (cmd/serve) runs an example as a subprocess, reads stdout line by line, and
// streams those same lines to the browser over SSE.
package instrument

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Event — one line on stdout; its fields mirror engine.Step in meaning.
type Event struct {
	Event     string `json:"event"`
	Label     string `json:"label,omitempty"`
	Goroutine int    `json:"goroutine"`
	Chan      int    `json:"chan,omitempty"`
}

// mu protects stdout: several of the example's goroutines write events at the
// same time; without a mutex the lines could interleave and break line-oriented NDJSON.
var mu sync.Mutex

func emit(e Event) {
	data, err := json.Marshal(e)
	if err != nil {
		return // Event always marshals cleanly; we never actually get here
	}

	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintln(os.Stdout, string(data))
}

// Spawn reports that goroutine was born, with label (may be empty).
func Spawn(goroutine int, label string) {
	emit(Event{Event: "spawn", Goroutine: goroutine, Label: label})
}

// Block reports that goroutine blocked on channel ch.
func Block(goroutine, ch int) {
	emit(Event{Event: "block", Goroutine: goroutine, Chan: ch})
}

// Unblock reports that goroutine unblocked (received a value from ch).
func Unblock(goroutine, ch int) {
	emit(Event{Event: "unblock", Goroutine: goroutine, Chan: ch})
}

// Send reports sending a value into channel ch.
func Send(goroutine, ch int) {
	emit(Event{Event: "send", Goroutine: goroutine, Chan: ch})
}

// Done reports that the goroutine finished.
func Done(goroutine int) {
	emit(Event{Event: "done", Goroutine: goroutine})
}
