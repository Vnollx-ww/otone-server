package auth

import (
	"testing"
	"time"
)

func TestNewAccessTokenRoundTrip(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Minute, time.Hour)

	token, err := manager.NewAccessToken("user-123")
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}

	userID, err := manager.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken: %v", err)
	}
	if userID != "user-123" {
		t.Fatalf("userID = %q, want %q", userID, "user-123")
	}
}

func TestAccessTokensAreUnique(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Minute, time.Hour)

	first, err := manager.NewAccessToken("user-123")
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}
	second, err := manager.NewAccessToken("user-123")
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}
	if first == second {
		t.Fatal("access tokens issued for the same user must be unique")
	}
}

func TestExpiredAccessTokenIsRejected(t *testing.T) {
	manager := NewTokenManager("test-secret", -time.Minute, time.Hour)

	token, err := manager.NewAccessToken("user-123")
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}
	if _, err := manager.ParseAccessToken(token); err == nil {
		t.Fatal("expected expired token to be rejected")
	}
}

func TestAccessTokenRejectsDifferentSecret(t *testing.T) {
	issuer := NewTokenManager("secret-a", time.Minute, time.Hour)
	verifier := NewTokenManager("secret-b", time.Minute, time.Hour)

	token, err := issuer.NewAccessToken("user-123")
	if err != nil {
		t.Fatalf("NewAccessToken: %v", err)
	}
	if _, err := verifier.ParseAccessToken(token); err == nil {
		t.Fatal("expected token signed with another secret to be rejected")
	}
}

func TestRefreshTokenHash(t *testing.T) {
	manager := NewTokenManager("test-secret", time.Minute, time.Hour)

	raw, hash, err := manager.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken: %v", err)
	}
	if raw == "" {
		t.Fatal("raw refresh token is empty")
	}
	if hash != HashRefreshToken(raw) {
		t.Fatal("hash does not match HashRefreshToken(raw)")
	}
	if len(hash) != 64 {
		t.Fatalf("hash length = %d, want 64", len(hash))
	}

	raw2, _, err := manager.NewRefreshToken()
	if err != nil {
		t.Fatalf("NewRefreshToken: %v", err)
	}
	if raw == raw2 {
		t.Fatal("refresh tokens should be unique")
	}
}
