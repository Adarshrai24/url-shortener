package utils

import (
	"crypto/sha256"
	"encoding/base64"
)

func Hash(url string) string {
	sum := sha256.Sum256([]byte(url))
	encoded := base64.RawURLEncoding.EncodeToString(sum[:])
	return encoded[:7]
}
