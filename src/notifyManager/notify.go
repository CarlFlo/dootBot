package notifyManager

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/malm"
)

// This module is responsible for sending notifications to users.
// Saving a time and who will be notified in a database
// Loading the from the database in intervals to notify the user

var stopper = make(chan interface{})

func Initialize() {

	ticker := time.NewTicker(time.Minute * config.CONFIG.NotifySettings.CheckInterval)

	go func() {
		for {
			select {
			case <-ticker.C:
				checkDatabase()
			case <-stopper:
				ticker.Stop()
				malm.Info("Notification manager stopped")
				return
			}
		}
	}()
	malm.Info("Notification manager initialized")
}

func Stop() {
	stopper <- nil
}

func checkDatabase() {

	// Check the database for any notifications.
	// Only check notifications that are time.Minute * config.CONFIG.NotifySettings.CheckInterval in the future
	// Delete those that are sent
	// Make sure there are not duplicates
	// Notifications that are old (i.e. due when the bot was turned off must also be run immediately)

	// Idea
	// Get all notifications that are due within the ticker's interval
	// Also get all notifications that were due before the ticker's interval. So 5 minutes in the future and 5+ minutes in the past

}
