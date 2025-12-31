package main

import (
	"sync"
	"time"
)

type LeakyBucket struct {
	capacity     float64
	queueSize    float64
	leakRate     float64
	lastLeakTime time.Time
	mu           sync.Mutex
}

func NewLeakyBucket(capacity int, leakRate float64) *LeakyBucket {
	if capacity <= 0 || leakRate <= 0 {
		panic("invalid leaky bucket params")
	}

	return &LeakyBucket{
		capacity:     float64(capacity),
		queueSize:    0,
		lastLeakTime: time.Now(),
		leakRate:     leakRate,
		mu:           sync.Mutex{},
	}

}

func (lb *LeakyBucket) Allow() bool {

	lb.mu.Lock()
	defer lb.mu.Unlock()

	currentTime := time.Now()

	elapsedTime := currentTime.Sub(lb.lastLeakTime).Seconds()

	lb.queueSize = max(0, lb.queueSize-elapsedTime*lb.leakRate)

	lb.lastLeakTime = currentTime

	if lb.queueSize+1 > lb.capacity {
		return false
	} else {
		lb.queueSize += 1
		return true
	}

}
