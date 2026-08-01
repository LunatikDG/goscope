//go:build js && wasm

package main

import (
	"syscall/js"
	"time"
	"strconv"

	"github.com/LunatikDG/goscope/internal/engine"
	"github.com/LunatikDG/goscope/internal/render"
)

func main() {
	scene := engine.WorkerPool(3)
	frames := scene.Frames()
	layout := render.NewLayout(scene, 640, 360)
	canvas := newCanvas("canvas")

	player := render.NewPlayer(len(frames), 600*time.Millisecond) // 600мс на шаг

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
			idx := player.StepForward()
			canvas.clear()
			canvas.draw(render.RenderFrame(frames[idx], layout)) // сразу перерисовать на паузе
			playPauseBtn.Set("textContent", "▶ Play")
		}),
		on("restart", "click", func() {
			player.Restart()
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

	_ = handlers // просто держим ссылки живыми

	select {} // держим программу и колбэк живыми
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
