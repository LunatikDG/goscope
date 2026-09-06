package render

import (
	"sort"

	"github.com/LunatikDG/goscope/internal/engine"
)

// MaxLanes caps how many goroutines get their own column. Scenes with more
// goroutines than this switch Layout into aggregated mode (see Aggregated):
// past this point individual lanes would be sub-pixel wide anyway, and
// building/drawing one line per goroutine per frame is what makes a
// thousands-of-goroutines scene (a real trace, or an extreme goroutine leak)
// hang the browser.
const MaxLanes = 40

// Layout assigns each goroutine and channel a fixed column (X), computed
// over the whole scene so lines don't "jump" between frames.
type Layout struct {
	gLanes                  map[int]float64                   // goroutine id → X (empty in aggregated mode)
	cLanes                  map[int]float64                   // channel id → X (empty in aggregated mode)
	labels                  map[int]string                    // goroutine id → label
	bandX                   map[engine.GoroutineState]float64 // aggregated mode only: state → X
	Width, Height           float64
	MarginTop, MarginBottom float64
	aggregated              bool
}

// NewLayout scans the whole scene and lays goroutines/channels out into columns.
// Once the goroutine count exceeds MaxLanes, it instead reserves one column
// per state for an aggregated (count-based) rendering — see Aggregated.
func NewLayout(s engine.Scene, w, h float64) Layout {
	const marginX = 40.0 // side margins so the outermost lines don't stick to the edges

	gset, cset := map[int]bool{}, map[int]bool{}
	labels := map[int]string{}

	for _, st := range s.Steps {
		gset[st.Goroutine] = true
		if st.Chan != 0 {
			cset[st.Chan] = true
		}
		if st.Label != "" {
			labels[st.Goroutine] = st.Label
		}
	}
	gids, cids := sortedKeys(gset), sortedKeys(cset)

	l := Layout{
		Width: w, Height: h,
		MarginTop: 30, MarginBottom: 30,
		gLanes: map[int]float64{},
		cLanes: map[int]float64{},
		labels: labels,
	}

	// usable width minus the side margins (falls back for a narrow canvas)
	usable, startX := w-2*marginX, marginX
	if usable <= 0 {
		usable, startX = w, 0
	}

	if len(gids) > MaxLanes {
		l.aggregated = true
		l.bandX = bandLayout(startX, usable)
		return l
	}

	total := len(gids) + len(cids)
	if total == 0 {
		return l
	}
	gap := usable / float64(total+1)

	x := startX + gap
	for _, id := range gids {
		l.gLanes[id] = x
		x += gap
	}
	for _, id := range cids {
		l.cLanes[id] = x
		x += gap
	}
	return l
}

// bandLayout spaces the three aggregated-mode bands (one per GoroutineState) evenly.
func bandLayout(startX, usable float64) map[engine.GoroutineState]float64 {
	states := []engine.GoroutineState{engine.Running, engine.Blocked, engine.Finished}
	gap := usable / float64(len(states)+1)
	m := make(map[engine.GoroutineState]float64, len(states))
	x := startX + gap
	for _, st := range states {
		m[st] = x
		x += gap
	}
	return m
}

func (l Layout) GoroutineX(id int) (float64, bool) { x, ok := l.gLanes[id]; return x, ok }
func (l Layout) ChannelX(id int) (float64, bool)   { x, ok := l.cLanes[id]; return x, ok }
func (l Layout) Label(id int) string               { return l.labels[id] }

// Aggregated reports whether this layout is in aggregated (count-based) mode
// because the scene has more than MaxLanes distinct goroutines.
func (l Layout) Aggregated() bool { return l.aggregated }

// BandX returns the X column reserved for a given state in aggregated mode.
func (l Layout) BandX(st engine.GoroutineState) (float64, bool) { x, ok := l.bandX[st]; return x, ok }

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}
