package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"gorm.io/gorm"
)

type Work struct {
	gorm.Model
	LastWorkedAt       time.Time
	ConsecutiveStreaks uint16
	Streak             uint16
	Tools              uint8
}

func (Work) TableName() string {
	return "userWorkData"
}

func (w *Work) AfterCreate(tx *gorm.DB) error {

	// January 1st 1970
	w.LastWorkedAt = time.Unix(0, 0).UTC()
	return nil
}

// Queries the database for the work data with the given discord ID.
// The object which calls the method will be updated with the user's work data
func (w *Work) GetWorkByDiscordID(discordID string) {
	DB.Raw("SELECT * FROM userWorkData JOIN users ON userWorkData.ID = users.ID WHERE users.discord_id = ?", discordID).First(&w)
}

// CanDoWork - Checks if the user can work again
// Returns true if the user can work and false if they cant
func (w *Work) CanDoWork() bool {

	since := time.Since(w.LastWorkedAt).Hours()
	return config.CONFIG.Debug.IgnoreWorkCooldown || since > float64(config.CONFIG.Work.Cooldown)
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
