package engine

// Frames unrolls the scene into a sequence of frames.
func (s Scene) Frames() []Frame {
	state := map[int]GoroutineState{}
	frames := make([]Frame, 0, len(s.Steps)+1)
	frames = append(frames, snapshot(0, state, nil)) // initial frame, no cause

	for i := range s.Steps {
		apply(state, s.Steps[i])
		frames = append(frames, snapshot(i+1, state, &s.Steps[i]))
	}
	return frames
}

func apply(state map[int]GoroutineState, step Step) {
	switch step.Event {
	case Spawn, Unblock:
		state[step.Goroutine] = Running
	case Block:
		state[step.Goroutine] = Blocked
	case Done:
		state[step.Goroutine] = Finished
	case Send:
		// v1: send itself doesn't change the sender's state;
		// later this is where we'd unblock the receiver
	case GCStart, GCEnd:
		// GC ranges aren't scoped to one goroutine's rendered state (yet);
		// they just ride along on the frame's Cause for a future GC visualization
	}
}

func snapshot(index int, state map[int]GoroutineState, cause *Step) Frame {
	cp := make(map[int]GoroutineState, len(state))
	for k, v := range state {
		cp[k] = v
	}
	return Frame{Index: index, Goroutines: cp, Cause: cause}
}

// LiveFolder unrolls steps into frames one at a time, for scenes whose full
// sequence isn't known upfront (e.g. streamed from the server).
type LiveFolder struct {
	state map[int]GoroutineState
	index int
}

// NewLiveFolder creates an empty folder, ready to accept steps as they arrive.
func NewLiveFolder() *LiveFolder {
	return &LiveFolder{state: map[int]GoroutineState{}}
}

// Apply applies the next step and returns a snapshot of the world after it.
func (f *LiveFolder) Apply(step Step) Frame {
	apply(f.state, step)
	f.index++
	return snapshot(f.index, f.state, &step)
}
