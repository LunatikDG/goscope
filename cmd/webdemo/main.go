//go:build js && wasm

package main

import (
	"syscall/js"
	"time"

	"github.com/LunatikDG/goscope/internal/engine"
	"github.com/LunatikDG/goscope/internal/render"
)

func main() {
	scene := engine.WorkerPool(3)
	frames := scene.Frames()
	layout := render.NewLayout(scene, 640, 360)
	canvas := newCanvas("canvas")

	player := render.NewPlayer(len(frames), 600*time.Millisecond) // 600мс на шаг

	var raf js.Func
	lastMs := 0.0

	var tick func(this js.Value, args []js.Value) any
	tick = func(this js.Value, args []js.Value) any {
		nowMs := args[0].Float() // rAF передаёт timestamp в миллисекундах
		if lastMs == 0 {
			lastMs = nowMs
		}
		dt := time.Duration((nowMs - lastMs) * float64(time.Millisecond))
		lastMs = nowMs

		idx := player.Advance(dt)
		canvas.clear()
		canvas.draw(render.RenderFrame(frames[idx], layout))

		js.Global().Call("requestAnimationFrame", raf) // запросить следующий кадр
		return nil
	}

	raf = js.FuncOf(tick)
	js.Global().Call("requestAnimationFrame", raf)

	select {} // держим программу и колбэк живыми
}
