package services

import (
	"crypto/rand"
	"encoding/base64"
)

func GenerateShortenURL(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)[:length]
}
