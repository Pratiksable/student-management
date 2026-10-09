package auth

import (
	"sync"
	"time"
)

type loginAttempt struct {
	count       int
	windowStart time.Time
}

type LoginLimiter struct {
	mu       sync.Mutex
	attempts map[string]loginAttempt
	limit    int
	window   time.Duration
	now      func() time.Time
}

func NewLoginLimiter(limit int, window time.Duration) *LoginLimiter {
	return &LoginLimiter{
		attempts: make(map[string]loginAttempt),
		limit:    limit,
		window:   window,
		now:      time.Now,
	}
}

func (l *LoginLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	attempt, found := l.attempts[key]
	if !found {
		return true
	}
	if l.now().Sub(attempt.windowStart) >= l.window {
		delete(l.attempts, key)
		return true
	}
	return attempt.count < l.limit
}

func (l *LoginLimiter) RecordFailure(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	for existingKey, existingAttempt := range l.attempts {
		if now.Sub(existingAttempt.windowStart) >= l.window {
			delete(l.attempts, existingKey)
		}
	}
	attempt, found := l.attempts[key]
	if !found || now.Sub(attempt.windowStart) >= l.window {
		l.attempts[key] = loginAttempt{count: 1, windowStart: now}
		return
	}
	attempt.count++
	l.attempts[key] = attempt
}

func (l *LoginLimiter) Reset(key string) {
	l.mu.Lock()
	delete(l.attempts, key)
	l.mu.Unlock()
}

var DefaultLoginLimiter = NewLoginLimiter(5, 15*time.Minute)
