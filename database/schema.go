package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	DiscordID string `gorm:"uniqueIndex"`
	Money     uint64
	Work      Work  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Daily     Daily `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

// Returns true if a user with that discord ID exists in the database
func (u *User) UserExists(discordID string) bool {
	// Works... but rewrite this function later
	if err := DB.Where("discord_ID = ?", discordID).First(&u).Error; err != nil {
		return false
	}
	return true
}

// Queries the database for the user with the given discord ID.
// The object which calls the method will be updated with the user's data
func (u *User) GetUserByDiscordID(discordID string) {
	DB.Table("Users").Where("discord_id = ?", discordID).First(&u)
}

type Work struct {
	gorm.Model
	LastWorkedAt       time.Time
	ConsecutiveStreaks uint16
	Streak             uint16
	Tools              uint8
}

// Queries the database for the work data with the given discord ID.
// The object which calls the method will be updated with the user's work data
func (w *Work) GetWorkByDiscordID(discordID string) {
	DB.Raw("select * from Works JOIN Users ON Works.ID = Users.ID WHERE Users.discord_id = ?", discordID).First(&w)
}

type Daily struct {
	gorm.Model
	LastDailyAt time.Time
	Streak      uint16
}

// Queries the database for the daily data with the given discord ID.
// The object which calls the method will be updated with the user's daily data
func (d *Daily) GetDailyByDiscordID(discordID string) {
	DB.Raw("select * from dalies JOIN Users ON dalies.ID = Users.ID WHERE Users.discord_id = ?", discordID).First(&d)
}
