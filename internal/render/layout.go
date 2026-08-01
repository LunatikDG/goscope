package render

import (
	"sort"

	"github.com/LunatikDG/goscope/internal/engine"
)

type Layout struct {
	gLanes                  map[int]float64
	cLanes                  map[int]float64
	labels                  map[int]string // id горутины → подпись
	Width, Height           float64
	MarginTop, MarginBottom float64
}

// NewLayout сканирует всю сцену и назначает каждой горутине/каналу колонку.
func NewLayout(s engine.Scene, w, h float64) Layout {
	gset, cset := map[int]bool{}, map[int]bool{}
	labels := map[int]string{} // ← НОВОЕ: id горутины → подпись

	for _, st := range s.Steps {
		gset[st.Goroutine] = true
		if st.Chan != 0 {
			cset[st.Chan] = true
		}
		if st.Label != "" { // ← НОВОЕ
			labels[st.Goroutine] = st.Label // ← НОВОЕ
		}
	}
	gids, cids := sortedKeys(gset), sortedKeys(cset)

	l := Layout{
		Width: w, Height: h, MarginTop: 30, MarginBottom: 30,
		gLanes: map[int]float64{}, cLanes: map[int]float64{},
		labels: labels,
	}
	total := len(gids) + len(cids)
	if total == 0 {
		return l
	}
	gap := w / float64(total+1)
	x := gap
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
