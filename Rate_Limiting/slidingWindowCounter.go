package main

import (
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	windowSize    time.Duration
	limit         int
	currentWindow int64
	currentCount  int
	previousCount int
	mu            sync.Mutex
}

func NewSlidingWindowCounter(limit int, windowSize time.Duration) *SlidingWindowCounter {
	if limit <= 0 || windowSize <= 0 {
		panic("invalid sliding window counter params")
	}

	return &SlidingWindowCounter{
		windowSize:    windowSize,
		limit:         limit,
		currentWindow: time.Now().Unix(),
	}
}

func (sw *SlidingWindowCounter) Allow() bool {
	now := time.Now().Unix()

	sw.mu.Lock()
	defer sw.mu.Unlock()

	if now != sw.currentWindow {
		sw.previousCount = sw.currentCount
		sw.currentCount = 0
		sw.currentWindow = now
	}

	elapsed := time.Since(time.Unix(sw.currentWindow, 0)).Seconds()
	weight := (sw.windowSize.Seconds() - elapsed) / sw.windowSize.Seconds()

	effectiveCount := float64(sw.currentCount) + float64(sw.previousCount)*weight

	if int(effectiveCount) >= sw.limit {
		return false
	}

	sw.currentCount++
	return true
}
