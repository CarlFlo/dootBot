package database

import (
	"time"
)

func PopulateDatabase() {
	//debug()
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
			Emoji:          "🍅",
			DurationToGrow: time.Minute * 20,
			HarvestReward:  100,
		}, {
			Name:           "Potato",
			Emoji:          "🥔",
			DurationToGrow: time.Minute * 30,
			HarvestReward:  125,
		}, {
			Name:           "Pineapple",
			Emoji:          "🍍",
			DurationToGrow: time.Hour,
			HarvestReward:  180,
		}, {
			Name:           "Strawberry",
			Emoji:          "🍓",
			DurationToGrow: time.Hour * 3,
			HarvestReward:  400,
		}, {
			Name:           "Corn",
			Emoji:          "🌽",
			DurationToGrow: time.Hour * 6,
			HarvestReward:  700,
		}, {
			Name:           "Mango",
			Emoji:          "🥭",
			DurationToGrow: time.Hour * 12,
			HarvestReward:  1250,
		}, {
			Name:           "Watermelon",
			Emoji:          "🍉",
			DurationToGrow: time.Hour * 24,
			HarvestReward:  2350,
		}, {
			Name:           "Apple",
			Emoji:          "🍎",
			DurationToGrow: time.Hour * 24 * 2,
			HarvestReward:  4400,
		}, {
			Name:           "Onion",
			Emoji:          "🧅",
			DurationToGrow: time.Hour * 24 * 3,
			HarvestReward:  6500,
		}, {
			Name:           "Carrot",
			Emoji:          "🥕",
			DurationToGrow: time.Hour * 24 * 4,
			HarvestReward:  8500,
		}, {
			Name:           "Banana",
			Emoji:          "🍌",
			DurationToGrow: time.Hour * 24 * 6,
			HarvestReward:  12500,
		}, {
			Name:           "Hot Pepper",
			Emoji:          "🌶️",
			DurationToGrow: time.Hour * 24 * 8,
			HarvestReward:  16500,
		}, {
			Name:           "Avocado",
			Emoji:          "🥑",
			DurationToGrow: time.Hour * 24 * 10,
			HarvestReward:  21500,
		}, {
			Name:           "Grapes",
			Emoji:          "🍇",
			DurationToGrow: time.Hour * 24 * 15,
			HarvestReward:  32500,
		}, {
			Name:           "Peach",
			Emoji:          "🍑",
			DurationToGrow: time.Hour * 24 * 25,
			HarvestReward:  55000,
		},
	}

	for _, crop := range crops {
		DB.Create(&crop)
	}

}
