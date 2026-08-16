package helpers

import "crypto/sha256"

// MakeHash256 turns the given string into a SHA256 hash.
func MakeHash256(s string) []byte {
	sum := sha256.Sum256([]byte(s))
	return sum[:]
}
