//go:build js && wasm

package main

import (
	"github.com/LunatikDG/goscope/internal/engine"
	"github.com/LunatikDG/goscope/internal/render"
)

func main() {
	scene := engine.WorkerPool(3)
	frames := scene.Frames()
	layout := render.NewLayout(scene, 640, 360)

	canvas := newCanvas("canvas")
	canvas.clear()
	canvas.draw(render.RenderFrame(frames[0], layout)) // первый статичный кадр

	select {} // держим программу живой (задел под анимацию/колбэки Дня 7–8)
}
