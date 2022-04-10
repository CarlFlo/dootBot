package database

import (
	"github.com/CarlFlo/malm"
)

func InitializeNewUser(userID string) {

	user := User{
		DiscordID: userID,
		Money:     0}

	result := DB.Create(&user)

	if result.Error != nil {
		malm.Error("Failed to create new user in database: %s", result.Error)
	}

}
