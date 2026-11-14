package helper

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

func NewQRToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate qr token: %w", err)
	}
	return hex.EncodeToString(b), nil
}
