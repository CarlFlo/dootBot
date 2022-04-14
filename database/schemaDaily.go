package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"gorm.io/gorm"
)

type Daily struct {
	gorm.Model
	LastDailyAt        time.Time
	ConsecutiveStreaks uint16
	Streak             uint16
}

func (Daily) TableName() string {
	return "userDailyData"
}

func (d *Daily) AfterCreate(tx *gorm.DB) error {

	// January 1st 1970
	d.LastDailyAt = time.Unix(0, 0).UTC()
	return nil
}

// Queries the database for the daily data with the given discord ID.
// The object which calls the method will be updated with the user's daily data
func (d *Daily) GetDailyByDiscordID(discordID string) {
	DB.Raw("SELECT * FROM userDailyData JOIN users ON userDailyData.ID = users.ID WHERE users.discord_id = ?", discordID).First(&d)
}

// CanDoDaily - Checks if the user can do their daily again
// Returns true if the user can do their daily and false if they cant
func (d *Daily) CanDoDaily() bool {

	since := time.Since(d.LastDailyAt).Hours()
	return config.CONFIG.Debug.IgnoreDailyCooldown || since > float64(config.CONFIG.Daily.Cooldown)
}

// CheckStreak - Checks the streak for the work object
// Resets it down to 0 if the user failed their streak. i.e. Waited too long since the last work
func (d *Daily) CheckStreak() {
	if time.Since(d.LastDailyAt).Hours() > float64(config.CONFIG.Daily.StreakResetHours) {
		d.ConsecutiveStreaks = 0
		d.Streak = 0
	}
}

// UpdateStreakAndTime - Updates the daily streak for the user i.e. adding one to the counters
// and ensuring the streak is not over the max streak
// and updating the time of the last daily
func (d *Daily) UpdateStreakAndTime() {
	// Updates the variables
	d.LastDailyAt = time.Now()

	d.ConsecutiveStreaks += 1
	d.Streak += 1

	// The StreakLength changed, so we need to update the streak for the player to avoid a crash
	if d.Streak > uint16(len(config.CONFIG.Daily.StreakOutput)) {
		d.Streak = uint16(len(config.CONFIG.Daily.StreakOutput))
	}
}
