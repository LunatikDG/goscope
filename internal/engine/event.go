package engine

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
