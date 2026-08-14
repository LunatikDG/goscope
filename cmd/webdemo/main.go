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

// order — patterns in the order they're shown in the nav: from simple to "broken".
var order = []string{"workerpool", "fanin_fanout", "pipeline", "deadlock", "goroutine_leak"}

// liveEvent — the schema of one SSE line sent by /api/run/{name};
// mirrors internal/instrument.Event.
type liveEvent struct {
	Event     string `json:"event"`
	Label     string `json:"label"`
	Goroutine int    `json:"goroutine"`
	Chan      int    `json:"chan"`
}

func main() {
	// --- data and player (static, precomputed scenes) ---
	stepEvery := 600 * time.Millisecond // 600ms per step (= slider at 5); changed by the slider, survives switching patterns

	var (
		scene       engine.Scene
		frames      []engine.Frame
		layout      render.Layout
		player      *render.Player
		currentName string
	)

	// --- live-mode state (streaming a real example from the server over SSE) ---
	var (
		live       bool
		liveFrame  engine.Frame
		liveSteps  []engine.Step
		liveFolder *engine.LiveFolder
		liveFuncs  []js.Func // callbacks of the current live stream — released when it stops
	)
	liveES := js.Undefined() // explicit, not js.Value's zero value — we use it to check whether a connection is active

	// canvas is mutable: resize recreates it
	canvas := newCanvas("canvas")

	doc := js.Global().Get("document")

	// respect prefers-reduced-motion: every newly loaded pattern starts paused
	reduced := js.Global().Call("matchMedia", "(prefers-reduced-motion: reduce)").Get("matches").Bool()

	// a single redraw for the current frame (used by step, resize, the tick loop, and the live stream)
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

	// setActiveNav highlights the current pattern's link in the nav.
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

	// closeLiveConnection closes the current SSE stream (if any) and releases its
	// callbacks. It doesn't touch live: the stream ending doesn't mean leaving
	// live mode — the last frame received should stay on screen, not get
	// replaced by a frame from the background static player (which keeps
	// quietly ticking via rAF the whole time).
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

	// loadPattern loads a built-in scene by name and fully rebuilds the player/layout/UI for it.
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

	// build the nav once: name + description come straight from the built-in
	// scenes — the single source of truth, no need to duplicate them in the HTML.
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

	// patternFromHash reads #pattern-name from the address bar — that's the permalink.
	patternFromHash := func() string {
		h := strings.TrimPrefix(js.Global().Get("location").Get("hash").String(), "#")
		if !validNames[h] {
			return order[0]
		}
		return h
	}

	loadPattern(patternFromHash())

	// keep all js callbacks alive for the whole session
	var handlers []js.Func
	keep := func(f js.Func) js.Func { handlers = append(handlers, f); return f }

	// --- rAF loop (autoplay) ---
	var raf js.Func
	lastMs := 0.0
	tick := func(this js.Value, args []js.Value) any {
		nowMs := args[0].Float() // rAF passes the timestamp in milliseconds
		if lastMs == 0 {
			lastMs = nowMs
		}
		dt := time.Duration((nowMs - lastMs) * float64(time.Millisecond))
		lastMs = nowMs

		player.Advance(dt) // returns the same frame while paused (or in live mode)
		redraw()

		js.Global().Call("requestAnimationFrame", raf) // ← keeps the loop going (without this it stops)
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
		player.StepForward() // step one frame forward + pause
		redraw()             // rAF doesn't change the frame while paused — draw it manually
		setPlayLabel()
	})

	on("restart", "click", func() {
		player.Restart()
		setPlayLabel()
	})

	// --- speed slider: 1..10 → step duration (inverted) ---
	on("speed", "input", func() {
		v := doc.Call("getElementById", "speed").Get("value").String()
		level, err := strconv.Atoi(v)
		if err != nil {
			return
		}
		// 1 (slow) → 1000ms ... 10 (fast) → 100ms
		stepEvery = time.Duration(1100-level*100) * time.Millisecond
		player.SetStepEvery(stepEvery)
	})

	// --- "Watch live": runs a real example on the server and streams its events over SSE ---
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

	// --- permalink navigation: clicking a nav link changes location.hash and fires hashchange ---
	hashCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		loadPattern(patternFromHash())
		return nil
	})
	js.Global().Call("addEventListener", "hashchange", keep(hashCb))

	// --- responsive to window width ---
	resizeCb := js.FuncOf(func(this js.Value, args []js.Value) any {
		canvas = newCanvas("canvas") // recompute dpr/dimensions
		if live {
			layout = render.NewLayout(engine.Scene{Steps: liveSteps}, canvas.width, canvas.height)
		} else {
			layout = render.NewLayout(scene, canvas.width, canvas.height)
		}
		redraw()
		return nil
	})
	js.Global().Call("addEventListener", "resize", keep(resizeCb))

	// first frame + start the loop
	redraw()
	js.Global().Call("requestAnimationFrame", raf)

	select {} // keep the program and its callbacks alive
}
