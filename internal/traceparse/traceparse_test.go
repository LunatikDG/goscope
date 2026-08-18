package traceparse

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/LunatikDG/goscope/internal/engine"
	"github.com/LunatikDG/goscope/internal/tracer"
)

// traceWorkload generates a real runtime/trace (via internal/tracer, the same
// package cmd/traced uses) over a small worker-pool workload and returns the
// path to the resulting file.
func traceWorkload(t *testing.T, workload func()) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "trace.out")

	stop, err := tracer.StartToFile(path)
	if err != nil {
		t.Fatalf("StartToFile: %v", err)
	}
	workload()
	if err := stop(); err != nil {
		t.Fatalf("stop: %v", err)
	}
	return path
}

func parseFile(t *testing.T, path string) []engine.Step {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open trace: %v", err)
	}
	defer f.Close()

	steps, err := Parse(f)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	return steps
}

func countByEvent(steps []engine.Step) map[engine.EventType]int {
	counts := map[engine.EventType]int{}
	for _, s := range steps {
		counts[s.Event]++
	}
	return counts
}

func TestParseWorkerPoolWorkload(t *testing.T) {
	path := traceWorkload(t, func() {
		const workers = 3
		jobs := make(chan int)
		var wg sync.WaitGroup

		for w := 0; w < workers; w++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				for j := range jobs {
					_ = j
				}
			}()
		}
		for j := 0; j < 10; j++ {
			jobs <- j
		}
		close(jobs)
		wg.Wait()
	})

	steps := parseFile(t, path)
	if len(steps) == 0 {
		t.Fatal("Parse returned no steps")
	}

	counts := countByEvent(steps)
	if counts[engine.Spawn] == 0 {
		t.Error("expected at least one Spawn step")
	}
	if counts[engine.Done] == 0 {
		t.Error("expected at least one Done step")
	}
	if counts[engine.Unblock] == 0 {
		t.Error("expected at least one Unblock step")
	}

	sawChanBlock := false
	for _, s := range steps {
		if s.Event == engine.Block && strings.Contains(s.Label, "chan") {
			sawChanBlock = true
			break
		}
	}
	if !sawChanBlock {
		t.Error("expected at least one Block step labeled with a channel reason (e.g. \"chan send\"/\"chan receive\")")
	}
}

func TestParseGCEvents(t *testing.T) {
	path := traceWorkload(t, func() {
		for i := 0; i < 3; i++ {
			buf := make([]byte, 10<<20) // 10MiB, to give GC something to do
			_ = buf
			runtime.GC()
		}
	})

	steps := parseFile(t, path)
	counts := countByEvent(steps)

	if counts[engine.GCStart] == 0 {
		t.Error("expected at least one GCStart step")
	}
	if counts[engine.GCEnd] == 0 {
		t.Error("expected at least one GCEnd step")
	}

	for _, s := range steps {
		if s.Event == engine.GCStart || s.Event == engine.GCEnd {
			if !strings.Contains(s.Label, "GC") {
				t.Errorf("GC step label %q doesn't mention GC", s.Label)
			}
		}
	}
}
