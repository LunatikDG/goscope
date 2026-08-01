package render

import (
	"testing"
	"time"
)

func TestPlayerConstantRate(t *testing.T) {
	p := NewPlayer(4, 100*time.Millisecond)

	// меньше шага — кадр не меняется
	if got := p.Advance(50 * time.Millisecond); got != 0 {
		t.Fatalf("через 50мс кадр = %d, ожидался 0", got)
	}
	// перевалили за 100мс — перешли на кадр 1
	if got := p.Advance(60 * time.Millisecond); got != 1 {
		t.Fatalf("через 110мс кадр = %d, ожидался 1", got)
	}
	// большой скачок времени пролистывает несколько шагов и зацикливает
	if got := p.Advance(250 * time.Millisecond); got != 3 {
		t.Fatalf("после скачка кадр = %d, ожидался 3", got)
	}
}

func TestPlayerPause(t *testing.T) {
	p := NewPlayer(4, 100*time.Millisecond)
	p.Pause()
	if got := p.Advance(500 * time.Millisecond); got != 0 {
		t.Fatalf("на паузе кадр не должен меняться, стало %d", got)
	}
}

func TestPlayerStepForwardPauses(t *testing.T) {
	p := NewPlayer(3, 100*time.Millisecond)
	if got := p.StepForward(); got != 1 {
		t.Fatalf("step → кадр %d, ожидался 1", got)
	}
	if p.Playing() {
		t.Fatal("step должен ставить на паузу")
	}
}

func TestPlayerRestart(t *testing.T) {
	p := NewPlayer(3, 100*time.Millisecond)
	p.Advance(250 * time.Millisecond) // уехали вперёд
	if got := p.Restart(); got != 0 || !p.Playing() {
		t.Fatalf("restart → кадр %d playing=%v, ожидалось 0/true", got, p.Playing())
	}
}
