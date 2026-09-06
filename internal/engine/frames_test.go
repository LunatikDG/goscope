package engine

import "testing"

// A scene small enough to fit under the cap is returned unchanged by SampledFrames.
func TestSampledFramesBelowCapMatchesFrames(t *testing.T) {
	scene := WorkerPool(3)
	want := scene.Frames()
	got := scene.SampledFrames(len(want) + 10)

	if len(got) != len(want) {
		t.Fatalf("frame count = %d, want %d (SampledFrames should fall back to Frames below the cap)", len(got), len(want))
	}
}

// A scene with many more steps than maxFrames is capped, but its last frame
// still reflects the true end state — nothing is skipped that would leave
// live goroutines out of the reported final snapshot.
func TestSampledFramesCapsCountAndKeepsFinalState(t *testing.T) {
	scene := WorkerPool(2000) // thousands of steps: spawn + block + send + unblock + done per worker
	full := scene.Frames()

	const maxFrames = 100
	sampled := scene.SampledFrames(maxFrames)

	if len(sampled) > maxFrames+1 {
		t.Fatalf("sampled frame count = %d, want <= %d", len(sampled), maxFrames+1)
	}
	if len(sampled) >= len(full) {
		t.Fatalf("sampled frame count = %d did not shrink from the full %d frames", len(sampled), len(full))
	}

	wantLast, gotLast := full[len(full)-1], sampled[len(sampled)-1]
	if len(gotLast.Goroutines) != len(wantLast.Goroutines) {
		t.Fatalf("final frame has %d goroutines, want %d", len(gotLast.Goroutines), len(wantLast.Goroutines))
	}
	for id, wantState := range wantLast.Goroutines {
		if got := gotLast.Goroutines[id]; got != wantState {
			t.Errorf("goroutine %d final state = %v, want %v", id, got, wantState)
		}
	}
}

// maxFrames <= 0 is treated as "no cap": SampledFrames falls back to Frames.
func TestSampledFramesNonPositiveMaxIsFullFrames(t *testing.T) {
	scene := WorkerPool(5)
	want := scene.Frames()
	got := scene.SampledFrames(0)

	if len(got) != len(want) {
		t.Fatalf("frame count = %d, want %d", len(got), len(want))
	}
}
