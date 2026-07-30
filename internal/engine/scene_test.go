package engine

import "testing"

// (б) Финальное состояние worker pool: все горутины завершены.
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

// (а) Finished терминально: если горутина завершилась в кадре i,
// она остаётся Finished во всех последующих кадрах.
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

// Переходы состояний: каждое событие приводит горутину в ожидаемое состояние.
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

// Число кадров = число шагов + 1 (начальный пустой кадр).
func TestFramesCount(t *testing.T) {
	scene := WorkerPool(3)
	got := len(scene.Frames())
	want := len(scene.Steps) + 1
	if got != want {
		t.Errorf("кадров = %d, ожидалось %d", got, want)
	}
}

// Инвариант: в любом кадре все состояния «известные»,
// а число живых (не Finished) горутин не отрицательно.
func TestFramesInvariant(t *testing.T) {
	for _, f := range WorkerPool(3).Frames() {
		alive := 0
		for id, st := range f.Goroutines {
			switch st {
			case Running, Blocked:
				alive++
			case Finished:
				// ок
			default:
				t.Fatalf("кадр %d: горутина %d в неизвестном состоянии %v", f.Index, id, st)
			}
		}
		if alive < 0 {
			t.Fatalf("кадр %d: отрицательное число живых горутин", f.Index)
		}
	}
}

// Главный тест на граблю: кадры НЕ делят одну map.
// Изменение состояния после снятия кадра не должно менять уже снятый кадр.
func TestFramesAreIndependentSnapshots(t *testing.T) {
	scene := Scene{Steps: []Step{
		{Event: Spawn, Goroutine: 1}, // кадр 1: g1 = Running
		{Event: Done, Goroutine: 1},  // кадр 2: g1 = Finished
	}}
	frames := scene.Frames()

	// frames[1] снят ПОСЛЕ первого шага (Spawn) → g1 должна быть Running
	if got := frames[1].Goroutines[1]; got != Running {
		t.Fatalf("кадр 1: g1 = %v, ожидалось Running (кадры делят одну map?)", got)
	}
	// frames[2] снят после Done → g1 = Finished, но кадр 1 остался прежним
	if got := frames[2].Goroutines[1]; got != Finished {
		t.Fatalf("кадр 2: g1 = %v, ожидалось Finished", got)
	}
	if frames[1].Goroutines[1] == frames[2].Goroutines[1] {
		t.Fatal("кадры 1 и 2 ссылаются на одно состояние — нарушена независимость снимков")
	}
}
