<p align="center">
  <a href="CHANGELOG.md">English</a> · <b>Русский</b>
</p>

# Changelog

Все заметные изменения проекта документируются в этом файле.
Формат основан на [Keep a Changelog](https://keepachangelog.com/ru/1.1.0/),
проект следует [семантическому версионированию](https://semver.org/lang/ru/).

## [0.3.0] — 2026-08-14

### Added
- Декларативный YAML-формат сцен: загрузчик с валидацией (`internal/engine`), сцены встроены в бинарник через `go:embed`.
- Библиотека паттернов: fan-in/fan-out, pipeline, deadlock, goroutine leak — в дополнение к worker pool.
- Навигация по паттернам с описаниями, светлая/тёмная тема (с сохранением выбора), пермалинки на конкретный паттерн (`#pattern-name`).
- «Watch live»: сервер запускает настоящие инструментированные Go-программы (`examples/`) подпроцессом и стримит их реальные события в браузер по Server-Sent Events — конкурентность подлинная, а не по сценарию.
- `internal/server`: конфиг из переменных окружения (`GOSCOPE_*`), структурированное логирование через `log/slog`, транспортный слой покрыт `httptest`-тестами.
- Docker multi-stage сборка образа; CI дополнительно проверяет, что образ собирается на каждый PR.

### Changed
- `cmd/serve` стал тонким entrypoint поверх `internal/server`, с graceful shutdown по SIGINT/SIGTERM.

## [0.1.0] — 2026-08-02

### Added
- Интерактивная визуализация конкурентности Go на WebAssembly.
- Паттерн worker pool: спавн, блокировка на канале, разблокировка, завершение.
- Управление: play / pause / step, регулировка скорости.
- Цветовая индикация состояний и связи каналов; подписи и легенда.
- Живое демо на GitHub Pages, CI (lint + race-тесты + сборка WASM).

[0.3.0]: https://github.com/LunatikDG/goscope/releases/tag/v0.3.0
[0.1.0]: https://github.com/LunatikDG/goscope/releases/tag/v0.1.0
