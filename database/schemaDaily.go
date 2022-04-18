package database

import (
	"fmt"
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

// Saves the data to the database
func (d *Daily) Save() {
	DB.Save(&d)
}

// Queries the database for the daily data with the given user object.
func (d *Daily) GetDailyInfo(u *User) {
	DB.Raw("SELECT * FROM userDailyData WHERE userDailyData.ID = ?", u.ID).First(&d)
	if d.ID == 0 {
		d.ID = u.ID
	}
}

// CanDoDaily - Checks if the user can do their daily again
// Returns true if the user can do their daily and false if they cant
func (d *Daily) CanDoDaily() bool {

	since := time.Since(d.LastDailyAt).Hours()
	return config.CONFIG.Debug.IgnoreDailyCooldown || since > float64(config.CONFIG.Daily.Cooldown)
}

// Returns the time the user can do their daily next as a formatted discord string
// https://hammertime.cyou/
func (d *Daily) CanDoDailyAt() string {
	nextTime := d.LastDailyAt.Add(time.Hour * config.CONFIG.Daily.Cooldown).Unix()
	return fmt.Sprintf("<t:%d:R>", nextTime)
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
