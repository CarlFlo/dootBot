package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
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
	DB.Table("Users").Where("discord_id = ?", discordID).First(&u)
}

type Work struct {
	gorm.Model
	LastWorkedAt       time.Time
	ConsecutiveStreaks uint16
	Streak             uint16
	Tools              uint8
}

func (w *Work) AfterCreate(tx *gorm.DB) error {

	// January 1st 1970
	w.LastWorkedAt = time.Unix(0, 0).UTC()
	return nil
}

// Queries the database for the work data with the given discord ID.
// The object which calls the method will be updated with the user's work data
func (w *Work) GetWorkByDiscordID(discordID string) {
	DB.Raw("select * from Works JOIN Users ON Works.ID = Users.ID WHERE Users.discord_id = ?", discordID).First(&w)
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
	if w.Streak > config.CONFIG.Work.StreakLength {
		w.Streak = config.CONFIG.Work.StreakLength
	}
}

type Daily struct {
	gorm.Model
	LastDailyAt        time.Time
	ConsecutiveStreaks uint16
	Streak             uint16
}

func (d *Daily) AfterCreate(tx *gorm.DB) error {

	// January 1st 1970
	d.LastDailyAt = time.Unix(0, 0).UTC()
	return nil
}

// Queries the database for the daily data with the given discord ID.
// The object which calls the method will be updated with the user's daily data
func (d *Daily) GetDailyByDiscordID(discordID string) {
	DB.Raw("select * from dalies JOIN Users ON dalies.ID = Users.ID WHERE Users.discord_id = ?", discordID).First(&d)
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
	if d.Streak > config.CONFIG.Daily.StreakLength {
		d.Streak = config.CONFIG.Daily.StreakLength
	}
}
