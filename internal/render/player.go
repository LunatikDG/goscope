package render

import "time"

// Player holds the position on the timeline and, from elapsed time,
// decides which frame to show. No canvas/js — pure logic.
type Player struct {
	total     int           // number of frames
	stepEvery time.Duration // constant rate: how long we hold each step
	elapsed   time.Duration // time accumulated within the current step
	current   int           // index of the current frame
	playing   bool
}

func NewPlayer(total int, stepEvery time.Duration) *Player {
	return &Player{total: total, stepEvery: stepEvery, playing: true}
}

// Advance adds the elapsed time dt and returns the frame index to draw.
// Once stepEvery has accumulated, we move to the next step (wrapping around).
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

// Progress — fraction of progress within the current step [0..1), useful for a fade.
func (p *Player) Progress() float64 {
	if p.stepEvery == 0 {
		return 0
	}
	return float64(p.elapsed) / float64(p.stepEvery)
}

// StepForward advances one frame and pauses.
func (p *Player) StepForward() int {
	p.playing = false
	p.elapsed = 0
	if p.total > 0 {
		p.current = (p.current + 1) % p.total
	}
	return p.current
}

// Restart goes back to the first frame and resumes playing.
func (p *Player) Restart() int {
	p.current = 0
	p.elapsed = 0
	p.playing = true
	return p.current
}

// SetStepEvery changes the rate (the duration of one step).
func (p *Player) SetStepEvery(d time.Duration) {
	if d > 0 {
		p.stepEvery = d
	}
}

// Playing reports the current state (useful for a button label).
func (p *Player) Playing() bool { return p.playing }
func (p *Player) Pause()        { p.playing = false }
func (p *Player) Play()         { p.playing = true }
func (p *Player) Current() int  { return p.current }
