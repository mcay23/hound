package providers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"

	"github.com/mcay23/hound/config"
)

func getAESKey(salt []byte) (*[]byte, error) {
	key, err := pbkdf2.Key(sha256.New, config.HoundSecret, salt, 4096, 16)
	return &key, err
}

var fixedSalt = "my-random-hound-salt-123"
var fixedNonce = "hound*nonce*"

/*
Encodes a stream into a string using AES
This also protects api keys in urls from being
exposed if hound link is shared

For our purposes, we want a stable hash, so we use a fixed salt and nonce
*/
func EncodeJsonStreamAES(streamObject StreamObjectFull) (string, error) {
	bytes, err := json.Marshal(streamObject)
	if err != nil {
		return "", fmt.Errorf("error marshaling stream object: %w", err)
	}
	return encrypt(bytes)
}

/*
Decode a string back into StreamObjectFull data
*/
func DecodeJsonStreamAES(encryptedText string) (*StreamObjectFull, error) {
	bytes, err := decrypt(encryptedText)
	if err != nil {
		return nil, err
	}
	var streamObject StreamObjectFull
	if err := json.Unmarshal(bytes, &streamObject); err != nil {
		return nil, fmt.Errorf("error unmarshaling stream object: %w", err)
	}
	return &streamObject, nil
}

func EncodeURIAES(uri string) (string, error) {
	return encrypt([]byte(uri))
}

func DecodeURIAES(encryptedText string) (string, error) {
	bytes, err := decrypt(encryptedText)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func encrypt(plaintext []byte) (string, error) {
	key, err := getAESKey([]byte(fixedSalt))
	if err != nil {
		return "", fmt.Errorf("error getting AES key: %w", err)
	}
	block, err := aes.NewCipher(*key)
	if err != nil {
		return "", fmt.Errorf("error creating AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", fmt.Errorf("error creating GCM: %w", err)
	}
	ciphertext := gcm.Seal([]byte(fixedNonce), []byte(fixedNonce), plaintext, nil)
	final := append([]byte(fixedSalt), ciphertext...)

	return base64.URLEncoding.EncodeToString(final), nil
}

func decrypt(encryptedText string) ([]byte, error) {
	fullCiphertext, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		return nil, fmt.Errorf("error decoding base64: %w", err)
	}
	saltLen := len(fixedSalt)
	if len(fullCiphertext) < saltLen {
		return nil, io.ErrUnexpectedEOF
	}
	salt := fullCiphertext[:saltLen]
	ciphertext := fullCiphertext[saltLen:]

	key, err := getAESKey(salt)
	if err != nil {
		return nil, fmt.Errorf("error getting AES key: %w", err)
	}
	block, err := aes.NewCipher(*key)
	if err != nil {
		return nil, fmt.Errorf("error creating AES cipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("error creating GCM: %w", err)
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, io.ErrUnexpectedEOF
	}
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	bytes, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, fmt.Errorf("error opening GCM: %w", err)
	}
	return bytes, nil
}
