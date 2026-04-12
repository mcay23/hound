package database

import (
	"fmt"
	"time"

	"github.com/mcay23/hound/internal"
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
	HashedPassword string    `xorm:"text 'hashed_password'" json:"hashed_password,omitempty"`
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

func GetUser(userID int64) (*User, error) {
	cacheKey := fmt.Sprintf("user:%d", userID)
	var user User
	cacheExists, _ := GetCache(cacheKey, &user)
	if cacheExists {
		return &user, nil
	}
	found, err := databaseEngine.Table(usersTable).Where("user_id = ?", userID).Get(&user)
	if err != nil {
		return nil, fmt.Errorf("query %s for user_id %d: %w", usersTable, userID, err)
	}
	if !found {
		return nil, fmt.Errorf("query %s for user_id %d: %w", usersTable, userID, internal.NotFoundError)
	}
	SetCache(cacheKey, user, 24*time.Hour)
	return &user, nil
}

// should be used sparingly, userID is preferred internally
func GetUserIDFromUsername(username string) (int64, error) {
	var user User
	found, err := databaseEngine.Table(usersTable).Where("username = ?", username).Get(&user)
	if !found {
		return 0, fmt.Errorf("query %s for username %s: %w", usersTable, username, internal.NotFoundError)
	}
	if err != nil {
		return 0, fmt.Errorf("query %s for username %s: %w", usersTable, username, err)
	}
	return user.UserID, nil
}

func GetUsernameFromID(userID int64) (string, error) {
	var user User
	found, err := databaseEngine.Table(usersTable).ID(userID).Get(&user)
	if !found {
		return "", fmt.Errorf("query %s for user_id %d: %w", usersTable, userID, internal.NotFoundError)
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
	_, err := databaseEngine.Table(usersTable).Where("user_id = ?", userID).Delete()
	if err != nil {
		return fmt.Errorf("delete %s for user_id %d: %w", usersTable, userID, err)
	}
	return nil
}

func UpdateUserPassword(userID int64, hashedPassword string) error {
	_, err := databaseEngine.Table(usersTable).Where("user_id = ?", userID).Update(&User{HashedPassword: hashedPassword})
	if err != nil {
		return fmt.Errorf("update password for user_id %d: %w", userID, err)
	}
	cacheKey := fmt.Sprintf("user:%d", userID)
	DeleteCache(cacheKey)
	return nil
}
