package engine

import "fmt"

// EventType — what happened to a goroutine/channel.
type EventType int

const (
	Spawn   EventType = iota // a new goroutine was born
	Block                    // blocked (waiting on a channel/mutex)
	Unblock                  // unblocked
	Send                     // sent a value into a channel
	Done                     // the goroutine finished
	GCStart                  // a garbage-collection range began
	GCEnd                    // a garbage-collection range ended
)

func (e EventType) String() string {
	return [...]string{"spawn", "block", "unblock", "send", "done", "gcstart", "gcend"}[e]
}

// ParseEventType is the reverse mapping string → EventType, used by format loaders.
func ParseEventType(s string) (EventType, error) {
	switch s {
	case "spawn":
		return Spawn, nil
	case "block":
		return Block, nil
	case "unblock":
		return Unblock, nil
	case "send":
		return Send, nil
	case "done":
		return Done, nil
	case "gcstart":
		return GCStart, nil
	case "gcend":
		return GCEnd, nil
	default:
		return 0, fmt.Errorf("неизвестный тип события %q", s)
	}
}
