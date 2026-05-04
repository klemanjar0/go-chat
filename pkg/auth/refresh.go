package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)

const refreshTokenBytes = 32

// GenerateRefreshToken returns an opaque, URL-safe random token suitable
// for handing to the client.
func GenerateRefreshToken() (string, error) {
	buf := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("read random bytes: %w", err)
	}
	return base64.RawURLEncoding.EncodeToString(buf), nil
}

// HashRefreshToken returns the hex-encoded SHA-256 of the token. We never
// store the raw token, only its hash, so a database leak does not expose
// usable credentials.
func HashRefreshToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
