package database

import (
	"errors"
	"hound/helpers"
	"time"
)

const usersTable = "users"

type UserMeta struct {
	Test1 string
	Test2 string
}

type User struct {
	UserID         int64  `xorm:"pk autoincr 'user_id'"`
	Username       string `xorm:"not null unique"`
	IsAdmin        bool   `xorm:"not null default false 'is_admin'"`
	FirstName      string
	LastName       string
	HashedPassword string
	UserMeta       UserMeta  `xorm:"json 'user_meta'"`
	CreatedAt      time.Time `xorm:"created"`
	UpdatedAt      time.Time `xorm:"updated"`
}

func instantiateUsersTable() error {
	err := databaseEngine.Table(usersTable).Sync2(new(User))
	if err != nil {
		return err
	}
	return nil
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
		return nil, err
	}
	if !found {
		return nil, errors.New(helpers.BadRequest)
	}
	return &user, nil
}

func GetUserIDFromUsername(username string) (int64, error) {
	user, err := GetUser(username)
	if err != nil {
		return -1, helpers.LogErrorWithMessage(err, "Error retrieving user_id from username")
	}
	return user.UserID, nil
}

func GetUsernameFromID(userID int64) (string, error) {
	var user User
	found, err := databaseEngine.Table(usersTable).ID(userID).Get(&user)
	if !found {
		return "", errors.New(helpers.BadRequest)
	}
	if err != nil {
		return "", err
	}
	return user.Username, nil
}
