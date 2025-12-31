package main

import (
	"sync"

	"time"
)

type TokenBucket struct {
	capacity        float64
	tokensAvailable float64
	refillRate      float64
	lastRefillTime  time.Time
	mu              sync.Mutex
}

func NewTokenBucket(capacity int, refill_rate int) *TokenBucket {

	if capacity <= 0 {
		panic("Capacity of the bucket must be a positive number")
	}
	if refill_rate <= 0 {
		panic("Refill rate must be a positive number")
	}

	return &TokenBucket{
		capacity:        float64(capacity),
		tokensAvailable: float64(capacity),
		refillRate:      float64(refill_rate),
		lastRefillTime:  time.Now(),
		mu:              sync.Mutex{},
	}

}

func (tb *TokenBucket) Allow(n int) bool {

	if n <= 0 {
		return true
	}

	currentTime := time.Now()

	tb.mu.Lock()
	defer tb.mu.Unlock()

	elapsedTime := currentTime.Sub(tb.lastRefillTime).Seconds()

	tb.tokensAvailable = min(tb.tokensAvailable+(elapsedTime*tb.refillRate), tb.capacity)

	tb.lastRefillTime = currentTime

	if tb.tokensAvailable >= float64(n) {
		tb.tokensAvailable -= float64(n)

		return true

	} else {

		return false
	}

}
