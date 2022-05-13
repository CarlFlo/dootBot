package database

import (
	"fmt"
	"math"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"gorm.io/gorm"
)

type Work struct {
	Model
	LastWorkedAt       time.Time
	ConsecutiveStreaks uint16
	Streak             uint16
	Tools              uint8
}

func (Work) TableName() string {
	return "userWorkData"
}

func (w *Work) AfterCreate(tx *gorm.DB) error {
	return nil
}

// Saves the data to the database
func (w *Work) Save() {
	DB.Save(&w)
}

// Queries the database for the work data with the given user object.
func (w *Work) GetWorkInfo(u *User) {
	DB.Raw("SELECT * FROM userWorkData WHERE userWorkData.ID = ?", u.ID).First(&w)
	if w.ID == 0 {
		w.ID = u.ID
	}
}

// CanDoWork - Checks if the user can work again
// Returns true if the user can work and false if they cant
func (w *Work) CanDoWork() bool {

	since := time.Since(w.LastWorkedAt).Hours()
	return config.CONFIG.Debug.IgnoreWorkCooldown || since > float64(config.CONFIG.Work.Cooldown)
}

// Returns the time the user can work next as a formatted discord string
// https://hammertime.cyou/
func (w *Work) CanDoWorkAt() string {
	nextTime := w.LastWorkedAt.Add(time.Hour * config.CONFIG.Work.Cooldown).Unix()
	return fmt.Sprintf("<t:%d:R>", nextTime)
}

// CheckStreak - Checks the streak for the work object
// Resets it down to 0 if the user failed their streak. i.e. Waited too long since the last work
func (w *Work) CheckStreak() {
	if time.Since(w.LastWorkedAt).Hours() > float64(config.CONFIG.Work.StreakResetHours) {
		w.ConsecutiveStreaks = 0
		w.Streak = 0
	}
}

// UpdateStreakAndTime - Updates the streak for the user i.e. adding one to the counters
// and ensuring the streak is not over the max streak
// and updating the time of the last work
func (w *Work) UpdateStreakAndTime() {
	// Updates the variables
	w.LastWorkedAt = time.Now()

	w.ConsecutiveStreaks += 1
	w.Streak += 1

	// The StreakLength changed, so we need to update the streak for the player to avoid a crash
	if w.Streak > uint16(len(config.CONFIG.Work.StreakOutput)) {
		w.Streak = uint16(len(config.CONFIG.Work.StreakOutput))
	}
}

func (w *Work) HasHitMaxToolLimit() bool {
	return w.Tools >= config.CONFIG.Work.MaxTools
}

// Returns how much it would cost the user to buy a new tool
// Returns the amount as an int and formatted string
func (w *Work) CalcBuyToolPrice() (int, string) {

	multiplier := config.CONFIG.Work.ToolBasePriceMultiplier
	basePrice := float64(config.CONFIG.Work.ToolBasePrice)

	price := int(basePrice * math.Pow(multiplier, float64(w.Tools)))
	priceString := utils.HumanReadableNumber(price)

	return price, priceString
}
