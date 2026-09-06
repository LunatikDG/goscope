package render

import (
	"fmt"

	"github.com/LunatikDG/goscope/internal/engine"
)

func colorFor(st engine.GoroutineState) Color {
	switch st {
	case engine.Running:
		return ColorRunning
	case engine.Blocked:
		return ColorBlocked
	default: // Finished
		return ColorFinished
	}
}

func stateLabel(st engine.GoroutineState) string {
	switch st {
	case engine.Running:
		return "running"
	case engine.Blocked:
		return "blocked"
	default: // Finished
		return "finished"
	}
}

// RenderFrame turns a frame into a list of draw commands (it doesn't draw anything itself).
//
//nolint:revive // name matches the package entry-point for frame→ops conversion
func RenderFrame(f engine.Frame, l Layout) []Op {
	if l.Aggregated() {
		return renderAggregatedFrame(f, l)
	}

	var ops []Op
	top, bottom := l.MarginTop, l.Height-l.MarginBottom

	// 1) goroutine → vertical line, colored by state
	for id, st := range f.Goroutines {
		if x, ok := l.GoroutineX(id); ok {
			ops = append(ops, Op{Kind: OpLine, X1: x, Y1: top, X2: x, Y2: bottom, Color: colorFor(st)})
			if label := l.Label(id); label != "" {
				ops = append(ops, Op{Kind: OpText, X1: x, Y1: top - 10, Color: ColorFinished, Text: label})
			}
		}
	}

	// 2) channel → horizontal link at the moment of send
	if f.Cause != nil && f.Cause.Event == engine.Send {
		if gx, ok := l.GoroutineX(f.Cause.Goroutine); ok {
			if cx, ok := l.ChannelX(f.Cause.Chan); ok {
				midY := (top + bottom) / 2
				ops = append(ops, Op{
					Kind: OpLine, X1: gx, Y1: midY, X2: cx, Y2: midY, Color: ColorChannel,
				})
			}
		}
	}
	return ops
}

// renderAggregatedFrame draws one band per state instead of one line per
// goroutine, so a scene with thousands of goroutines still costs a handful
// of Ops per frame — the count is O(states), not O(goroutines). Per-goroutine
// detail (labels, the exact channel a Send traveled on) doesn't survive
// aggregation; that's the deliberate trade for staying responsive at scale.
func renderAggregatedFrame(f engine.Frame, l Layout) []Op {
	counts := map[engine.GoroutineState]int{}
	for _, st := range f.Goroutines {
		counts[st]++
	}

	top, bottom := l.MarginTop, l.Height-l.MarginBottom
	states := []engine.GoroutineState{engine.Running, engine.Blocked, engine.Finished}

	ops := make([]Op, 0, 2*len(states))
	for _, st := range states {
		n := counts[st]
		if n == 0 {
			continue
		}
		x, ok := l.BandX(st)
		if !ok {
			continue
		}
		ops = append(ops,
			Op{Kind: OpLine, X1: x, Y1: top, X2: x, Y2: bottom, Color: colorFor(st)},
			Op{
				Kind: OpText, X1: x, Y1: top - 10, Color: colorFor(st),
				Text: fmt.Sprintf("%s: %d", stateLabel(st), n),
			},
		)
	}
	return ops
}
