package traceparse

import (
	"sync"
	"testing"

	"github.com/LunatikDG/goscope/internal/engine"
	"github.com/LunatikDG/goscope/internal/render"
)

// TestParsedStepsRenderThroughExistingPipeline is Phase C's proof: a parsed
// trace's Steps are just engine.Steps. They fold into Frames via the exact
// Scene/Frames machinery the canned patterns and live NDJSON examples
// already use, and render.RenderFrame draws them with no changes needed on
// that side — real block/unblock states show up in the same colors.
func TestParsedStepsRenderThroughExistingPipeline(t *testing.T) {
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
		t.Fatal("Parse returned no steps to render")
	}

	scene := engine.Scene{Name: "real trace", Steps: steps}
	frames := scene.Frames()
	if len(frames) != len(steps)+1 {
		t.Fatalf("frames = %d, want %d (len(steps)+1)", len(frames), len(steps)+1)
	}

	layout := render.NewLayout(scene, 900, 360)

	sawBlockedLine := false
	for _, f := range frames {
		for _, op := range render.RenderFrame(f, layout) {
			if op.Kind == render.OpLine && op.Color == render.ColorBlocked {
				sawBlockedLine = true
			}
		}
	}
	if !sawBlockedLine {
		t.Error("expected at least one Blocked-colored line while rendering a real trace's frames")
	}

	// A real trace also captures long-lived Go runtime goroutines (GC
	// workers, sysmon, ...) that are still Running/Waiting when the trace
	// ends — unlike a scripted YAML scene, not everything reaches Finished.
	// What we *can* assert: every goroutine the trace itself reported a Done
	// for is rendered Finished in the final frame.
	doneGoroutines := map[int]bool{}
	for _, s := range steps {
		if s.Event == engine.Done {
			doneGoroutines[s.Goroutine] = true
		}
	}
	if len(doneGoroutines) == 0 {
		t.Fatal("no goroutine reached Done in this trace")
	}

	last := frames[len(frames)-1]
	for id := range doneGoroutines {
		if st := last.Goroutines[id]; st != engine.Finished {
			t.Errorf("goroutine %d had a Done step but its final rendered state is %v, want Finished", id, st)
		}
	}

	finishedLines := 0
	for _, op := range render.RenderFrame(last, layout) {
		if op.Kind == render.OpLine && op.Color == render.ColorFinished {
			finishedLines++
		}
	}
	if finishedLines == 0 {
		t.Error("expected at least one ColorFinished line in the final frame")
	}
}
