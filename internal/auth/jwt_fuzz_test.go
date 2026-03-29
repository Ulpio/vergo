package auth

import (
	"testing"
)

// FuzzParse feeds arbitrary strings into auth.Parse to ensure it never panics
// and always returns a clean error for invalid input.
func FuzzParse(f *testing.F) {
	// Seed corpus: valid-looking tokens, partial tokens, edge cases
	f.Add("", "secret")
	f.Add("not.a.token", "secret")
	f.Add("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1aWQiOiIxMjMiLCJ0eXAiOiJhY2Nlc3MifQ.invalid", "secret")
	f.Add("eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJ1aWQiOiIxMjMifQ.", "secret")
	f.Add("....", "secret")
	f.Add("a]]]", "")
	f.Add("\x00\xff\xfe", "secret")

	f.Fuzz(func(t *testing.T, tokenStr, secret string) {
		// Must never panic regardless of input
		claims, err := Parse(tokenStr, secret)
		if err == nil && claims == nil {
			t.Error("Parse returned nil error and nil claims")
		}
	})
}

// FuzzNewAccessToken feeds arbitrary user IDs and secrets to ensure token
// generation never panics.
func FuzzNewAccessToken(f *testing.F) {
	f.Add("user-123", "my-secret", 15)
	f.Add("", "", 0)
	f.Add("\x00\xff", "s", 1)
	f.Add("a]]]{}\"", "secret-with-special!@#$", 525600)

	f.Fuzz(func(t *testing.T, userID, secret string, ttl int) {
		if ttl < 0 || ttl > 525600 {
			t.Skip()
		}
		token, err := NewAccessToken(userID, secret, ttl)
		if err != nil {
			return // signing errors are acceptable
		}
		if token == "" {
			t.Error("NewAccessToken returned empty token without error")
		}
	})
}

// FuzzRoundTrip generates a token and parses it back, verifying that valid
// tokens always round-trip correctly.
func FuzzRoundTrip(f *testing.F) {
	f.Add("user-abc", "test-secret", 15)
	f.Add("", "s", 1)
	f.Add("uid-with-special/chars", "long-secret-key-for-hmac-256", 60)

	f.Fuzz(func(t *testing.T, userID, secret string, ttl int) {
		if ttl <= 0 || ttl > 525600 || secret == "" {
			t.Skip()
		}
		token, err := NewAccessToken(userID, secret, ttl)
		if err != nil {
			t.Fatalf("NewAccessToken failed: %v", err)
		}
		claims, err := Parse(token, secret)
		if err != nil {
			t.Fatalf("Parse failed on valid token: %v", err)
		}
		if claims.UserID != userID {
			t.Errorf("UserID mismatch: got %q, want %q", claims.UserID, userID)
		}
		if claims.TokenType != "access" {
			t.Errorf("TokenType mismatch: got %q, want %q", claims.TokenType, "access")
		}
	})
}

// FuzzParseWrongSecret verifies that parsing a valid token with the wrong
// secret always fails.
func FuzzParseWrongSecret(f *testing.F) {
	f.Add("user-1", "correct-secret", "wrong-secret")

	f.Fuzz(func(t *testing.T, userID, signSecret, parseSecret string) {
		if signSecret == parseSecret || signSecret == "" {
			t.Skip()
		}
		token, err := NewAccessToken(userID, signSecret, 15)
		if err != nil {
			t.Skip()
		}
		_, err = Parse(token, parseSecret)
		if err == nil {
			t.Error("Parse should fail with wrong secret")
		}
	})
}
