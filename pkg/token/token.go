package token

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

const (
	Prefix   = "rp_"
	ByteSize = 32
)

func Generate() (string, error) {
	b := make([]byte, ByteSize)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return Prefix + hex.EncodeToString(b), nil
}
