package render

import (
	"testing"

	"github.com/LunatikDG/goscope/internal/engine"
)

// Вертикальных линий столько же, сколько горутин в кадре.
func TestRenderFrameVerticalLines(t *testing.T) {
	scene := engine.WorkerPool(3)
	l := NewLayout(scene, 640, 360)
	f := scene.Frames()[len(scene.Steps)/2] // кадр из середины

	lines := 0
	for _, op := range RenderFrame(f, l) {
		if op.Kind == OpLine && op.X1 == op.X2 { // вертикальная
			lines++
		}
	}
	if lines != len(f.Goroutines) {
		t.Errorf("вертикальных линий = %d, ожидалось %d", lines, len(f.Goroutines))
	}
}

// Цвет линии соответствует состоянию.
func TestRenderFrameColorByState(t *testing.T) {
	l := NewLayout(engine.Scene{Steps: []engine.Step{{Event: engine.Spawn, Goroutine: 1}}}, 640, 360)
	f := engine.Frame{Goroutines: map[int]engine.GoroutineState{1: engine.Blocked}}

	ops := RenderFrame(f, l)
	if len(ops) != 1 || ops[0].Color != ColorBlocked {
		t.Errorf("ожидалась одна красная линия для Blocked, получено %+v", ops)
	}
}

// В момент Send появляется ровно одна горизонтальная связь канала.
func TestRenderFrameSendConnector(t *testing.T) {
	scene := engine.Scene{Steps: []engine.Step{
		{Event: engine.Spawn, Goroutine: 0},
		{Event: engine.Send, Goroutine: 0, Chan: 1},
	}}
	l := NewLayout(scene, 640, 360)
	sendFrame := scene.Frames()[2] // кадр после Send

	horizontal := 0
	for _, op := range RenderFrame(sendFrame, l) {
		if op.Kind == OpLine && op.Y1 == op.Y2 && op.Color == ColorChannel {
			horizontal++
		}
	}
	if horizontal != 1 {
		t.Errorf("ожидалась 1 горизонтальная связь канала, получено %d", horizontal)
	}
}
