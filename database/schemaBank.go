package database

import (
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"gorm.io/gorm"
)

type Bank struct {
	gorm.Model
	Money uint64
}

func (Bank) TableName() string {
	return "userBankData"
}

func (b *Bank) PrettyPrintMoney() string {

	return utils.HumanReadableNumber(b.Money)
}

// Queries the database for the bank data with the given user object.
func (b *Bank) GetBankInfo(user *User) {
	DB.Raw("SELECT * FROM userBankData WHERE userBankData.ID = ?", user.ID).First(&b)
}

// Users can deposit money into their bank account and gain interest over time. Every 24 hours.
// Limit on how much that can be withdrawn from the bank at a time.
// Takes time before the bank transfer is completed. 1 minute for small amount, but up to several days for big amounts.
// Purpose is to not make people overuse the bank to gain interest on their money. I.e. keep running the deposit and withdraw command.
// So there is a payoff and reason to keep money in seperate locations.
// Only one bank withdrawl, per user, can be in progress at a time.
