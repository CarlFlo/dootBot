package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"gorm.io/gorm"
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
		Model:        gorm.Model{ID: user.ID},
		LastWorkedAt: time.Now().Add(time.Hour * -config.CONFIG.Work.WorkCooldown),
		Streak:       0,
		Tools:        0}

	result = DB.Create(&work)
	if result.Error != nil {
		malm.Error("Failed to create user work table: %s", result.Error)
	}
}
