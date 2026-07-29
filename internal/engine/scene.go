package engine

// Step — одно событие на таймлайне сцены (что и с кем).
type Step struct {
	Label     string // подпись, напр. "worker-1"
	Event     EventType
	Goroutine int // id горутины — «актёра» события
	Chan      int // id канала (для Send/Block по каналу); 0 если неприменимо
}

// Scene — сценарий: имя + упорядоченные шаги.
type Scene struct {
	Name  string
	Steps []Step
}
