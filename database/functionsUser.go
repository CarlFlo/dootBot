package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
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

	work := Work{
		UserID:      user.ID,
		LastUpdated: time.Now().Add(time.Hour * -config.CONFIG.Work.WorkCooldown),
		Streak:      0,
		Tools:       0}

	result = DB.Create(&work)
	if result.Error != nil {
		malm.Error("Failed to create user work table: %s", result.Error)
	}
}
