package providers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
)

func getAESKey(salt []byte) (*[]byte, error) {
	key, err := pbkdf2.Key(sha256.New, os.Getenv("HOUND_SECRET"), salt, 4096, 16)
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
	key, err := getAESKey([]byte(fixedSalt))
	if err != nil {
		return "", err
	}
	bytes, err := json.Marshal(streamObject)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(*key)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	ciphertext := gcm.Seal([]byte(fixedNonce), []byte(fixedNonce), bytes, nil)
	final := append([]byte(fixedSalt), ciphertext...)

	return base64.URLEncoding.EncodeToString(final), nil
}

/*
Decode a string back into StreamObjectFull data
*/
func DecodeJsonStreamAES(encryptedText string) (*StreamObjectFull, error) {
	fullCiphertext, err := base64.URLEncoding.DecodeString(encryptedText)
	if err != nil {
		return nil, err
	}
	saltLen := len(fixedSalt)
	if len(fullCiphertext) < saltLen {
		return nil, io.ErrUnexpectedEOF
	}
	salt := fullCiphertext[:saltLen]
	ciphertext := fullCiphertext[saltLen:]

	key, err := getAESKey(salt)
	if err != nil {
		return nil, err
	}
	block, err := aes.NewCipher(*key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, io.ErrUnexpectedEOF
	}
	nonce, actualCiphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	bytes, err := gcm.Open(nil, nonce, actualCiphertext, nil)
	if err != nil {
		return nil, err
	}
	var streamObject StreamObjectFull
	if err := json.Unmarshal(bytes, &streamObject); err != nil {
		return nil, err
	}
	return &streamObject, nil
}
