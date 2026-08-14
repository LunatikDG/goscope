package engine

// GoroutineState — a goroutine's state at a given moment.
type GoroutineState int

const (
	Running GoroutineState = iota
	Blocked
	Finished
)

// Frame - a snapshot of the world at one moment (this is what render draws).
type Frame struct {
	Goroutines map[int]GoroutineState
	Cause      *Step // the event that produced this frame; nil for the initial one
	Index      int
}
