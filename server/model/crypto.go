package model

import (
	"encoding/json"
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"os"
)

// to encode into JWT string
type StreamObjectFull struct {
	StreamMediaDetails
	StreamObject
}

/*
	Encodes a stream into a string using JWT
 */
func EncodeJsonStreamJWT(streamObject StreamObjectFull) (string, error) {
	var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	// Marshal the struct into JSON
	bytes, _ := json.Marshal(streamObject)

	// Create a token with the JSON stored under one claim (e.g. "data")
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"data": string(bytes),
	})
	return token.SignedString(jwtKey)
}

/*
	Decode a string back into StreamObjectFull data
 */
func DecodeJsonStreamJWT(tokenString string) (*StreamObjectFull, error){
	var jwtKey = []byte(os.Getenv("JWT_SECRET_KEY"))
	parsed, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := parsed.Claims.(jwt.MapClaims); ok && parsed.Valid {
		var decoded StreamObjectFull
		err := json.Unmarshal([]byte(claims["data"].(string)), &decoded)
		if err != nil {
			return nil, err
		}
		return &decoded, nil
	}
	return nil, errors.New("DecodeJsonStreamJWT(): could not parse JWT claims")
}