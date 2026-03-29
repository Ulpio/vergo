// Package ratelimit provides an in-memory token bucket rate limiter.
// Each key (IP or tenant) gets an independent bucket that refills at a
// configured rate. Stale buckets are cleaned up periodically.
package ratelimit

import (
	"sync"
	"time"
)

// Limiter is a thread-safe token bucket rate limiter keyed by string.
type Limiter struct {
	rate     float64       // tokens added per second
	burst    int           // max tokens (bucket capacity)
	mu       sync.Mutex
	buckets  map[string]*bucket
	stopOnce sync.Once
	stopCh   chan struct{}
}

type bucket struct {
	tokens   float64
	lastSeen time.Time
}

// New creates a Limiter that allows `rate` requests/second with a burst of `burst`.
// It starts a background goroutine to evict stale entries every minute.
func New(rate float64, burst int) *Limiter {
	l := &Limiter{
		rate:    rate,
		burst:   burst,
		buckets: make(map[string]*bucket),
		stopCh:  make(chan struct{}),
	}
	go l.cleanup()
	return l
}

// Allow reports whether a request for the given key is allowed.
// It consumes one token if available.
func (l *Limiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	b, ok := l.buckets[key]
	if !ok {
		b = &bucket{tokens: float64(l.burst), lastSeen: now}
		l.buckets[key] = b
	}

	// Refill tokens based on elapsed time
	elapsed := now.Sub(b.lastSeen).Seconds()
	b.tokens += elapsed * l.rate
	if b.tokens > float64(l.burst) {
		b.tokens = float64(l.burst)
	}
	b.lastSeen = now

	if b.tokens < 1 {
		return false
	}
	b.tokens--
	return true
}

// Remaining returns the current token count for a key (for headers).
func (l *Limiter) Remaining(key string) int {
	l.mu.Lock()
	defer l.mu.Unlock()

	b, ok := l.buckets[key]
	if !ok {
		return l.burst
	}

	// Calculate current tokens without consuming
	elapsed := time.Since(b.lastSeen).Seconds()
	tokens := b.tokens + elapsed*l.rate
	if tokens > float64(l.burst) {
		tokens = float64(l.burst)
	}
	return int(tokens)
}

// Burst returns the max bucket capacity.
func (l *Limiter) Burst() int {
	return l.burst
}

// Stop halts the background cleanup goroutine.
func (l *Limiter) Stop() {
	l.stopOnce.Do(func() { close(l.stopCh) })
}

// cleanup evicts buckets not seen in the last 5 minutes.
func (l *Limiter) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-l.stopCh:
			return
		case <-ticker.C:
			l.mu.Lock()
			cutoff := time.Now().Add(-5 * time.Minute)
			for k, b := range l.buckets {
				if b.lastSeen.Before(cutoff) {
					delete(l.buckets, k)
				}
			}
			l.mu.Unlock()
		}
	}
}
