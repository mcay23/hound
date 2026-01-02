package providers

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/pbkdf2"
	"crypto/rand"
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

/*
Encodes a stream into a string using AES
This also protects api keys in urls from being
exposed if hound link is shared
*/
func EncodeJsonStreamAES(streamObject StreamObjectFull) (string, error) {
	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}
	key, err := getAESKey(salt)
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
	nonce := make([]byte, gcm.NonceSize())
	io.ReadFull(rand.Reader, nonce)
	ciphertext := gcm.Seal(nonce, nonce, bytes, nil)
	final := append(salt, ciphertext...)

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
	if len(fullCiphertext) < 16 {
		return nil, io.ErrUnexpectedEOF
	}
	salt := fullCiphertext[:16]
	ciphertext := fullCiphertext[16:]

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
