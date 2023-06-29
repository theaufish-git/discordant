package period

import (
	"math/rand"
	"time"
)

type Period interface {
	Period() time.Duration
}

type Random struct {
	min     time.Duration
	max     time.Duration
	window  time.Duration
	seconds int
}

func NewRandom(min time.Duration, max time.Duration) *Random {
	window := max - min
	return &Random{
		min:     min,
		max:     max,
		window:  window,
		seconds: int(window.Seconds()),
	}
}

func (r *Random) Period() time.Duration {
	dur := time.Duration(rand.Intn(r.seconds)) * time.Second
	return dur + r.min
}
