package render

import "github.com/LunatikDG/goscope/internal/engine"

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

// RenderFrame превращает кадр в список команд отрисовки (сам не рисует).
//
//nolint:revive // name matches the package entry-point for frame→ops conversion
func RenderFrame(f engine.Frame, l Layout) []Op {
	var ops []Op
	top, bottom := l.MarginTop, l.Height-l.MarginBottom

	// 1) горутина → вертикальная линия, цвет по состоянию
	for id, st := range f.Goroutines {
		if x, ok := l.GoroutineX(id); ok {
			ops = append(ops, Op{
				Kind: OpLine, X1: x, Y1: top, X2: x, Y2: bottom, Color: colorFor(st),
			})
		}
	}

	// 2) канал → горизонтальная связь в момент send
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
