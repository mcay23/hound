package model

import (
	"fmt"

	"github.com/mcay23/hound/database"
	"github.com/mcay23/hound/internal"

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

var SupportedClientPlatforms = []string{"", ClientPlatformIOSMobile, ClientPlatformTVOS, ClientPlatformAndroidMobile, ClientPlatformAndroidTV, ClientPlatformWeb}
var SupportedClientIDs = []string{"", ClientIDWeb, ClientIDApp}

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

func AuthenticateUser(userID int64, password string, clientID string,
	clientPlatform string, deviceID string) (string, string, string, error) {
	dbUser, err := database.GetUser(userID)
	if err != nil {
		return "", "", "", fmt.Errorf("Failed to fetch user from database: %w", err)
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.HashedPassword), []byte(password))
	if err != nil {
		return "", "", "", fmt.Errorf("Failed to verify password (incorrect?): %w", internal.UnauthorizedError)
	}
	var role string
	// should change to a scope-based system in the future
	if dbUser.IsAdmin {
		role = "admin"
	} else {
		role = "user"
	}
	sessionID, err := database.GenerateAuthSession(userID, clientID, clientPlatform, deviceID)
	if err != nil {
		return "", "", "", fmt.Errorf("Error generating auth session: %w", internal.InternalServerError)
	}
	return sessionID, dbUser.DisplayName, role, nil
}

func ParseAuthSession(sessionID string) (*database.AuthSession, error) {
	session, err := database.ValidateAuthSession(sessionID)
	if err != nil {
		return nil, fmt.Errorf("auth session validation failed: %w", internal.UnauthorizedError)
	}
	if session == nil {
		return nil, fmt.Errorf("Access token invalid or expired: %w: %w", err, internal.UnauthorizedError)
	}
	return session, nil
}

func ChangePassword(userID int64, oldPassword, newPassword string) error {
	dbUser, err := database.GetUser(userID)
	if err != nil {
		return err
	}
	err = bcrypt.CompareHashAndPassword([]byte(dbUser.HashedPassword), []byte(oldPassword))
	if err != nil {
		return fmt.Errorf("incorrect old password: %w", internal.UnauthorizedError)
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: bcrypt failed to hash password", internal.InternalServerError)
	}
	err = database.UpdateUserPassword(userID, string(hashedPassword))
	if err != nil {
		return err
	}
	return database.DeleteUserAuthSessions(userID)
}

func ResetPassword(userID int64, newPassword string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("%w: bcrypt failed to hash password", internal.InternalServerError)
	}
	err = database.UpdateUserPassword(userID, string(hashedPassword))
	if err != nil {
		return err
	}
	return database.DeleteUserAuthSessions(userID)
}
