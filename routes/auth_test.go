package routes

import (
	"mywetransfergo/config"
	"strings"
	"testing"
	"time"
)

func TestSessionRoundTrip(t *testing.T) {
	previous := config.App
	config.App.SessionSecret = []byte("a-session-secret-that-is-long-enough-for-testing")
	t.Cleanup(func() { config.App = previous })

	want := sessionClaims{UserID: 42, ExpiresAt: time.Now().Add(time.Hour).Unix(), CSRFToken: "csrf-token"}
	token, err := signSession(want)
	if err != nil {
		t.Fatalf("signSession returned an error: %v", err)
	}

	got, err := verifySession(token)
	if err != nil {
		t.Fatalf("verifySession returned an error: %v", err)
	}
	if got != want {
		t.Fatalf("claims mismatch: got %#v, want %#v", got, want)
	}
}

func TestSessionRejectsTamperingAndExpiry(t *testing.T) {
	previous := config.App
	config.App.SessionSecret = []byte("a-session-secret-that-is-long-enough-for-testing")
	t.Cleanup(func() { config.App = previous })

	token, err := signSession(sessionClaims{UserID: 7, ExpiresAt: time.Now().Add(time.Hour).Unix(), CSRFToken: "csrf-token"})
	if err != nil {
		t.Fatalf("signSession returned an error: %v", err)
	}
	tampered := token[:len(token)-1] + "x"
	if strings.HasSuffix(token, "x") {
		tampered = token[:len(token)-1] + "y"
	}
	if _, err := verifySession(tampered); err == nil {
		t.Fatal("tampered session token was accepted")
	}

	expired, err := signSession(sessionClaims{UserID: 7, ExpiresAt: time.Now().Add(-time.Second).Unix(), CSRFToken: "csrf-token"})
	if err != nil {
		t.Fatalf("signSession returned an error: %v", err)
	}
	if _, err := verifySession(expired); err == nil {
		t.Fatal("expired session token was accepted")
	}
}
