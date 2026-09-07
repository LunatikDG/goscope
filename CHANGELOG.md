<p align="center">
  <b>English</b> · <a href="CHANGELOG.ru.md">Русский</a>
</p>

# Changelog

All notable changes to this project are documented in this file.
The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/).

## [0.5.0] — 2026-09-07

### Added
- `internal/tracer`: a thin wrapper around `runtime/trace.Start`/`Stop`, plus `cmd/traced` — an example program that records a real trace to a file.
- `internal/traceparse`: parses a `runtime/trace` stream via `golang.org/x/exp/trace` into `engine.Step` values — goroutine spawn/block/unblock/done and GC ranges — so real traces flow through the same engine the built-in patterns use.
- Proof that parsed real-trace steps render through the existing `engine`/`render` pipeline unchanged — no special-casing needed for real data versus scripted patterns.
- Scale: `Scene.SampledFrames` snapshots state at evenly spaced points instead of after every single step once a scene has more steps than a cap, and `render.Layout` switches into an aggregated (count-based) rendering mode past a goroutine-count threshold — together keeping a scene with thousands of goroutines (a real trace, or an extreme leak) from hanging the browser.
- `cmd/goscope`: a CLI — `goscope trace <path>` parses a trace file, serves it to the existing WASM demo at `GET /api/trace`, and opens the visualization in the browser. `cmd/webdemo` picks this endpoint up automatically, falling back to the built-in pattern gallery when it's absent.

## [0.3.0] — 2026-08-14

### Added
- Declarative YAML scene format: a loader with validation (`internal/engine`), scenes embedded into the binary via `go:embed`.
- A pattern library: fan-in/fan-out, pipeline, deadlock, goroutine leak — alongside worker pool.
- Pattern navigation with descriptions, a light/dark theme (persisted), permalinks to a specific pattern (`#pattern-name`).
- **Watch live**: the server runs real instrumented Go programs (`examples/`) as a subprocess and streams their actual events to the browser over Server-Sent Events — genuine concurrency, not a script.
- `internal/server`: config from environment variables (`GOSCOPE_*`), structured logging via `log/slog`, transport layer covered by `httptest`.
- Multi-stage Docker image build; CI additionally verifies the image builds on every PR.

### Changed
- `cmd/serve` is now a thin entrypoint over `internal/server`, with graceful shutdown on SIGINT/SIGTERM.

## [0.1.0] — 2026-08-02

### Added
- An interactive visualization of Go concurrency on WebAssembly.
- The worker pool pattern: spawn, block on a channel, unblock, finish.
- Controls: play / pause / step, adjustable speed.
- Color-coded states and channel links; labels and a legend.
- A live demo on GitHub Pages, CI (lint + race tests + WASM build).

[0.5.0]: https://github.com/LunatikDG/goscope/releases/tag/v0.5.0
[0.3.0]: https://github.com/LunatikDG/goscope/releases/tag/v0.3.0
[0.1.0]: https://github.com/LunatikDG/goscope/releases/tag/v0.1.0
