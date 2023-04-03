package database

import (
	"encoding/json"
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
	Id             int64
	Username       string
	FirstName      string
	LastName       string
	HashedPassword string
	UserMeta       UserMeta
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type UserXorm struct {
	Id             int64  `xorm:"pk autoincr"`
	Username       string `xorm:"not null unique"`
	FirstName      string
	LastName       string
	HashedPassword string
	UserMeta       []byte
	CreatedAt      time.Time `xorm:"created"`
	UpdatedAt      time.Time `xorm:"updated"`
}

func instantiateUsersTable() error {
	err := databaseEngine.Table(usersTable).Sync2(new(UserXorm))
	if err != nil {
		return err
	}
	return nil
}

func InsertUser(user User) error {
	userMetaBytes, err := json.Marshal(user.UserMeta)
	if err != nil {
		return err
	}
	userXorm := UserXorm{
		Username:       user.Username,
		FirstName:      user.FirstName,
		LastName:       user.LastName,
		HashedPassword: user.HashedPassword,
		UserMeta:       userMetaBytes,
	}
	_, err = databaseEngine.Table(usersTable).Insert(userXorm)
	if err != nil {
		return err
	}
	return nil
}

func GetUser(username string) (*User, error) {
	var userXorm UserXorm
	found, err := databaseEngine.Table(usersTable).Where("username = ?", username).Get(&userXorm)
	if !found {
		return nil, errors.New(helpers.BadRequest)
	}
	if err != nil {
		return nil, err
	}
	var userMeta UserMeta
	err = json.Unmarshal(userXorm.UserMeta, &userMeta)
	if err != nil {
		return nil, err
	}
	user := &User{
		Id:             userXorm.Id,
		Username:       userXorm.Username,
		FirstName:      userXorm.FirstName,
		LastName:       userXorm.LastName,
		HashedPassword: userXorm.HashedPassword,
		UserMeta:       userMeta,
		CreatedAt:      userXorm.CreatedAt,
		UpdatedAt:      userXorm.UpdatedAt,
	}
	return user, nil
}

func GetUserIDFromUsername(username string) (int64, error) {
	user, err := GetUser(username)
	if err != nil {
		return -1, helpers.LogErrorWithMessage(err, "Error retrieving user_id from username")
	}
	return user.Id, nil
}

func GetUsernameFromID(userID int64) (string, error) {
	var userXorm UserXorm
	found, err := databaseEngine.Table(usersTable).ID(userID).Get(&userXorm)
	if !found {
		return "", errors.New(helpers.BadRequest)
	}
	if err != nil {
		return "", err
	}
	return userXorm.Username, nil
}