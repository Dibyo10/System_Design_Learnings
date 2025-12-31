package main

import (
	"sync"
	"time"
)

type SlidingWindowLog struct {
	windowSize time.Duration
	limit      int
	timestamps []time.Time
	mu         sync.Mutex
}

func NewSlidingWindowLog(limit int, windowSize time.Duration) *SlidingWindowLog {
	if limit <= 0 || windowSize <= 0 {
		panic("invalid sliding window log params")
	}

	return &SlidingWindowLog{
		windowSize: windowSize,
		limit:      limit,
		timestamps: make([]time.Time, 0),
	}
}

func (sw *SlidingWindowLog) Allow() bool {
	now := time.Now()

	sw.mu.Lock()
	defer sw.mu.Unlock()

	cutoff := now.Add(-sw.windowSize)

	idx := 0
	for idx < len(sw.timestamps) && sw.timestamps[idx].Before(cutoff) {
		idx++
	}
	sw.timestamps = sw.timestamps[idx:]

	if len(sw.timestamps) >= sw.limit {
		return false
	}

	sw.timestamps = append(sw.timestamps, now)
	return true
}
