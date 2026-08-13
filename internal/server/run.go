package server

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os/exec"
	"path/filepath"
)

// handleRun запускает инструментированный пример из ExamplesDir как подпроцесс
// и построчно стримит его stdout (NDJSON, одна строка — одно событие) в
// браузер через Server-Sent Events. Deadlock-примеры естественно валят
// подпроцесс рантаймом Go — это ожидаемый исход, а не ошибка стрима.
func (s *Server) handleRun(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if !s.examples[name] {
		http.Error(w, "unknown example", http.StatusNotFound)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), s.cfg.RunTimeout)
	defer cancel()

	logger := s.logger.With(slog.String("example", name))

	examplePath := filepath.Join(s.cfg.ExamplesDir, name)
	if !filepath.IsAbs(examplePath) {
		examplePath = "./" + examplePath // без "./" go run ищет импорт-путь, а не локальную папку
	}

	//nolint:gosec // name чист: проверен по белому списку s.examples выше, не сырой ввод
	cmd := exec.CommandContext(ctx, "go", "run", examplePath)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	cmd.Stderr = &slogLineWriter{logger: s.logger, example: name} // компиляция/паника примера — в лог сервера, не клиенту

	if err := cmd.Start(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	logger.Info("example started")

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	events := 0
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		events++
		fmt.Fprintf(w, "data: %s\n\n", scanner.Text())
		flusher.Flush()
	}

	waitErr := cmd.Wait() // подпроцесс мог упасть с deadlock — это ожидаемый исход, не ошибка стрима
	logger.Info("example finished", slog.Int("events", events), slog.Any("error", waitErr))

	fmt.Fprint(w, "event: end\ndata: {}\n\n")
	flusher.Flush()

	// EventSource у браузера переподключается на любой обрыв соединения, включая
	// чистое закрытие сервером, — а нам нужен ровно один запуск. Поэтому держим
	// соединение открытым и ждём, пока клиент отключится сам (или не наступит
	// общий таймаут ctx) — тогда закрытие будет клиентским, без реконнекта.
	<-ctx.Done()
}

// slogLineWriter превращает построчно записанные в него байты в структурные
// slog-записи. Буферизует до перевода строки: subprocess.Stderr отдаёт данные
// произвольными кусками, не обязательно по границам строк.
type slogLineWriter struct {
	logger  *slog.Logger
	example string
	buf     []byte
}

func (w *slogLineWriter) Write(p []byte) (int, error) {
	w.buf = append(w.buf, p...)
	for {
		i := bytes.IndexByte(w.buf, '\n')
		if i < 0 {
			break
		}
		line := string(w.buf[:i])
		w.buf = w.buf[i+1:]
		if line != "" {
			w.logger.Warn("example stderr", slog.String("example", w.example), slog.String("line", line))
		}
	}
	return len(p), nil
}
