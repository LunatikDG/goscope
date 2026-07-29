package engine

// GoroutineState — состояние горутины в конкретный момент.
type GoroutineState int

const (
	Running GoroutineState = iota
	Blocked
	Finished
)

// Frame - снимок мира в один момент (это и рисует рендер).
type Frame struct {
	Goroutines map[int]GoroutineState
	Index      int
}
