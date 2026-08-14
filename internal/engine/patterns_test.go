package engine

import "testing"

// Every pattern in the library must be valid YAML and load without errors.
func TestPatternLibraryLoads(t *testing.T) {
	names := []string{"workerpool", "fanin_fanout", "pipeline", "deadlock", "goroutine_leak"}
	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			scene, err := LoadScene(name)
			if err != nil {
				t.Fatalf("LoadScene(%q): %v", name, err)
			}
			if len(scene.Steps) == 0 {
				t.Fatalf("сцена %q загрузилась без шагов", name)
			}
			if scene.Description == "" {
				t.Errorf("сцена %q: пустое описание — нечего показать в UI", name)
			}
		})
	}
}

func countStates(f Frame) (running, blocked, finished int) {
	for _, st := range f.Goroutines {
		switch st {
		case Running:
			running++
		case Blocked:
			blocked++
		case Finished:
			finished++
		}
	}
	return running, blocked, finished
}

// fan-in/fan-out and pipeline are the "happy" patterns: by the end of the scene every goroutine has finished.
func TestHappyPatternsEndClean(t *testing.T) {
	for _, name := range []string{"fanin_fanout", "pipeline"} {
		t.Run(name, func(t *testing.T) {
			scene, err := LoadScene(name)
			if err != nil {
				t.Fatalf("LoadScene(%q): %v", name, err)
			}
			frames := scene.Frames()
			last := frames[len(frames)-1]
			running, blocked, _ := countStates(last)
			if running != 0 || blocked != 0 {
				t.Errorf("%s: в финальном кадре running=%d blocked=%d, ожидалось 0/0", name, running, blocked)
			}
		})
	}
}

// Deadlock: circular wait — both goroutines stay Blocked forever,
// neither ever reaches Finished.
func TestDeadlockStaysStuck(t *testing.T) {
	scene, err := LoadScene("deadlock")
	if err != nil {
		t.Fatalf("LoadScene: %v", err)
	}
	frames := scene.Frames()
	last := frames[len(frames)-1]

	_, blocked, finished := countStates(last)
	if blocked != 2 {
		t.Errorf("deadlock: в финальном кадре blocked=%d, ожидалось 2", blocked)
	}
	if finished != 0 {
		t.Errorf("deadlock: в финальном кадре finished=%d, ожидалось 0 — тупик не должен разрешаться", finished)
	}
}

// Goroutine leak: the orchestrator finishes, while the "leaked" goroutines pile
// up in Blocked and never free up — their count only grows from frame to frame.
func TestGoroutineLeakGrowsAndNeverShrinks(t *testing.T) {
	scene, err := LoadScene("goroutine_leak")
	if err != nil {
		t.Fatalf("LoadScene: %v", err)
	}
	frames := scene.Frames()

	prevBlocked := 0
	for _, f := range frames {
		_, blocked, _ := countStates(f)
		if blocked < prevBlocked {
			t.Fatalf("кадр %d: число зависших горутин уменьшилось с %d до %d — утечка не должна разрешаться", f.Index, prevBlocked, blocked)
		}
		prevBlocked = blocked
	}

	last := frames[len(frames)-1]
	running, blocked, finished := countStates(last)
	if running != 0 {
		t.Errorf("leak: в финальном кадре running=%d, ожидалось 0 (оркестратор уже завершился)", running)
	}
	if blocked != 5 {
		t.Errorf("leak: в финальном кадре зависших горутин=%d, ожидалось 5", blocked)
	}
	if finished != 1 {
		t.Errorf("leak: в финальном кадре finished=%d, ожидалось 1 (только оркестратор)", finished)
	}
}
