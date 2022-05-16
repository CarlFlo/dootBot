package database

import (
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"gorm.io/gorm"
)

type User struct {
	Model
	DiscordID        string `gorm:"uniqueIndex"`
	Money            uint64
	LifetimeEarnings uint64
	Work             Work  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Daily            Daily `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Farm             Farm  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	// Log in debug DB maybe
	return nil
}

// Saves the data to the database
func (u *User) Save() {
	DB.Save(&u)
}

// Returns true if a user with that discord ID exists in the database
func (u *User) DoesUserExist(discordID string) bool {

	var count int
	DB.Raw("SELECT COUNT(*) FROM users WHERE discord_id = ?", discordID).Scan(&count)

	return count == 1
}

// Queries the database for the user with the given discord ID.
// The object which calls the method will be updated with the user's data
func (u *User) QueryUserByDiscordID(discordID string) {
	DB.Table("users").Where("discord_id = ?", discordID).First(&u)
}

func (u *User) PrettyPrintMoney() string {
	return utils.HumanReadableNumber(u.Money)
}

func (u *User) PrettyPrintLifetimeEarnings() string {
	return utils.HumanReadableNumber(u.LifetimeEarnings)
}

func (u *User) AddMoney(amount uint64) {
	u.Money += amount
	u.LifetimeEarnings += amount
}

func (u *User) DeductMoney(amount uint64) {
	u.Money -= amount
}

func (u *User) CanAfford(number uint64) bool {
	return u.Money >= number
}
