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
		return RegisterNewUser(&RegistrationUser{
			Username:  "admin",
			FirstName: "Admin",
			LastName:  "User",
			Password:  "password",
		}, true)
	}
	return nil
}
