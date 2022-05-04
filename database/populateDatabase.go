package database

import (
	"time"

	"github.com/CarlFlo/malm"
)

func PopulateDatabase() {
	malm.Info("Populating database...")
	debug()
	farming()
}

func debug() {
	DB.Create(&Debug{
		DailyCount: 0,
		WorkCount:  0,
	})
}

func farming() {

	// Default crops
	// (reward-seedprice) / duration  = ratio

	crops := []FarmCrop{
		{
			Name:           "Tomato",
			Emoji:          ":tomato:",
			DurationToGrow: time.Minute * 20,
			HarvestReward:  100,
		}, {
			Name:           "Potato",
			Emoji:          ":potato:",
			DurationToGrow: time.Minute * 30,
			HarvestReward:  120,
		}, {
			Name:           "Pineapple",
			Emoji:          ":pineapple:",
			DurationToGrow: time.Hour,
			HarvestReward:  175,
		}, {
			Name:           "Strawberry",
			Emoji:          ":strawberry:",
			DurationToGrow: time.Hour * 3,
			HarvestReward:  380,
		}, {
			Name:           "Corn",
			Emoji:          ":corn:",
			DurationToGrow: time.Hour * 6,
			HarvestReward:  680,
		}, {
			Name:           "Mango",
			Emoji:          ":mango: ",
			DurationToGrow: time.Hour * 12,
			HarvestReward:  1220,
		}, {
			Name:           "Watermelon",
			Emoji:          ":watermelon:",
			DurationToGrow: time.Hour * 24,
			HarvestReward:  2300,
		}, {
			Name:           "Apple",
			Emoji:          ":apple:",
			DurationToGrow: time.Hour * 24 * 2,
			HarvestReward:  4300,
		}, {
			Name:           "Onion",
			Emoji:          ":onion:",
			DurationToGrow: time.Hour * 24 * 3,
			HarvestReward:  6200,
		}, {
			Name:           "Carrot",
			Emoji:          ":carrot:",
			DurationToGrow: time.Hour * 24 * 4,
			HarvestReward:  8000,
		}, {
			Name:           "Banana",
			Emoji:          ":banana:",
			DurationToGrow: time.Hour * 24 * 6,
			HarvestReward:  11500,
		}, {
			Name:           "Hot Pepper",
			Emoji:          ":hot_pepper:",
			DurationToGrow: time.Hour * 24 * 8,
			HarvestReward:  15000,
		}, {
			Name:           "Avocado",
			Emoji:          ":avocado:",
			DurationToGrow: time.Hour * 24 * 10,
			HarvestReward:  20000,
		}, {
			Name:           "Grapes",
			Emoji:          ":grapes:",
			DurationToGrow: time.Hour * 24 * 15,
			HarvestReward:  30000,
		}, {
			Name:           "Peach",
			Emoji:          ":peach:",
			DurationToGrow: time.Hour * 24 * 25,
			HarvestReward:  50000,
		},
	}

	for _, crop := range crops {
		DB.Create(&crop)
	}

}
