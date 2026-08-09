// Package instrument даёт инструментированным примерам в examples/ единый способ
// сообщить о происходящем: одна строка NDJSON на stdout на каждое событие.
// Сервер (cmd/serve) запускает пример как подпроцесс, построчно читает stdout
// и стримит эти же строки в браузер по SSE.
package instrument

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// Event — одна строка на stdout; схема совпадает с engine.Step по смыслу полей.
type Event struct {
	Event     string `json:"event"`
	Label     string `json:"label,omitempty"`
	Goroutine int    `json:"goroutine"`
	Chan      int    `json:"chan,omitempty"`
}

// mu защищает stdout: несколько горутин примера пишут события одновременно,
// без мьютекса строки могли бы перемежаться и ломать построчный NDJSON.
var mu sync.Mutex

func emit(e Event) {
	data, err := json.Marshal(e)
	if err != nil {
		return // Event сериализуется всегда; сюда не попадём
	}

	mu.Lock()
	defer mu.Unlock()
	fmt.Fprintln(os.Stdout, string(data))
}

// Spawn сообщает о рождении горутины goroutine с подписью label (может быть пустой).
func Spawn(goroutine int, label string) {
	emit(Event{Event: "spawn", Goroutine: goroutine, Label: label})
}

// Block сообщает, что горутина заблокировалась на канале ch.
func Block(goroutine, ch int) {
	emit(Event{Event: "block", Goroutine: goroutine, Chan: ch})
}

// Unblock сообщает, что горутина разблокировалась (получила значение из ch).
func Unblock(goroutine, ch int) {
	emit(Event{Event: "unblock", Goroutine: goroutine, Chan: ch})
}

// Send сообщает об отправке значения в канал ch.
func Send(goroutine, ch int) {
	emit(Event{Event: "send", Goroutine: goroutine, Chan: ch})
}

// Done сообщает о завершении горутины.
func Done(goroutine int) {
	emit(Event{Event: "done", Goroutine: goroutine})
}
