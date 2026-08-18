// Package traceparse reads a runtime/trace stream (as produced by
// internal/tracer) via golang.org/x/exp/trace and converts goroutine
// lifecycle, channel-blocking, and GC events into engine.Step values, in the
// order they occurred — the same shape internal/instrument produces for the
// live browser demo, so a real trace can eventually flow through the same
// Frames()/LiveFolder pipeline.
//
// This is a best-effort, lossy conversion:
//   - The trace format doesn't expose which specific channel a goroutine
//     blocked on, only that it did (via the state-transition Reason, e.g.
//     "chan send") — so Step.Chan is always left at 0. That Reason is kept
//     in Step.Label for context.
//   - GC ranges are usually scoped to the whole runtime rather than one
//     goroutine; Step.Goroutine is 0 for those, a placeholder "global actor".
package traceparse

import (
	"errors"
	"fmt"
	"io"
	"strings"

	texp "golang.org/x/exp/trace"

	"github.com/LunatikDG/goscope/internal/engine"
)

// Parse reads r as a runtime/trace stream and returns the events it
// recognizes as engine.Step values, in trace order.
func Parse(r io.Reader) ([]engine.Step, error) {
	tr, err := texp.NewReader(r)
	if err != nil {
		return nil, fmt.Errorf("open trace: %w", err)
	}

	var steps []engine.Step
	for {
		ev, err := tr.ReadEvent()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read event: %w", err)
		}

		switch ev.Kind() {
		case texp.EventStateTransition:
			if step, ok := goroutineStep(ev); ok {
				steps = append(steps, step)
			}
		case texp.EventRangeBegin, texp.EventRangeActive:
			if step, ok := gcStep(ev, engine.GCStart); ok {
				steps = append(steps, step)
			}
		case texp.EventRangeEnd:
			if step, ok := gcStep(ev, engine.GCEnd); ok {
				steps = append(steps, step)
			}
		default:
			// Sync, Metric, Label, StackSample, Task*, Region*, Log,
			// Experimental — nothing we translate into a Step (yet).
		}
	}
	return steps, nil
}

// goroutineStep turns a goroutine state transition into a Step. Order of the
// cases matters: a goroutine that was never observed before tracing began
// (GoUndetermined) and immediately transitions to GoWaiting is a Block, not
// a Spawn — the "first appearance" heuristic for Spawn only applies when the
// goroutine isn't also blocking in that same event.
//
//nolint:gocritic // texp.Event is a value type by design in golang.org/x/exp/trace's API; not ours to repack
func goroutineStep(ev texp.Event) (engine.Step, bool) {
	st := ev.StateTransition()
	if st.Resource.Kind != texp.ResourceGoroutine {
		return engine.Step{}, false
	}
	id := int(st.Resource.Goroutine())
	from, to := st.Goroutine()

	switch {
	case to == texp.GoNotExist:
		return engine.Step{Event: engine.Done, Goroutine: id}, true
	case to == texp.GoWaiting:
		return engine.Step{Event: engine.Block, Goroutine: id, Label: st.Reason}, true
	case from == texp.GoNotExist, from == texp.GoUndetermined:
		return engine.Step{Event: engine.Spawn, Goroutine: id}, true
	case from == texp.GoWaiting:
		return engine.Step{Event: engine.Unblock, Goroutine: id, Label: st.Reason}, true
	default:
		// Scheduling noise we don't represent: e.g. Runnable->Running,
		// Running->Runnable (preemption), Running->Syscall and back.
		return engine.Step{}, false
	}
}

// gcStep turns a range event into a GC step if the range looks GC-related.
// Real range names observed in practice include "GC concurrent mark phase",
// "stop-the-world (GC sweep termination)", and "stop-the-world (GC mark
// termination)" — hence a substring match rather than an exact/prefix one.
//
//nolint:gocritic // texp.Event is a value type by design in golang.org/x/exp/trace's API; not ours to repack
func gcStep(ev texp.Event, kind engine.EventType) (engine.Step, bool) {
	rg := ev.Range()
	if !strings.Contains(rg.Name, "GC") {
		return engine.Step{}, false
	}

	goroutine := 0
	if g := ev.Goroutine(); g != texp.NoGoroutine {
		goroutine = int(g)
	}
	return engine.Step{Event: kind, Goroutine: goroutine, Label: rg.Name}, true
}
