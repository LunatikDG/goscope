package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"time"
)

// examples — белый список того, что можно запустить через /api/run/{name}:
// имя приходит из URL, и мы не хотим передавать его в exec.Command без проверки.
var examples = map[string]bool{
	"workerpool":     true,
	"fanin_fanout":   true,
	"pipeline":       true,
	"deadlock":       true,
	"goroutine_leak": true,
}

// runTimeout — жёсткий потолок на один запуск примера: подчищает то, что
// само не завершается (например, если бы пример забыл ограничить leak сном).
const runTimeout = 15 * time.Second

// handleRun запускает инструментированный пример из examples/ как подпроцесс
// и построчно стримит его stdout (NDJSON, одна строка — одно событие) в
// браузер через Server-Sent Events. Deadlock-примеры естественно валят
// подпроцесс рантаймом Go — это ожидаемый исход, а не ошибка стрима.
func handleRun(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !examples[name] {
		http.Error(w, "unknown example", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), runTimeout)
	defer cancel()

	// r.Context() отменяется и при разрыве соединения клиентом — WithTimeout
	// поверх него убивает подпроцесс и в этом случае, не только по таймауту.
	//nolint:gosec // name чист: проверен по белому списку examples выше, не сырой ввод
	cmd := exec.CommandContext(ctx, "go", "run", "./examples/"+name)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cmd.Stderr = log.Writer() // ошибки компиляции/паники видно в консоли сервера, не у клиента

	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		fmt.Fprintf(w, "data: %s\n\n", scanner.Text())
		flusher.Flush()
	}

	_ = cmd.Wait() // подпроцесс мог упасть с deadlock — это ожидаемый исход, не ошибка стрима

	fmt.Fprint(w, "event: end\ndata: {}\n\n")
	flusher.Flush()

	// EventSource у браузера переподключается на любой обрыв соединения, включая
	// чистое закрытие сервером, — а нам нужен ровно один запуск. Поэтому держим
	// соединение открытым и ждём, пока клиент отключится сам (или не наступит
	// общий таймаут ctx) — тогда закрытие будет клиентским, без реконнекта.
	<-ctx.Done()
}
