package model

import (
	"hound/database"
)

func InitializeOnboarding() {
	err := initializeDefaultUser()
	if err != nil {
		panic(err)
	}
}

func initializeDefaultUser() error {
	users, err := database.GetUsers()
	if err != nil {
		return err
	}
	if len(users) == 0 {
		// create a new admin user
		_, err := RegisterNewUser(&RegistrationUser{
			Username:    "admin",
			DisplayName: "Admin",
			Password:    "password",
		}, true)
		if err != nil {
			return err
		}
	}
	return nil
}
