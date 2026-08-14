package render

import (
	"sort"

	"github.com/LunatikDG/goscope/internal/engine"
)

// Layout assigns each goroutine and channel a fixed column (X), computed
// over the whole scene so lines don't "jump" between frames.
type Layout struct {
	gLanes                  map[int]float64 // goroutine id → X
	cLanes                  map[int]float64 // channel id → X
	labels                  map[int]string  // goroutine id → label
	Width, Height           float64
	MarginTop, MarginBottom float64
}

// NewLayout scans the whole scene and lays goroutines/channels out into columns.
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

	total := len(gids) + len(cids)
	if total == 0 {
		return l
	}

	// usable width minus the side margins (falls back for a narrow canvas)
	usable, startX := w-2*marginX, marginX
	if usable <= 0 {
		usable, startX = w, 0
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

func (l Layout) GoroutineX(id int) (float64, bool) { x, ok := l.gLanes[id]; return x, ok }
func (l Layout) ChannelX(id int) (float64, bool)   { x, ok := l.cLanes[id]; return x, ok }
func (l Layout) Label(id int) string               { return l.labels[id] }

func sortedKeys(m map[int]bool) []int {
	out := make([]int, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	sort.Ints(out)
	return out
}
