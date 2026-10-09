package auth

import (
	"testing"
	"time"
)

func TestSessionLifecycle(t *testing.T) {
	store := NewSessionStore(time.Hour)
	token, expiresAt, err := store.Create(42)
	if err != nil {
		t.Fatal(err)
	}
	if token == "" || !expiresAt.After(time.Now()) {
		t.Fatal("session token or expiration was not created")
	}

	session, err := store.Get(token)
	if err != nil || session.ExecID != 42 {
		t.Fatalf("could not read session: session=%+v err=%v", session, err)
	}

	store.Delete(token)
	if _, err := store.Get(token); err == nil {
		t.Fatal("deleted session remained valid")
	}
}

func TestExpiredSessionIsRejected(t *testing.T) {
	store := NewSessionStore(time.Minute)
	currentTime := time.Now()
	store.now = func() time.Time { return currentTime }
	token, _, err := store.Create(7)
	if err != nil {
		t.Fatal(err)
	}

	currentTime = currentTime.Add(2 * time.Minute)
	if _, err := store.Get(token); err == nil {
		t.Fatal("expired session was accepted")
	}
}

func TestSessionTokensAreUnique(t *testing.T) {
	store := NewSessionStore(time.Hour)
	first, _, err := store.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	second, _, err := store.Create(1)
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("session tokens unexpectedly matched")
	}
}
