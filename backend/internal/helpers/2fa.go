package helpers

import (
	"crypto/rand"
	"encoding/base32"
)

func GenerateRecoveryCode() (string, error) {
	b := make([]byte, 8)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	s := base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(b)

	return s[:4] + "-" + s[4:8] + "-" + s[8:], nil
}
