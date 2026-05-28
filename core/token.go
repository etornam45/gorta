package core

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
)

func GenerateID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// rand.Read only fails if the OS entropy pool is broken.
		// This is unrecoverable — panic is appropriate here.
		panic(fmt.Sprintf("gorta: failed to generate random ID: %v", err))
	}
	return fmt.Sprintf("%x", b)
}

func GenerateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("gorta: generating token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}
