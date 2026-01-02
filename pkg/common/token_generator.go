package common

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

// GenerateRegistrationToken generates a secure random token for registration links
func GenerateRegistrationToken() (string, error) {
	// Generate 32 random bytes
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}

	// Encode to base64 URL-safe string
	token := base64.URLEncoding.EncodeToString(tokenBytes)
	return token, nil
}
