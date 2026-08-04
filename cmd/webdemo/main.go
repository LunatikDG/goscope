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
	// --- данные и плеер ---
	stepEvery := 600 * time.Millisecond // 600мс на шаг (= ползунок 5); меняется ползунком, переживает смену паттерна

	var (
		scene  engine.Scene
		frames []engine.Frame
		layout render.Layout
		player *render.Player
	)

	// canvas делаем изменяемым: resize его пересоздаёт
	canvas := newCanvas("canvas")

	doc := js.Global().Get("document")

	// уважить prefers-reduced-motion: каждый вновь загруженный паттерн стартует на паузе
	reduced := js.Global().Call("matchMedia", "(prefers-reduced-motion: reduce)").Get("matches").Bool()

	// единая перерисовка текущего кадра (для step, resize и цикла)
	redraw := func() {
		if len(frames) == 0 {
			return
		}
		canvas.clear()
		canvas.draw(render.RenderFrame(frames[player.Current()], layout))
	}

	// loadPattern грузит встроенную сцену по имени и полностью пересобирает плеер/раскладку под неё.
	loadPattern := func(name string) {
		s, err := engine.LoadScene(name)
		if err != nil {
			js.Global().Get("console").Call("error", "goscope: не удалось загрузить сцену "+name+": "+err.Error())
			return
		}
		scene = s
		frames = scene.Frames()
		layout = render.NewLayout(scene, canvas.width, canvas.height)
		player = render.NewPlayer(len(frames), stepEvery)
		if reduced {
			player.Pause()
		}
		redraw()
	}

	loadPattern("workerpool")

	// держим все js-колбэки живыми весь сеанс
	var handlers []js.Func
	keep := func(f js.Func) js.Func { handlers = append(handlers, f); return f }

	// --- rAF-цикл (автопрогон) ---
	var raf js.Func
	lastMs := 0.0
	tick := func(this js.Value, args []js.Value) any {
		nowMs := args[0].Float() // rAF передаёт timestamp в миллисекундах
		if lastMs == 0 {
			lastMs = nowMs
		}
		dt := time.Duration((nowMs - lastMs) * float64(time.Millisecond))
		lastMs = nowMs

		player.Advance(dt) // на паузе вернёт тот же кадр
		redraw()

		js.Global().Call("requestAnimationFrame", raf) // ← самоподдержка цикла (без этого — стоп)
		return nil
	}
	raf = keep(js.FuncOf(tick))

	// --- кнопки ---
	playPauseBtn := doc.Call("getElementById", "playPause")
	setPlayLabel := func() {
		if player.Playing() {
			playPauseBtn.Set("textContent", "⏸ Pause")
		} else {
			playPauseBtn.Set("textContent", "▶ Play")
		}
	}

	on := func(id, event string, fn func()) {
		cb := js.FuncOf(func(this js.Value, args []js.Value) any {
			fn()
			return nil
		})
		doc.Call("getElementById", id).Call("addEventListener", event, keep(cb))
	}

	on("playPause", "click", func() {
		if player.Playing() {
			player.Pause()
		} else {
			player.Play()
		}
		setPlayLabel()
	})

	on("step", "click", func() {
		player.StepForward() // сдвиг на кадр + пауза
		redraw()             // на паузе rAF кадр не меняет — рисуем вручную
		setPlayLabel()
	})

	on("restart", "click", func() {
		player.Restart()
		setPlayLabel()
	})

	// --- ползунок скорости: 1..10 → длительность шага (инверсия) ---
	on("speed", "input", func() {
		v := doc.Call("getElementById", "speed").Get("value").String()
		level, err := strconv.Atoi(v)
		if err != nil {
			return
		}
		// 1 (медленно) → 1000мс ... 10 (быстро) → 100мс
		stepEvery = time.Duration(1100-level*100) * time.Millisecond
		player.SetStepEvery(stepEvery)
	})

	// --- выбор паттерна: полная перезагрузка сцены/плеера/раскладки ---
	on("pattern", "change", func() {
		v := doc.Call("getElementById", "pattern").Get("value").String()
		loadPattern(v)
		setPlayLabel()
	})

	// --- адаптив под ширину окна ---
	resizeCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		canvas = newCanvas("canvas")                                  // пересчитать dpr/размеры
		layout = render.NewLayout(scene, canvas.width, canvas.height) // новая раскладка
		redraw()
		return nil
	})
	js.Global().Call("addEventListener", "resize", keep(resizeCb))

	setPlayLabel()

	// первый кадр + запуск цикла
	redraw()
	js.Global().Call("requestAnimationFrame", raf)

	select {} // держим программу и колбэки живыми
}
