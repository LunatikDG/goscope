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
  <img src="docs/demo.gif" alt="Demo: worker pool visualization" width="800">
</p>

<p align="center">
  <img src="https://github.com/LunatikDG/goscope/actions/workflows/pages.yml/badge.svg" alt="CI">
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
- 🧩 Built-in worker pool pattern (more to come).

## Run locally

```bash
git clone https://github.com/LunatikDG/goscope
cd goscope
make wasm        # build WebAssembly + copy wasm_exec.js
make serve       # start a local server
# open http://localhost:8080
```

## How it works

The core (`internal/engine`) models concurrency as a sequence of events and folds
them into frames. The render layer (`internal/render`) turns a frame into draw
commands — browser-agnostic and fully tested. A thin WebAssembly layer
(`cmd/webdemo`) executes those commands on a canvas.

## Status

v0.1 — visualizes a single pattern. Roadmap: a gallery of patterns → visualizing
traces of **real** programs (`runtime/trace`). Ideas and PRs welcome.

## License

MIT © 2026 Dmitry Golovin
