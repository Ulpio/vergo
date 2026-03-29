package ratelimit

import (
	"testing"
	"time"
)

func TestAllow_BurstThenReject(t *testing.T) {
	l := New(10, 3) // 10/s, burst 3
	defer l.Stop()

	// Should allow burst of 3
	for i := range 3 {
		if !l.Allow("key1") {
			t.Fatalf("request %d should be allowed within burst", i+1)
		}
	}

	// 4th should be rejected
	if l.Allow("key1") {
		t.Fatal("request 4 should be rejected after burst exhausted")
	}
}

func TestAllow_RefillAfterWait(t *testing.T) {
	l := New(100, 1) // 100/s, burst 1 — refills fast
	defer l.Stop()

	if !l.Allow("key1") {
		t.Fatal("first request should be allowed")
	}
	if l.Allow("key1") {
		t.Fatal("second request should be rejected immediately")
	}

	// Wait for refill (at 100/s, 1 token refills in 10ms)
	time.Sleep(15 * time.Millisecond)

	if !l.Allow("key1") {
		t.Fatal("request should be allowed after refill")
	}
}

func TestAllow_IndependentKeys(t *testing.T) {
	l := New(10, 1) // burst 1
	defer l.Stop()

	if !l.Allow("ip1") {
		t.Fatal("ip1 first request should be allowed")
	}
	if !l.Allow("ip2") {
		t.Fatal("ip2 first request should be allowed (independent)")
	}

	// ip1 exhausted, ip2 exhausted
	if l.Allow("ip1") {
		t.Fatal("ip1 should be rejected")
	}
	if l.Allow("ip2") {
		t.Fatal("ip2 should be rejected")
	}
}

func TestRemaining(t *testing.T) {
	l := New(10, 5)
	defer l.Stop()

	if got := l.Remaining("new-key"); got != 5 {
		t.Fatalf("remaining = %d, want 5 for new key", got)
	}

	l.Allow("new-key") // consume 1
	l.Allow("new-key") // consume 2

	got := l.Remaining("new-key")
	if got < 2 || got > 3 {
		t.Fatalf("remaining = %d, want ~3 after consuming 2 from burst 5", got)
	}
}

func TestBurst(t *testing.T) {
	l := New(10, 42)
	defer l.Stop()

	if l.Burst() != 42 {
		t.Fatalf("burst = %d, want 42", l.Burst())
	}
}
