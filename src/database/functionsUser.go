package database

import (
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/malm"
)

func InitializeNewUser(userID string) {

	user := User{
		DiscordID: userID,
		Money:     config.CONFIG.Economy.StartingMoney}

	result := DB.Create(&user)

	if result.Error != nil {
		malm.Error("Failed to create new user in database: %s", result.Error)
	}

}
