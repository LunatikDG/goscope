package render

import "time"

// Player хранит позицию на таймлайне и по прошедшему времени
// решает, какой кадр показывать. Никакого canvas/js — чистая логика.
type Player struct {
	total     int           // число кадров
	stepEvery time.Duration // константная скорость: сколько держим один шаг
	elapsed   time.Duration // накопленное время в текущем шаге
	current   int           // индекс текущего кадра
	playing   bool
}

func NewPlayer(total int, stepEvery time.Duration) *Player {
	return &Player{total: total, stepEvery: stepEvery, playing: true}
}

// Advance добавляет прошедшее время dt и возвращает индекс кадра для отрисовки.
// Когда накопили stepEvery — переходим к следующему шагу (по кругу).
func (p *Player) Advance(dt time.Duration) int {
	if !p.playing || p.total == 0 {
		return p.current
	}
	p.elapsed += dt
	for p.elapsed >= p.stepEvery {
		p.elapsed -= p.stepEvery
		p.current = (p.current + 1) % p.total
	}
	return p.current
}

// Progress — доля прогресса внутри текущего шага [0..1), пригодится для fade.
func (p *Player) Progress() float64 {
	if p.stepEvery == 0 {
		return 0
	}
	return float64(p.elapsed) / float64(p.stepEvery)
}

// StepForward сдвигает на один кадр вперёд и встаёт на паузу.
func (p *Player) StepForward() int {
	p.playing = false
	p.elapsed = 0
	if p.total > 0 {
		p.current = (p.current + 1) % p.total
	}
	return p.current
}

// Restart возвращает на первый кадр и продолжает играть.
func (p *Player) Restart() int {
	p.current = 0
	p.elapsed = 0
	p.playing = true
	return p.current
}

// SetStepEvery меняет скорость (длительность одного шага).
func (p *Player) SetStepEvery(d time.Duration) {
	if d > 0 {
		p.stepEvery = d
	}
}

// Playing сообщает текущее состояние (пригодится для подписи кнопки).
func (p *Player) Playing() bool { return p.playing }
func (p *Player) Pause()        { p.playing = false }
func (p *Player) Play()         { p.playing = true }
func (p *Player) Current() int  { return p.current }
