package model

import (
	"fmt"
	"os"
	"time"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"

	"github.com/golang-jwt/jwt/v4"
	"github.com/spf13/viper"
	"golang.org/x/crypto/bcrypt"
)

const (
	ClientIDWeb                 = "hound-web"
	ClientIDApp                 = "hound-app"
	ClientPlatformIOSMobile     = "ios-mobile"
	ClientPlatformTVOS          = "ios-tv"
	ClientPlatformAndroidMobile = "android-mobile"
	ClientPlatformAndroidTV     = "android-tv"
	ClientPlatformWeb           = "web"
)

var SupportedClientPlatforms = []string{ClientPlatformIOSMobile, ClientPlatformTVOS, ClientPlatformAndroidMobile, ClientPlatformAndroidTV, ClientPlatformWeb}
var SupportedClientIDs = []string{ClientIDWeb, ClientIDApp}

type RegistrationUser struct {
	Username    string `json:"username" binding:"required,gt=0"`
	DisplayName string `json:"display_name" binding:"required,gt=0"`
	Password    string `json:"password" binding:"required,gte=8"`
}

type LoginUser struct {
	Username string `json:"username" binding:"required,gt=0"`
	Password string `json:"password" binding:"required,gt=0"`
	//Audience string `json:"audience" binding:"required,gt=0"`
}

type JWTClaims struct {
	UserID         int64  `json:"user_id"`
	ClientID       string `json:"client_id"`
	ClientPlatform string `json:"client_platform"`
	Role           string `json:"role"`
	jwt.RegisteredClaims
}

func RegisterNewUser(user *RegistrationUser, isAdmin bool) (*database.User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("%w: Bcrypt failed to hash password", internal.InternalServerError)
	}
	insertUser := database.User{
		Username:       user.Username,
		DisplayName:    user.DisplayName,
		IsAdmin:        isAdmin,
		HashedPassword: string(hashedPassword),
		UserMeta:       database.UserMeta{},
	}
	newID, err := database.InsertUser(insertUser)
	if err != nil || newID == nil {
		return nil, fmt.Errorf("%w: Failed to insert user to database", internal.InternalServerError)
	}
	newUser, err := database.GetUser(*newID)
	if err != nil {
		return nil, fmt.Errorf("%w: Failed to get user from database", internal.InternalServerError)
	}
	// create 'My Library' collection for user
	// userLibrary := database.CollectionRecord{
	// 	OwnerUserID:     *userID,
	// 	CollectionTitle: "My Library",
	// 	Description:     "Your main collection",
	// 	IsPublic:        false,
	// }
	// _, err = database.CreateCollection(userLibrary)
	// if err != nil {
	// 	return err
	// }
	newUser.HashedPassword = ""
	return newUser, nil
}

// GenerateAccessToken JWT access token
func GenerateAccessToken(userID int64, password string, clientID string, clientPlatform string) (string, string, error) {
	jwtKey := []byte(os.Getenv("HOUND_SECRET"))
	dbUser, err := database.GetUser(userID)
	if err != nil {
		return "", "", fmt.Errorf("Failed to fetch user from database: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.HashedPassword), []byte(password))
	if err != nil {
		return "", "", fmt.Errorf("Failed to verify password (incorrect?): %w", internal.UnauthorizedError)
	}
	// expiration time in seconds
	expirationTime := time.Now().
		Add(time.Duration(viper.GetInt("auth.jwt-access-token-expiration")) * time.Second)
	var role string
	// should change to a scope-based system in the future
	if dbUser.IsAdmin {
		role = "admin"
	} else {
		role = "user"
	}
	claims := &JWTClaims{
		UserID:         dbUser.UserID,
		ClientID:       clientID,
		ClientPlatform: clientPlatform,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
		Role: role,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Create the JWT string
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", "", fmt.Errorf("Error signing JWT token: %w", internal.InternalServerError)
	}
	return tokenString, role, nil
}

func ParseAccessToken(token string) (*JWTClaims, error) {
	jwtKey := []byte(os.Getenv("HOUND_SECRET"))
	claims := JWTClaims{}
	tkn, err := jwt.ParseWithClaims(token, &claims, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("Error decoding access token: %w: %w", err, internal.InternalServerError)
	}
	if !tkn.Valid {
		return nil, fmt.Errorf("Access token invalid or expired: %w: %w", err, internal.UnauthorizedError)
	}
	return &claims, nil
}
