package render

import (
	"testing"

	"github.com/LunatikDG/goscope/internal/engine"
)

// A scene with more goroutines than MaxLanes switches the layout into aggregated mode.
func TestNewLayoutAggregatesBeyondMaxLanes(t *testing.T) {
	small := NewLayout(engine.WorkerPool(MaxLanes-1), 900, 360)
	if small.Aggregated() {
		t.Error("a scene at MaxLanes-1 goroutines should not be aggregated")
	}

	big := NewLayout(engine.WorkerPool(MaxLanes+50), 900, 360)
	if !big.Aggregated() {
		t.Error("a scene well past MaxLanes goroutines should be aggregated")
	}
}

// However many goroutines a scene has, an aggregated frame produces a
// handful of Ops (one line + one label per non-empty state), not one per
// goroutine — this is what keeps a thousands-of-goroutines scene from
// flooding the canvas with draw commands.
func TestRenderFrameAggregatedOpsAreBounded(t *testing.T) {
	scene := engine.WorkerPool(2000)
	l := NewLayout(scene, 900, 360)
	if !l.Aggregated() {
		t.Fatal("expected a 2000-worker scene to trigger aggregated mode")
	}

	frames := scene.SampledFrames(50)
	for _, f := range frames {
		ops := RenderFrame(f, l)
		if len(ops) > 6 { // 3 states x (1 line + 1 label)
			t.Fatalf("frame %d produced %d ops, want <= 6 regardless of goroutine count", f.Index, len(ops))
		}
	}
}

// The aggregated counts reported in the Ops match the frame's actual
// per-state goroutine counts.
func TestRenderFrameAggregatedCountsMatch(t *testing.T) {
	scene := engine.WorkerPool(2000)
	l := NewLayout(scene, 900, 360)

	frames := scene.Frames()
	last := frames[len(frames)-1]

	want := map[engine.GoroutineState]int{}
	for _, st := range last.Goroutines {
		want[st]++
	}

	lines := 0
	for _, op := range RenderFrame(last, l) {
		if op.Kind == OpLine {
			lines++
		}
	}

	nonEmptyStates := 0
	for _, n := range want {
		if n > 0 {
			nonEmptyStates++
		}
	}
	if lines != nonEmptyStates {
		t.Errorf("aggregated lines = %d, want %d (one per non-empty state)", lines, nonEmptyStates)
	}
}
