package auth_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kwekkwekpatu/chirpy/internal/auth"
)

func TestValidateJWT(t *testing.T) {
	// Setup
	secret := "test_secret"
	userID := uuid.New()

	// Test 1: Valid token
	token, err := auth.MakeJWT(userID, secret, time.Hour)
	if err != nil {
		t.Fatalf("Failed to create test token: %v", err)
	}
	gotID, err := auth.ValidateJWT(token, secret)
	if err != nil {
		t.Errorf("ValidateJWT failed with valid token: %v", err)
	}
	if gotID != userID {
		t.Errorf("Got wrong user ID. Want %v, got %v", userID, gotID)
	}

	// Test 2: Wrong secret
	_, err = auth.ValidateJWT(token, "wrong_secret")
	if err == nil {
		t.Error("ValidateJWT succeeded with wrong secret")
	}

	// Test 3: Expired token
	expiredToken, _ := auth.MakeJWT(userID, secret, -time.Hour) // negative duration makes it already expired
	_, err = auth.ValidateJWT(expiredToken, secret)
	if err == nil {
		t.Error("ValidateJWT succeeded with expired token")
	}

	// Test 4: Malformed token
	_, err = auth.ValidateJWT("not.a.token", secret)
	if err == nil {
		t.Error("ValidateJWT succeeded with malformed token")
	}

	// Test 5: Empty secret
	_, err = auth.ValidateJWT(token, "")
	if err == nil {
		t.Error("ValidateJWT accepted token with empty secret when it should have failed")
	}

	// Test 6: Empty UUID
	emptyUUIDToken, _ := auth.MakeJWT(uuid.UUID{}, secret, -time.Hour)
	_, err = auth.ValidateJWT(emptyUUIDToken, secret)
	if err == nil {
		t.Error("ValidateJWT succeeded with empty UUID in token")
	}

	// Test 7: Check if token expires correctly
	// Create token that expires in 1 second
	expiringToken, _ := auth.MakeJWT(userID, secret, time.Second)
	// First validation should pass
	_, err = auth.ValidateJWT(expiringToken, secret)
	if err != nil {
		t.Error("Token should be valid initially")
	}
	// Wait 2 seconds
	time.Sleep(2 * time.Second)
	// Now validation should fail
	_, err = auth.ValidateJWT(expiringToken, secret)
	if err == nil {
		t.Error("Token should have expired")
	}
}

func TestValidateJWTFormat(t *testing.T) {
	secret := "test_secret"

	// Test different malformed tokens
	malformedTokens := []struct {
		name  string
		token string
	}{
		{
			name:  "single part token",
			token: "justonepart",
		},
		{
			name:  "two part token",
			token: "header.payload",
		},
		{
			name:  "four part token",
			token: "header.payload.signature.extra",
		},
		{
			name:  "empty token",
			token: "",
		},
		{
			name:  "token with empty parts",
			token: "...",
		},
	}

	for _, tc := range malformedTokens {
		t.Run(tc.name, func(t *testing.T) {
			_, err := auth.ValidateJWT(tc.token, secret)
			if err == nil {
				t.Errorf("ValidateJWT accepted malformed token '%s' when it should have failed", tc.token)
			}
		})
	}
}
