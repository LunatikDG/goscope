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
)

func (e EventType) String() string {
	return [...]string{"spawn", "block", "unblock", "send", "done"}[e]
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
	default:
		return 0, fmt.Errorf("неизвестный тип события %q", s)
	}
}
