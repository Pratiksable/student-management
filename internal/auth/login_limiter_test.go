package auth

import (
	"testing"
	"time"
)

func TestLoginLimiter(t *testing.T) {
	limiter := NewLoginLimiter(2, time.Minute)
	if !limiter.Allow("user|ip") {
		t.Fatal("new login key was blocked")
	}
	limiter.RecordFailure("user|ip")
	limiter.RecordFailure("user|ip")
	if limiter.Allow("user|ip") {
		t.Fatal("login key was not blocked at the failure limit")
	}
	limiter.Reset("user|ip")
	if !limiter.Allow("user|ip") {
		t.Fatal("successful login did not reset failures")
	}
}

func TestLoginLimiterWindowExpires(t *testing.T) {
	limiter := NewLoginLimiter(1, time.Minute)
	currentTime := time.Now()
	limiter.now = func() time.Time { return currentTime }
	limiter.RecordFailure("user|ip")
	if limiter.Allow("user|ip") {
		t.Fatal("login key was not blocked")
	}
	currentTime = currentTime.Add(2 * time.Minute)
	if !limiter.Allow("user|ip") {
		t.Fatal("login key remained blocked after its window expired")
	}
}
