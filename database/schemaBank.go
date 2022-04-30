package database

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
)

type Bank struct {
	Model
	Money uint64
}

func (Bank) TableName() string {
	return "userBankData"
}

// Saves the data to the database
func (b *Bank) Save() {
	DB.Save(&b)
}

func (b *Bank) PrettyPrintMoney() string {
	return utils.HumanReadableNumber(b.Money)
}

// Queries the database for the bank data with the given user object.
func (b *Bank) GetBankInfo(u *User) {
	DB.Raw("SELECT * FROM userBankData WHERE userBankData.ID = ?", u.ID).First(&b)
	if b.ID == 0 {
		b.ID = u.ID
	}
}

// Deposit - Deposits the given amount to the user's bank
// and updates the database with the new values
func (b *Bank) Deposit(u *User, depositAmount uint64) error {
	// Does the user have enought money?
	if depositAmount > u.Money {
		return fmt.Errorf("insufficient wallet funds. Your wallet has %d %s, but you are trying to deposit %d %s", u.Money, config.CONFIG.Economy.Name, depositAmount, config.CONFIG.Economy.Name)
	}

	u.Money -= depositAmount
	b.Money += depositAmount

	return nil
}

// Withdraw - Withdraws the given amount from the user's bank
// and updates the database with the new values
func (b *Bank) Withdraw(u *User, withdrawAmount uint64) error {
	// Does the bank account have enough money?

	if withdrawAmount > b.Money {
		return fmt.Errorf("insufficient bank funds. Your bank has %d %s, but you are trying to withdraw %d %s", b.Money, config.CONFIG.Economy.Name, withdrawAmount, config.CONFIG.Economy.Name)
	}

	u.Money += withdrawAmount
	b.Money -= withdrawAmount

	return nil
}

// Users can deposit money into their bank account and gain interest over time. Every 24 hours.
// Limit on how much that can be withdrawn from the bank at a time.
// Takes time before the bank transfer is completed. 1 minute for small amount, but up to several days for big amounts.
// Purpose is to not make people overuse the bank to gain interest on their money. I.e. keep running the deposit and withdraw command.
// So there is a payoff and reason to keep money in seperate locations.
// Only one bank withdrawl, per user, can be in progress at a time.
