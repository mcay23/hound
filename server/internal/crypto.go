package internal

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"io"

	"github.com/mcay23/hound/config"
)

var aesGCM cipher.AEAD

func InitializeCrypto() {
	key := getAESKey()
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(fmt.Errorf("failed to create AES cipher: %w", err))
	}
	aesGCM, err = cipher.NewGCM(block)
	if err != nil {
		panic(fmt.Errorf("failed to create AES GCM: %w", err))
	}
}

func getAESKey() []byte {
	hash := sha256.Sum256([]byte(config.HoundSecret))
	return hash[:]
}

// nonce is stored with ciphertext
func EncryptGCM(plaintext []byte) (string, error) {
	nonce := make([]byte, aesGCM.NonceSize())
	_, err := io.ReadFull(rand.Reader, nonce)
	if err != nil {
		return "", err
	}
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return base64.RawURLEncoding.EncodeToString(ciphertext), nil
}

func DecryptGCM(encrypted string) ([]byte, error) {
	data, err := base64.RawURLEncoding.DecodeString(encrypted)
	if err != nil {
		return nil, err
	}
	nonceSize := aesGCM.NonceSize()
	if len(data) < nonceSize {
		return nil, fmt.Errorf("ciphertext too short")
	}
	nonce := data[:nonceSize]
	ciphertext := data[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}
	return plaintext, nil
}
