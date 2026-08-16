package helpers

import (
	"encoding/hex"
	"fmt"
	"os"

	"github.com/gtank/cryptopasta"
)

func LoadAppEncKey() ([32]byte, error) {
	// Load encryption key
	encKeyStr := os.Getenv("APP_ENCRYPTION_KEY")
	if encKeyStr == "" {
		return [32]byte{}, fmt.Errorf("APP_ENCRYPTION_KEY environment variable not set")
	}

	encKey, err := hex.DecodeString(encKeyStr)
	if err != nil {
		return [32]byte{}, fmt.Errorf("APP_ENCRYPTION_KEY is not valid hex")
	}

	if len(encKey) != 32 {
		return [32]byte{}, fmt.Errorf("APP_ENCRYPTION_KEY must decode to exactly 32 bytes")
	}

	var key32 [32]byte
	copy(key32[:], encKey)

	return key32, nil
}

func EncryptString(s string, key [32]byte) ([]byte, error) {
	encrypted, err := cryptopasta.Encrypt([]byte(s), &key)
	if err != nil {
		return nil, err
	}

	return encrypted, nil
}

func DecryptString(ciphertext []byte, key [32]byte) (string, error) {
	decrypted, err := cryptopasta.Decrypt(ciphertext, &key)
	if err != nil {
		return "", err
	}

	return string(decrypted), nil
}
