package database

import (
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	DiscordID string `gorm:"uniqueIndex"`
	Money     uint64
	Work      Work  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Daily     Daily `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Farm      Farm  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "users"
}

// Saves the data to the database
func (u *User) Save() {
	DB.Save(&u)
}

// Returns true if a user with that discord ID exists in the database
func (u *User) DoesUserExists(discordID string) bool {
	// Works... but rewrite this function later
	if err := DB.Where("discord_ID = ?", discordID).First(&u).Error; err != nil {
		return false
	}
	return true
}

// Queries the database for the user with the given discord ID.
// The object which calls the method will be updated with the user's data
func (u *User) GetUserByDiscordID(discordID string) {
	DB.Table("users").Where("discord_id = ?", discordID).First(&u)
}

func (u *User) PrettyPrintMoney() string {

	return utils.HumanReadableNumber(u.Money)
}
