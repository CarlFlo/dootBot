package database

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
	"gorm.io/gorm"
)

type Daily struct {
	Model
	LastDailyAt        time.Time
	ConsecutiveStreaks uint16
	Streak             uint16
}

func (Daily) TableName() string {
	return "userDailyData"
}

func (d *Daily) AfterCreate(tx *gorm.DB) error {
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

// Does the daily
// Returns
// If the daily could be executed
// The money earned as a pretty string
// Streak reward value
// Percentage of streak completed
// Title for the message
// Footer for the message
func (d *Daily) DoDaily(user *User) (bool, string, string, string, string, string) {

	// Resets streaks down to 0 if the user failed their streak.
	d.checkStreak()

	// Can't do their daily
	if !d.CanDoDaily() {
		streakReward, streakPercentage := d.generateDailyStreakMessage()
		return false,
			"",
			streakReward,
			streakPercentage,
			fmt.Sprintf("%s Slow down!", config.CONFIG.Emojis.Failure),
			fmt.Sprintf("You can get your daily once every %d hours!", int(config.CONFIG.Daily.Cooldown))
	}

	d.updateStreakAndTime()

	moneyEarned := d.generateDailyIncome()
	user.AddMoney(uint64(moneyEarned))

	moneyEarnedString := utils.HumanReadableNumber(moneyEarned)
	streakReward, streakPercentage := d.generateDailyStreakMessage()

	title := "Daily Bonus"
	footer := fmt.Sprintf("Completing your streak will earn you an extra %d %s!\nThe streak resets after %d hours of inactivity.",
		config.CONFIG.Daily.StreakBonus,
		config.CONFIG.Economy.Name,
		config.CONFIG.Daily.StreakResetHours)

	return true, moneyEarnedString, streakReward, streakPercentage, title, footer
}

func (d *Daily) generateDailyIncome() int {

	// Generate a random int between config.CONFIG.Daily.MinMoney and config.CONFIG.Daily.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Daily.MaxMoney-config.CONFIG.Daily.MinMoney) + config.CONFIG.Daily.MinMoney

	// Adds the streak bonus to the amount
	if d.Streak == uint16(len(config.CONFIG.Daily.StreakOutput)) {
		moneyEarned += config.CONFIG.Daily.StreakBonus
	}

	return moneyEarned
}

func (d *Daily) generateDailyStreakMessage() (string, string) {

	percentage := float64(d.Streak) / float64(len(config.CONFIG.Daily.StreakOutput))
	upTo := int(float64(len(config.CONFIG.Daily.StreakOutput)) * percentage)

	// Append to a string values in config.CONFIG.Daily.StreakOutput up to the index of upTo
	var visualStreakProgress string

	for i := 0; i < upTo; i++ {
		visualStreakProgress += fmt.Sprintf("%s ", config.CONFIG.Daily.StreakOutput[i])
	}
	for i := upTo; i < len(config.CONFIG.Daily.StreakOutput); i++ {
		visualStreakProgress += "- "
	}

	percentageText := fmt.Sprintf("%d%%", int(percentage*100))

	var streakMessage string
	if d.CanDoDaily() && d.Streak == uint16(len(config.CONFIG.Daily.StreakOutput)) {
		streakMessage = fmt.Sprintf("An additional ``%s`` %s were added to your daily earnings!", utils.HumanReadableNumber(config.CONFIG.Daily.StreakBonus), config.CONFIG.Economy.Name)
	}

	return fmt.Sprintf("%s %s", visualStreakProgress, streakMessage), percentageText
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

// checkStreak - Checks the streak for the work object
// Resets it down to 0 if the user failed their streak. i.e. Waited too long since the last work
func (d *Daily) checkStreak() {
	if time.Since(d.LastDailyAt).Hours() > float64(config.CONFIG.Daily.StreakResetHours) {
		d.ConsecutiveStreaks = 0
		d.Streak = 0
	}
}

// updateStreakAndTime - Updates the daily streak for the user i.e. adding one to the counters
// and ensuring the streak is not over the max streak
// and updating the time of the last daily
func (d *Daily) updateStreakAndTime() {
	// Updates the variables
	d.LastDailyAt = time.Now()

	d.ConsecutiveStreaks += 1
	d.Streak += 1

	// The StreakLength changed, so we need to update the streak for the player to avoid a crash
	if d.Streak > uint16(len(config.CONFIG.Daily.StreakOutput)) {
		d.Streak = uint16(len(config.CONFIG.Daily.StreakOutput))
	}

}
