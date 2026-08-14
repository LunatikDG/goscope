<p align="center">
  <a href="README.md">English</a> · <b>Русский</b>
</p>

<h1 align="center">goscope</h1>

<p align="center">
  <b>Конкурентность Go — наглядно.</b><br>
  Интерактивный визуализатор горутин, каналов, блокировок и завершений.
</p>

<p align="center">
  <a href="https://lunatikdg.github.io/goscope/"><b>▶ Живое демо</b></a>
</p>


<p align="center">
  <img src="https://github.com/LunatikDG/goscope/actions/workflows/ci.yml/badge.svg" alt="CI">
  <img src="https://github.com/LunatikDG/goscope/actions/workflows/pages.yml/badge.svg" alt="Pages">
  <img src="https://img.shields.io/github/license/LunatikDG/goscope" alt="License">
  <img src="https://img.shields.io/github/go-mod/go-version/LunatikDG/goscope" alt="Go version">
</p>

---

## Что это

`goscope` показывает то, что обычно скрыто: как горутины рождаются, блокируются
на каналах, разблокируются и завершаются. Конкурентность Go тяжело представить по
коду — здесь она превращается в наглядную анимацию, которую можно проиграть,
поставить на паузу и пройти по шагам.

## Возможности

- ▶ Проигрывание / пауза / пошаговый режим, регулировка скорости.
- 🎨 Цвет = состояние горутины: running / blocked / finished.
- 🔗 Связи каналов в момент передачи.
- 🧩 Галерея паттернов: worker pool, fan-in/fan-out, pipeline, deadlock, goroutine leak.
- 🌓 Светлая/тёмная тема и пермалинки на конкретный паттерн.
- 📡 **Watch live**: запуск настоящей инструментированной Go-программы на сервере
  со стримом её реальных событий по SSE — подлинная конкурентность, не сценарий.

## Как запустить локально

```bash
git clone https://github.com/LunatikDG/goscope
cd goscope
make wasm        # собрать WebAssembly + скопировать wasm_exec.js
make serve       # поднять локальный сервер
# открой http://localhost:8080
```

Либо через Docker (в образе — сборка WASM, сервер и инструментированные примеры):

```bash
docker build -t goscope .
docker run --rm -p 8080:8080 goscope
```

## Как это устроено

Ядро (`internal/engine`) моделирует конкурентность как последовательность событий
и сворачивает их в кадры. Слой рендера (`internal/render`) превращает кадр в
команды отрисовки — без привязки к браузеру, полностью покрыт тестами. Тонкий
WebAssembly-слой (`cmd/webdemo`) исполняет эти команды на canvas.

Для живой галереи `internal/server` запускает настоящую программу из `examples/`
(инструментированную через `internal/instrument`) подпроцессом и стримит её
реальные события в браузер по Server-Sent Events — WASM-слой сворачивает их в
кадры так же, как и для готовых сценариев.

## Статус

v0.3 — галерея паттернов и живая визуализация настоящих инструментированных
программ. Дорожная карта: визуализация трейсов произвольных программ
(`runtime/trace`). Идеи и PR welcome.

## Лицензия

MIT © 2026 Дмитрий Головин
