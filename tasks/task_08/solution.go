package main

import (
	"math"
	"sync"
	"time"
)

type Clock interface{ Now() time.Time }

type Limiter struct {
	mu     sync.Mutex
	clock  Clock
	rate   float64
	burst  int
	tokens float64
	last   time.Time
}

func NewLimiter(clock Clock, ratePerSec float64, burst int) *Limiter {
	var start time.Time
	if clock != nil {
		start = clock.Now()
	}
	return &Limiter{
		clock:  clock,
		rate:   ratePerSec,
		burst:  burst,
		tokens: float64(max(burst, 0)),
		last:   start,
	}
}

func (l *Limiter) AllowN(n int) bool {
	if l.clock == nil || n <= 0 || n > l.burst {
		return false
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.clock.Now()
	if now.After(l.last) {
		if l.rate > 0 {
			gained := now.Sub(l.last).Seconds() * l.rate
			l.tokens = math.Min(l.tokens+gained, float64(l.burst))
		}
		l.last = now
	}

	want := float64(n)
	if want > l.tokens {
		return false
	}
	l.tokens -= want
	return true
}
