<p align="center">
  <b>English</b> · <a href="README.ru.md">Русский</a>
</p>

<h1 align="center">goscope</h1>

<p align="center">
  <b>Watch Go concurrency come alive.</b><br>
  An interactive visualizer for goroutines, channels, blocking, and completion.
</p>

<p align="center">
  <a href="https://lunatikdg.github.io/goscope/"><b>▶ Live demo</b></a>
</p>


<p align="center">
  <img src="https://github.com/LunatikDG/goscope/actions/workflows/ci.yml/badge.svg" alt="CI">
  <img src="https://github.com/LunatikDG/goscope/actions/workflows/pages.yml/badge.svg" alt="Pages">
  <img src="https://img.shields.io/github/license/LunatikDG/goscope" alt="License">
  <img src="https://img.shields.io/github/go-mod/go-version/LunatikDG/goscope" alt="Go version">
</p>

---

## What is this

`goscope` shows what's usually hidden: how goroutines spawn, block on channels,
unblock, and finish. Go concurrency is hard to picture from code — here it becomes
a visual animation you can play, pause, and step through.

## Features

- ▶ Play / pause / step mode, adjustable speed.
- 🎨 Color = goroutine state: running / blocked / finished.
- 🔗 Channel links at the moment of transfer.
- 🧩 A gallery of patterns: worker pool, fan-in/fan-out, pipeline, deadlock, goroutine leak.
- 🌓 Light/dark theme and permalinks to a specific pattern.
- 📡 **Watch live**: run a real instrumented Go program on the server and stream its
  actual events over SSE — genuine concurrency, not a script.

## Run locally

```bash
git clone https://github.com/LunatikDG/goscope
cd goscope
make wasm        # build WebAssembly + copy wasm_exec.js
make serve       # start a local server
# open http://localhost:8080
```

Or with Docker (bundles the WASM build, the server, and the instrumented examples):

```bash
docker build -t goscope .
docker run --rm -p 8080:8080 goscope
```

## How it works

The core (`internal/engine`) models concurrency as a sequence of events and folds
them into frames. The render layer (`internal/render`) turns a frame into draw
commands — browser-agnostic and fully tested. A thin WebAssembly layer
(`cmd/webdemo`) executes those commands on a canvas.

For the live gallery, `internal/server` runs a real program from `examples/`
(instrumented with `internal/instrument`) as a subprocess and streams its actual
events to the browser over Server-Sent Events — the WASM layer folds them into
frames the same way it does for the canned patterns.

## Status

v0.3 — a gallery of patterns, plus live visualization of real instrumented
programs. Roadmap: visualizing traces of arbitrary programs (`runtime/trace`).
Ideas and PRs welcome — CI runs on every pull request.

## License

MIT © 2026 Dmitry Golovin
