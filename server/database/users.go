package database

import (
	"fmt"
	"hound/helpers"
	"time"
)

const usersTable = "users"

type UserMeta struct {
	Test1 string
	Test2 string
}

type User struct {
	UserID         int64     `xorm:"pk autoincr 'user_id'"`
	Username       string    `xorm:"not null unique"`
	IsAdmin        bool      `xorm:"not null default false 'is_admin'"`
	FirstName      string    `xorm:"'first_name'"`
	LastName       string    `xorm:"'last_name'"`
	HashedPassword string    `xorm:"text 'hashed_password'"`
	UserMeta       UserMeta  `xorm:"json 'user_meta'"`
	CreatedAt      time.Time `xorm:"timestampz created"`
	UpdatedAt      time.Time `xorm:"timestampz updated"`
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
	err := databaseEngine.Table(usersTable).Find(&users)
	if err != nil {
		return nil, fmt.Errorf("query %s: %w", usersTable, err)
	}
	return users, nil
}
