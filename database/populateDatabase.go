package database

import (
	"time"

	"github.com/CarlFlo/malm"
)

func PopulateDatabase() {
	malm.Info("Populating database...")
	farming()
}

func farming() {

	// Default crops
	// duration / reward = ratio

	crops := []FarmCrop{
		{
			Name:           "Tomato",
			Emoji:          ":tomato:",
			DurationToGrow: time.Minute * 20,
			HarvestReward:  1000,
		}, {
			Name:           "Potato",
			Emoji:          ":potato:",
			DurationToGrow: time.Minute * 30,
			HarvestReward:  1200,
		}, {
			Name:           "Pineapple",
			Emoji:          ":pineapple:",
			DurationToGrow: time.Hour,
			HarvestReward:  1750,
		}, {
			Name:           "Strawberry",
			Emoji:          ":strawberry:",
			DurationToGrow: time.Hour * 3,
			HarvestReward:  3800,
		}, {
			Name:           "Corn",
			Emoji:          ":corn:",
			DurationToGrow: time.Hour * 6,
			HarvestReward:  6800,
		}, {
			Name:           "Mango",
			Emoji:          ":mango: ",
			DurationToGrow: time.Hour * 12,
			HarvestReward:  12200,
		}, {
			Name:           "Watermelon",
			Emoji:          ":watermelon:",
			DurationToGrow: time.Hour * 24,
			HarvestReward:  23000,
		}, {
			Name:           "Apple",
			Emoji:          ":apple:",
			DurationToGrow: time.Hour * 24 * 2,
			HarvestReward:  43000,
		}, {
			Name:           "Onion",
			Emoji:          ":onion:",
			DurationToGrow: time.Hour * 24 * 3,
			HarvestReward:  62000,
		}, {
			Name:           "Carrot",
			Emoji:          ":carrot:",
			DurationToGrow: time.Hour * 24 * 4,
			HarvestReward:  76000,
		}, {
			Name:           "Banana",
			Emoji:          ":banana:",
			DurationToGrow: time.Hour * 24 * 6,
			HarvestReward:  110000,
		}, {
			Name:           "Hot Pepper",
			Emoji:          ":hot_pepper:",
			DurationToGrow: time.Hour * 24 * 8,
			HarvestReward:  150000,
		}, {
			Name:           "Avocado",
			Emoji:          ":avocado:",
			DurationToGrow: time.Hour * 24 * 10,
			HarvestReward:  195000,
		}, {
			Name:           "Grapes",
			Emoji:          ":grapes:",
			DurationToGrow: time.Hour * 24 * 15,
			HarvestReward:  305000,
		}, {
			Name:           "Peach",
			Emoji:          ":peach:",
			DurationToGrow: time.Hour * 24 * 25,
			HarvestReward:  550000,
		},
	}

	for _, crop := range crops {
		DB.Create(&crop)
	}

}
