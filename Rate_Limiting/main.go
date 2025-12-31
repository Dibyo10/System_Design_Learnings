package main

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

type TimeTracker struct {
	last time.Time
	mu   sync.Mutex
}

func (t *TimeTracker) log(prefix string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := time.Now()
	if !t.last.IsZero() {
		fmt.Printf("%s gap = %v\n", prefix, now.Sub(t.last))
	} else {
		fmt.Printf("%s first allowed\n", prefix)
	}
	t.last = now
}
func tokenBucketTest() {
	tb := NewTokenBucket(20, 10)

	var allowed int64
	var rejected int64
	var wg sync.WaitGroup
	tracker := &TimeTracker{}

	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			if tb.Allow(1) {
				atomic.AddInt64(&allowed, 1)
				tracker.log("[TokenBucket]")
			} else {
				atomic.AddInt64(&rejected, 1)
			}
		}()
	}

	wg.Wait()

	fmt.Println("\nTokenBucket → Initial burst")
	fmt.Println("Allowed:", allowed)
	fmt.Println("Rejected:", rejected)

	time.Sleep(1 * time.Second)

	wg.Add(11)
	for i := 0; i < 11; i++ {
		go func() {
			defer wg.Done()
			if tb.Allow(1) {
				atomic.AddInt64(&allowed, 1)
				tracker.log("[TokenBucket]")
			} else {
				atomic.AddInt64(&rejected, 1)
			}
		}()
	}

	wg.Wait()

	fmt.Println("\nTokenBucket → After 1 second")
	fmt.Println("Allowed:", allowed)
	fmt.Println("Rejected:", rejected)
}
func leakyBucketTest() {
	lb := NewLeakyBucket(20, 10)

	var allowed int64
	var rejected int64
	var wg sync.WaitGroup
	tracker := &TimeTracker{}

	wg.Add(100)

	for i := 0; i < 100; i++ {
		go func() {
			defer wg.Done()

			if lb.Allow() {
				atomic.AddInt64(&allowed, 1)
				tracker.log("[LeakyBucket]")
			} else {
				atomic.AddInt64(&rejected, 1)
			}
		}()
	}

	wg.Wait()

	fmt.Println("\nLeakyBucket → Initial burst")
	fmt.Println("Allowed:", allowed)
	fmt.Println("Rejected:", rejected)

	time.Sleep(1 * time.Second)

	wg.Add(11)
	for i := 0; i < 11; i++ {
		go func() {
			defer wg.Done()
			if lb.Allow() {
				atomic.AddInt64(&allowed, 1)
				tracker.log("[LeakyBucket]")
			} else {
				atomic.AddInt64(&rejected, 1)
			}
		}()
	}

	wg.Wait()

	fmt.Println("\nLeakyBucket → After 1 second")
	fmt.Println("Allowed:", allowed)
	fmt.Println("Rejected:", rejected)
}
func main() {
	fmt.Println("Choose test:")
	fmt.Println("1 → Token Bucket")
	fmt.Println("2 → Leaky Bucket")

	var choice int
	fmt.Scan(&choice)

	switch choice {
	case 1:
		fmt.Println("\nRunning Token Bucket Test")
		tokenBucketTest()
	case 2:
		fmt.Println("\nRunning Leaky Bucket Test")
		leakyBucketTest()
	default:
		fmt.Println("Invalid choice")
	}
}
