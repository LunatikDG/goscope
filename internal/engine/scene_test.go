package engine

import (
	"reflect"
	"testing"
)

// (b) Final state of the worker pool: all goroutines finished.
func TestWorkerPoolEndsClean(t *testing.T) {
	frames := WorkerPool(3).Frames()
	last := frames[len(frames)-1]

	alive := 0
	for _, st := range last.Goroutines {
		if st == Running || st == Blocked {
			alive++
		}
	}
	if alive != 0 {
		t.Errorf("в финальном кадре осталось живых горутин: %d, ожидалось 0", alive)
	}
}

// (a) Finished is terminal: if a goroutine finished in frame i,
// it stays Finished in every frame after that.
func TestFinishedIsTerminal(t *testing.T) {
	frames := WorkerPool(3).Frames()
	finished := map[int]bool{}

	for _, f := range frames {
		for id, st := range f.Goroutines {
			if finished[id] && st != Finished {
				t.Errorf("кадр %d: горутина %d ожила из Finished в %v", f.Index, id, st)
			}
			if st == Finished {
				finished[id] = true
			}
		}
	}
}

// State transitions: each event drives a goroutine to the expected state.
func TestApplyTransitions(t *testing.T) {
	tests := []struct {
		name  string
		event EventType
		want  GoroutineState
	}{
		{"spawn → running", Spawn, Running},
		{"unblock → running", Unblock, Running},
		{"block → blocked", Block, Blocked},
		{"done → finished", Done, Finished},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			state := map[int]GoroutineState{}
			apply(state, Step{Event: tt.event, Goroutine: 7})
			if got := state[7]; got != tt.want {
				t.Errorf("после %s: состояние = %v, ожидалось %v", tt.event, got, tt.want)
			}
		})
	}
}

// Frame count = step count + 1 (the initial empty frame).
func TestFramesCount(t *testing.T) {
	scene := WorkerPool(3)
	got := len(scene.Frames())
	want := len(scene.Steps) + 1
	if got != want {
		t.Errorf("кадров = %d, ожидалось %d", got, want)
	}
}

// Invariant: every frame only has "known" states, and the number
// of live (non-Finished) goroutines is never negative.
func TestFramesInvariant(t *testing.T) {
	for _, f := range WorkerPool(3).Frames() {
		alive := 0
		for id, st := range f.Goroutines {
			switch st {
			case Running, Blocked:
				alive++
			case Finished:
				// ok
			default:
				t.Fatalf("кадр %d: горутина %d в неизвестном состоянии %v", f.Index, id, st)
			}
		}
		if alive < 0 {
			t.Fatalf("кадр %d: отрицательное число живых горутин", f.Index)
		}
	}
}

// The main regression test: frames must NOT share one map.
// Changing state after a frame was taken must not change that already-taken frame.
func TestFramesAreIndependentSnapshots(t *testing.T) {
	scene := Scene{Steps: []Step{
		{Event: Spawn, Goroutine: 1}, // frame 1: g1 = Running
		{Event: Done, Goroutine: 1},  // frame 2: g1 = Finished
	}}
	frames := scene.Frames()

	// frames[1] is taken AFTER the first step (Spawn) → g1 should be Running
	if got := frames[1].Goroutines[1]; got != Running {
		t.Fatalf("кадр 1: g1 = %v, ожидалось Running (кадры делят одну map?)", got)
	}
	// frames[2] is taken after Done → g1 = Finished, but frame 1 stays unchanged
	if got := frames[2].Goroutines[1]; got != Finished {
		t.Fatalf("кадр 2: g1 = %v, ожидалось Finished", got)
	}
	if frames[1].Goroutines[1] == frames[2].Goroutines[1] {
		t.Fatal("кадры 1 и 2 ссылаются на одно состояние — нарушена независимость снимков")
	}
}

// LiveFolder, applying steps one at a time, must produce the same frames as
// Scene.Frames() does for the same set of steps all at once.
func TestLiveFolderMatchesFrames(t *testing.T) {
	steps := WorkerPool(3).Steps
	want := (Scene{Steps: steps}).Frames()

	folder := NewLiveFolder()
	got := make([]Frame, 0, len(steps)+1)
	got = append(got, Frame{Index: 0, Goroutines: map[int]GoroutineState{}})
	for _, step := range steps {
		got = append(got, folder.Apply(step))
	}

	if len(got) != len(want) {
		t.Fatalf("кадров = %d, ожидалось %d", len(got), len(want))
	}
	for i := range want {
		if !reflect.DeepEqual(got[i].Goroutines, want[i].Goroutines) {
			t.Fatalf("кадр %d: live = %+v, batch = %+v", i, got[i].Goroutines, want[i].Goroutines)
		}
	}
}
