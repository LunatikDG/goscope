package render

import (
	"testing"

	"github.com/LunatikDG/goscope/internal/engine"
)

func TestRenderFrameHasLabels(t *testing.T) {
	scene := engine.WorkerPool(2)
	l := NewLayout(scene, 640, 360)
	f := scene.Frames()[1] // после спавна диспетчера

	hasText := false
	for _, op := range RenderFrame(f, l) {
		if op.Kind == OpText && op.Text == "dispatcher" {
			hasText = true
		}
	}
	if !hasText {
		t.Error("ожидалась подпись 'dispatcher' в кадре")
	}
}
