package engine

// Step — one event on the scene's timeline (what happened, to whom).
type Step struct {
	Label     string // label, e.g. "worker-1"
	Event     EventType
	Goroutine int // the goroutine id — the event's "actor"
	Chan      int // channel id (for Send/Block on a channel); 0 if not applicable
}

// Scene — a scenario: name, a short UI description, and ordered steps.
type Scene struct {
	Name        string
	Description string
	Steps       []Step
}
