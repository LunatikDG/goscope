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

// SampledFrames behaves like Frames, but for scenes with more than maxFrames
// steps it only snapshots the state at evenly spaced points instead of after
// every single step. Every step is still applied — the state stays correct —
// only the (expensive) full-map copy in snapshot() is skipped for the steps
// in between. This bounds both the number of frames a caller has to hold
// onto and the number of map copies to O(maxFrames) rather than O(len(Steps)),
// which matters once a scene has thousands of steps (a real trace, or an
// extreme goroutine leak) and building every frame would otherwise be what
// hangs the browser before a single one is drawn.
//
// The final step is always snapshotted, so the last frame always reflects
// the scene's true end state. maxFrames <= 0 falls back to Frames.
func (s Scene) SampledFrames(maxFrames int) []Frame {
	if maxFrames <= 0 || len(s.Steps) <= maxFrames {
		return s.Frames()
	}

	state := map[int]GoroutineState{}
	frames := make([]Frame, 0, maxFrames+1)
	frames = append(frames, snapshot(0, state, nil))

	last := len(s.Steps) - 1
	// The last step is always kept (below), so only budget the rest of
	// maxFrames for evenly-spaced picks — otherwise an unaligned last step
	// would tip the total one over the cap.
	budget := maxFrames - 1
	if budget < 1 {
		budget = 1
	}
	stride := ceilDiv(last+1, budget)
	if stride == 0 {
		stride = 1
	}
	for i := range s.Steps {
		apply(state, s.Steps[i])
		if i == last || i%stride == 0 {
			frames = append(frames, snapshot(i+1, state, &s.Steps[i]))
		}
	}
	return frames
}

func ceilDiv(a, b int) int { return (a + b - 1) / b }

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
