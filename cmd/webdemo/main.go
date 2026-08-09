//go:build js && wasm

package main

import (
	"encoding/json"
	"strconv"
	"strings"
	"syscall/js"
	"time"

	"github.com/LunatikDG/goscope/internal/engine"
	"github.com/LunatikDG/goscope/internal/render"
)

// order — паттерны в порядке показа в навигации: от простого к «сломанному».
var order = []string{"workerpool", "fanin_fanout", "pipeline", "deadlock", "goroutine_leak"}

// liveEvent — схема одной SSE-строки, которую шлёт /api/run/{name};
// повторяет internal/instrument.Event.
type liveEvent struct {
	Event     string `json:"event"`
	Label     string `json:"label"`
	Goroutine int    `json:"goroutine"`
	Chan      int    `json:"chan"`
}

func main() {
	// --- данные и плеер (статичные, заранее просчитанные сцены) ---
	stepEvery := 600 * time.Millisecond // 600мс на шаг (= ползунок 5); меняется ползунком, переживает смену паттерна

	var (
		scene       engine.Scene
		frames      []engine.Frame
		layout      render.Layout
		player      *render.Player
		currentName string
	)

	// --- состояние живого режима (стрим реального примера с сервера по SSE) ---
	var (
		live       bool
		liveFrame  engine.Frame
		liveSteps  []engine.Step
		liveFolder *engine.LiveFolder
		liveFuncs  []js.Func // колбэки текущего живого стрима — освобождаем при остановке
	)
	liveES := js.Undefined() // явно, а не нулевое значение js.Value — им проверяем, есть ли активное соединение

	// canvas делаем изменяемым: resize его пересоздаёт
	canvas := newCanvas("canvas")

	doc := js.Global().Get("document")

	// уважить prefers-reduced-motion: каждый вновь загруженный паттерн стартует на паузе
	reduced := js.Global().Call("matchMedia", "(prefers-reduced-motion: reduce)").Get("matches").Bool()

	// единая перерисовка текущего кадра (для step, resize, цикла и живого стрима)
	redraw := func() {
		if live {
			canvas.clear()
			canvas.draw(render.RenderFrame(liveFrame, layout))
			return
		}
		if len(frames) == 0 {
			return
		}
		canvas.clear()
		canvas.draw(render.RenderFrame(frames[player.Current()], layout))
	}

	playPauseBtn := doc.Call("getElementById", "playPause")
	setPlayLabel := func() {
		if player.Playing() {
			playPauseBtn.Set("textContent", "⏸ Pause")
		} else {
			playPauseBtn.Set("textContent", "▶ Play")
		}
	}

	titleEl := doc.Call("getElementById", "patternTitle")
	descEl := doc.Call("getElementById", "patternDescription")
	navEl := doc.Call("getElementById", "patterns")
	liveStatusEl := doc.Call("getElementById", "liveStatus")
	watchLiveBtn := doc.Call("getElementById", "watchLive")

	// setActiveNav подсвечивает ссылку текущего паттерна в навигации.
	setActiveNav := func(name string) {
		links := navEl.Get("children")
		for i := 0; i < links.Length(); i++ {
			a := links.Index(i)
			classList := a.Get("classList")
			if a.Get("dataset").Get("pattern").String() == name {
				classList.Call("add", "active")
				a.Call("setAttribute", "aria-current", "page")
			} else {
				classList.Call("remove", "active")
				a.Call("removeAttribute", "aria-current")
			}
		}
	}

	// closeLiveConnection закрывает текущий SSE-стрим (если есть) и освобождает его
	// колбэки. Не трогает live: конец стрима не значит выход из живого режима —
	// последний полученный кадр должен остаться на экране, а не смениться кадром
	// фонового статичного плеера (тот всё это время тихо тикает по rAF).
	closeLiveConnection := func() {
		if liveES.IsUndefined() {
			return
		}
		liveES.Call("close")
		liveES = js.Undefined()
		for _, f := range liveFuncs {
			f.Release()
		}
		liveFuncs = nil
	}

	// loadPattern грузит встроенную сцену по имени и полностью пересобирает плеер/раскладку/UI под неё.
	loadPattern := func(name string) {
		s, err := engine.LoadScene(name)
		if err != nil {
			js.Global().Get("console").Call("error", "goscope: не удалось загрузить сцену "+name+": "+err.Error())
			return
		}
		closeLiveConnection()
		live = false
		liveStatusEl.Set("textContent", "")

		currentName = name
		watchLiveBtn.Set("textContent", "▶ Watch live: "+s.Name)

		scene = s
		frames = scene.Frames()
		layout = render.NewLayout(scene, canvas.width, canvas.height)
		player = render.NewPlayer(len(frames), stepEvery)
		if reduced {
			player.Pause()
		}
		titleEl.Set("textContent", scene.Name)
		descEl.Set("textContent", scene.Description)
		setActiveNav(name)
		setPlayLabel()
		redraw()
	}

	// строим навигацию один раз: имя + описание берём прямо из встроенных сцен —
	// единственный источник истины, дублировать их в HTML не нужно.
	for _, name := range order {
		s, err := engine.LoadScene(name)
		if err != nil {
			continue
		}

		a := doc.Call("createElement", "a")
		a.Set("href", "#"+name)
		a.Set("className", "pattern-link")
		a.Get("dataset").Set("pattern", name)

		title := doc.Call("createElement", "strong")
		title.Set("textContent", s.Name)
		a.Call("appendChild", title)

		desc := doc.Call("createElement", "span")
		desc.Set("textContent", s.Description)
		a.Call("appendChild", desc)

		navEl.Call("appendChild", a)
	}

	validNames := make(map[string]bool, len(order))
	for _, name := range order {
		validNames[name] = true
	}

	// patternFromHash читает #имя-паттерна из адресной строки — это и есть пермалинк.
	patternFromHash := func() string {
		h := strings.TrimPrefix(js.Global().Get("location").Get("hash").String(), "#")
		if !validNames[h] {
			return order[0]
		}
		return h
	}

	loadPattern(patternFromHash())

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

		player.Advance(dt) // на паузе (или в живом режиме) вернёт тот же кадр
		redraw()

		js.Global().Call("requestAnimationFrame", raf) // ← самоподдержка цикла (без этого — стоп)
		return nil
	}
	raf = keep(js.FuncOf(tick))

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

	// --- «Watch live»: запускает настоящий пример на сервере и стримит его события по SSE ---
	on("watchLive", "click", func() {
		closeLiveConnection()

		live = true
		liveSteps = nil
		liveFolder = engine.NewLiveFolder()
		liveFrame = engine.Frame{Goroutines: map[int]engine.GoroutineState{}}
		layout = render.NewLayout(engine.Scene{}, canvas.width, canvas.height)
		liveStatusEl.Set("textContent", "● connecting…")
		redraw()

		ended := false
		addLive := func(f js.Func) js.Func { liveFuncs = append(liveFuncs, f); return f }

		es := js.Global().Get("EventSource").New("/api/run/" + currentName)
		liveES = es

		es.Set("onopen", addLive(js.FuncOf(func(this js.Value, args []js.Value) any {
			liveStatusEl.Set("textContent", "● live")
			return nil
		})))

		es.Set("onmessage", addLive(js.FuncOf(func(this js.Value, args []js.Value) any {
			var le liveEvent
			if err := json.Unmarshal([]byte(args[0].Get("data").String()), &le); err != nil {
				return nil
			}
			event, err := engine.ParseEventType(le.Event)
			if err != nil {
				return nil
			}
			step := engine.Step{Event: event, Goroutine: le.Goroutine, Chan: le.Chan, Label: le.Label}

			liveSteps = append(liveSteps, step)
			layout = render.NewLayout(engine.Scene{Steps: liveSteps}, canvas.width, canvas.height)
			liveFrame = liveFolder.Apply(step)
			redraw()
			return nil
		})))

		es.Call("addEventListener", "end", addLive(js.FuncOf(func(this js.Value, args []js.Value) any {
			ended = true
			liveStatusEl.Set("textContent", "● finished")
			closeLiveConnection()
			return nil
		})))

		es.Set("onerror", addLive(js.FuncOf(func(this js.Value, args []js.Value) any {
			if !ended {
				liveStatusEl.Set("textContent", "⚠ live examples need the local dev server (make serve)")
			}
			closeLiveConnection()
			return nil
		})))
	})

	// --- переход по пермалинку: клик по навигации меняет location.hash и рождает hashchange ---
	hashCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		loadPattern(patternFromHash())
		return nil
	})
	js.Global().Call("addEventListener", "hashchange", keep(hashCb))

	// --- адаптив под ширину окна ---
	resizeCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		canvas = newCanvas("canvas") // пересчитать dpr/размеры
		if live {
			layout = render.NewLayout(engine.Scene{Steps: liveSteps}, canvas.width, canvas.height)
		} else {
			layout = render.NewLayout(scene, canvas.width, canvas.height)
		}
		redraw()
		return nil
	})
	js.Global().Call("addEventListener", "resize", keep(resizeCb))

	// первый кадр + запуск цикла
	redraw()
	js.Global().Call("requestAnimationFrame", raf)

	select {} // держим программу и колбэки живыми
}
