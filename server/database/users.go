package database

import (
	"fmt"
	"github.com/mcay23/hound/helpers"
	"time"
)

const usersTable = "users"

type UserMeta struct {
	Test1 string
	Test2 string
}

type User struct {
	UserID         int64     `xorm:"pk autoincr 'user_id'" json:"user_id"`
	Username       string    `xorm:"not null unique" json:"username"`
	IsAdmin        bool      `xorm:"not null default false 'is_admin'" json:"is_admin"`
	DisplayName    string    `xorm:"'display_name'" json:"display_name"`
	HashedPassword string    `xorm:"text 'hashed_password'" json:"-"`
	UserMeta       UserMeta  `xorm:"json 'user_meta'" json:"user_meta"`
	CreatedAt      time.Time `xorm:"timestampz created" json:"created_at"`
	UpdatedAt      time.Time `xorm:"timestampz updated" json:"updated_at"`
}

func instantiateUsersTable() error {
	err := databaseEngine.Table(usersTable).Sync2(new(User))
	// if no user exists, insert a default admin user
	return err
}

func InsertUser(user User) (*int64, error) {
	_, err := databaseEngine.Table(usersTable).Insert(&user)
	if err != nil {
		return nil, err
	}
	return &user.UserID, nil
}

func GetUser(username string) (*User, error) {
	var user User
	found, err := databaseEngine.Table(usersTable).Where("username = ?", username).Get(&user)
	if err != nil {
		return nil, fmt.Errorf("query %s for username %s: %w", usersTable, username, err)
	}
	if !found {
		return nil, fmt.Errorf("query %s for username %s: %w", usersTable, username, helpers.NotFoundError)
	}
	return &user, nil
}

func GetUserIDFromUsername(username string) (int64, error) {
	cacheKey := fmt.Sprintf("user_id_mapping:%s", username)
	var userID int64
	cacheExists, _ := GetCache(cacheKey, &userID)
	if cacheExists {
		return userID, nil
	}
	user, err := GetUser(username)
	if err != nil {
		return -1, fmt.Errorf("query %s for username %s: %w", usersTable, username, err)
	}
	SetCache(cacheKey, user.UserID, 48*time.Hour)
	return user.UserID, nil
}

func GetUsernameFromID(userID int64) (string, error) {
	var user User
	found, err := databaseEngine.Table(usersTable).ID(userID).Get(&user)
	if !found {
		return "", fmt.Errorf("query %s for user_id %d: %w", usersTable, userID, helpers.NotFoundError)
	}
	if err != nil {
		return "", fmt.Errorf("query %s for user_id %d: %w", usersTable, userID, err)
	}
	return user.Username, nil
}

func GetUsers() ([]User, error) {
	var users []User
	err := databaseEngine.Table(usersTable).OrderBy("user_id asc").Find(&users)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", usersTable, err)
	}
	return users, nil
}

func DeleteUser(userID int64) error {
	username, err := GetUsernameFromID(userID)
	if err == nil {
		cacheKey := fmt.Sprintf("user_id_mapping:%s", username)
		_ = DeleteCache(cacheKey)
	}
	_, err = databaseEngine.Table(usersTable).Where("user_id = ?", userID).Delete()
	if err != nil {
		return fmt.Errorf("delete %s for user_id %d: %w", usersTable, userID, err)
	}
	return nil
}
