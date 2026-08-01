//go:build js && wasm

package main

import (
	"strconv"
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

	draw := func(idx int) {
		canvas.clear()
		canvas.draw(render.RenderFrame(frames[idx], layout))
	}
	draw(player.Current()) // стартовый кадр сразу

	doc := js.Global().Get("document")

	// держим колбэки живыми весь сеанс
	var handlers []js.Func

	playPauseBtn := doc.Call("getElementById", "playPause")
	handlers = append(handlers,
		on("playPause", "click", func() {
			if player.Playing() {
				player.Pause()
				playPauseBtn.Set("textContent", "▶ Play")
			} else {
				player.Play()
				playPauseBtn.Set("textContent", "⏸ Pause")
			}
		}),
		on("step", "click", func() {
			draw(player.StepForward())
			playPauseBtn.Set("textContent", "▶ Play")
		}),
		on("restart", "click", func() {
			draw(player.Restart())
			playPauseBtn.Set("textContent", "⏸ Pause")
		}),
	)

	// ползунок скорости: 1..10 → инвертируем в длительность шага
	speedCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		v := doc.Call("getElementById", "speed").Get("value").String()
		level, _ := strconv.Atoi(v) // 1..10
		// 1 (медленно) → ~1000мс, 10 (быстро) → ~100мс
		player.SetStepEvery(time.Duration(1100-level*100) * time.Millisecond)
		return nil
	})
	doc.Call("getElementById", "speed").Call("addEventListener", "input", speedCb)
	handlers = append(handlers, speedCb)

	var raf js.Func
	lastMs := 0.0
	raf = js.FuncOf(func(this js.Value, args []js.Value) any {
		nowMs := args[0].Float() // rAF передаёт timestamp в миллисекундах
		if lastMs == 0 {
			lastMs = nowMs
		}
		dt := time.Duration((nowMs - lastMs) * float64(time.Millisecond))
		lastMs = nowMs

		draw(player.Advance(dt))
		js.Global().Call("requestAnimationFrame", raf)
		return nil
	})
	handlers = append(handlers, raf)
	js.Global().Call("requestAnimationFrame", raf)

	_ = handlers // просто держим ссылки живыми
	select {}    // держим программу и колбэки живыми
}

// on навешивает обработчик события на элемент по id и сохраняет js.Func живым.
func on(id, event string, fn func()) js.Func {
	cb := js.FuncOf(func(this js.Value, args []js.Value) any {
		fn()
		return nil
	})
	js.Global().Get("document").
		Call("getElementById", id).
		Call("addEventListener", event, cb)
	return cb
}
