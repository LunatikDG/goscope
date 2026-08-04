package engine

import "fmt"

// EventType — что произошло с горутиной/каналом.
type EventType int

const (
	Spawn   EventType = iota // родилась новая горутина
	Block                    // заблокировалась (ждёт канал/мьютекс)
	Unblock                  // разблокировалась
	Send                     // отправка значения в канал
	Done                     // горутина завершилась
)

func (e EventType) String() string {
	return [...]string{"spawn", "block", "unblock", "send", "done"}[e]
}

// ParseEventType — обратное отображение строка → EventType, для загрузчиков форматов.
func ParseEventType(s string) (EventType, error) {
	switch s {
	case "spawn":
		return Spawn, nil
	case "block":
		return Block, nil
	case "unblock":
		return Unblock, nil
	case "send":
		return Send, nil
	case "done":
		return Done, nil
	default:
		return 0, fmt.Errorf("неизвестный тип события %q", s)
	}
}
