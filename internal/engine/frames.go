package engine

// Frames разворачивает сцену в последовательность кадров.
func (s Scene) Frames() []Frame {
	state := map[int]GoroutineState{}
	frames := make([]Frame, 0, len(s.Steps)+1)
	frames = append(frames, snapshot(0, state, nil)) // начальный кадр без причины

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
		// v1: сам send состояние актёра не меняет;
		// позже здесь разблокируем получателя
	}
}

func snapshot(index int, state map[int]GoroutineState, cause *Step) Frame {
	cp := make(map[int]GoroutineState, len(state))
	for k, v := range state {
		cp[k] = v
	}
	return Frame{Index: index, Goroutines: cp, Cause: cause}
}
