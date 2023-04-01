package model

import (
	"errors"
	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
	"hound/helpers"
	"hound/model/database"
	"time"
)

type RegistrationUser struct {
	Username  string `json:"username" binding:"required,gt=0"`
	FirstName string `json:"first_name" binding:"required,gt=0"`
	LastName  string `json:"last_name" binding:"required,gt=0"`
	Password  string `json:"password" binding:"required,gte=8"`
}

type LoginUser struct {
	Username string `json:"username" binding:"required,gt=0"`
	Password string `json:"password" binding:"required,gt=0"`
	//Audience string `json:"audience" binding:"required,gt=0"`
}

type JWTClaims struct {
	Username string `json:"username"`
	Client   string `json:"client"`
	jwt.RegisteredClaims
}

func RegisterNewUser(user *RegistrationUser) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Bcrypt failed to hash password")
	}
	insertUser := database.User{
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		HashedPassword: string(hashedPassword),
		UserMeta:       database.UserMeta{},
	}
	err = database.InsertUser(insertUser)
	if err != nil {
		return helpers.LogErrorWithMessage(err, "Failed to insert user to database")
	}
	return nil
}

// GenerateAccessToken JWT access token
func GenerateAccessToken(user LoginUser, client string) (string, error) {
	jwtKey := []byte(viper.GetString("JWT_SECRET_KEY"))
	dbUser, err := database.GetUser(user.Username)
	if err != nil {
		return "", helpers.LogErrorWithMessage(err, "Failed to fetch user from database")
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.HashedPassword), []byte(user.Password))
	if err != nil {
		return "", helpers.LogErrorWithMessage(err, "Failed to verify password (incorrect?)")
	}
	// expiration time in seconds
	expirationTime := time.Now().Add(time.Duration(viper.GetInt("auth.jwt-access-token-expiration")) * time.Second)
	claims := &JWTClaims{
		Username: user.Username,
		Client:   client,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", helpers.LogErrorWithMessage(err, "Error signing JWT token")
	}
	return tokenString, nil
}

func ParseAccessToken(token string) (*JWTClaims, error) {
	jwtKey := []byte(viper.GetString("JWT_SECRET_KEY"))
	claims := JWTClaims{}
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.Unauthorized), "Error decoding access token "+err.Error())
	}
	if !tkn.Valid {
		return nil, helpers.LogErrorWithMessage(errors.New(helpers.Unauthorized), err.Error()+"Access token expired "+err.Error())
	}
	return &claims, nil
}
