package config

import (
	"bytes"
	"encoding/json"
	"os"
	"time"

	"github.com/CarlFlo/malm"
)

// CONFIG holds all the config data
var CONFIG *configStruct

type configStruct struct {
	Token               string            `json:"token"`
	BotPrefix           string            `json:"botPrefix"`
	OwnerID             string            `json:"ownerID"`
	BoundChannels       []string          `json:"boundChannels"`
	AllowDirectMessages bool              `json:"allowDirectMessages"`
	BotInfo             botInfo           `json:"botInfo"`
	Music               music             `json:"music"`
	MessageProcessing   messageProcessing `json:"messageProcessing"`
	Database            database          `json:"database"`
	Debug               debug             `json:"debug"`
	Economy             economy           `json:"economy"`
	Work                work              `json:"work"`
	Daily               daily             `json:"daily"`
	Farm                farm              `json:"farm"`
	Colors              colors            `json:"colors"`
	Emojis              emojis            `json:"emojis"`
}

type botInfo struct {
	AppID      string `json:"appID"`
	Permission uint64 `json:"permission"`
	VersionURL string `json:"versionURL"`
	DepositURL string `json:"depositURL"`
}
type music struct {
	YoutubeAPIKey        string        `json:"youtubeAPIKey"`
	MaxSongLengthMinutes int           `json:"maxSongLengthMinutes"`
	MaxCacheAgeMin       time.Duration `json:"maxCacheAgeMin"`
	MusicEnabled         bool          `json:"musicEnabled"`
}

type messageProcessing struct {
	MessageLengthLimit    int `json:"messageLengthLimit"`
	MaxIncommingMsgLength int `json:"maxIncommingMsgLength"`
}

type database struct {
	FileName string `json:"fileName"`
}

type debug struct {
	IgnoreWorkCooldown  bool `json:"ignoreWorkCooldown"`
	IgnoreDailyCooldown bool `json:"ignoreDailyCooldown"`
	IgnoreWaterCooldown bool `json:"ignoreWaterCooldown"`
}

type economy struct {
	Name          string `json:"name"`
	StartingMoney uint64 `json:"startingMoney"`
}

type work struct {
	// Cooldown in hours
	Cooldown                time.Duration `json:"cooldown"`
	MinMoney                int           `json:"minMoney"`
	MaxMoney                int           `json:"maxMoney"`
	ToolBonus               int           `json:"toolBonus"`
	ToolBasePrice           int           `json:"toolBasePrice"`
	ToolBasePriceMultiplier float64       `json:"toolBasePriceMultiplier"`
	MaxTools                uint8         `json:"maxTools"`
	StreakOutput            []string      `json:"streakOutput"`
	StreakBonus             int           `json:"streakBonus"`
	StreakResetHours        int           `json:"streakResetHours"`
}
type daily struct {
	// Cooldown in hours
	Cooldown         time.Duration `json:"cooldown"`
	MinMoney         int           `json:"minMoney"`
	MaxMoney         int           `json:"maxMoney"`
	StreakOutput     []string      `json:"streakOutput"`
	StreakBonus      int           `json:"streakBonus"`
	StreakResetHours int           `json:"streakResetHours"`
}

type farm struct {
	DefaultOwnedFarmPlots       uint8         `json:"defaultOwnedFarmPlots"`
	CropSeedPrice               int           `json:"cropSeedPrice"`
	FarmPlotPrice               int           `json:"farmPlotPrice"`
	FarmPlotCostMultiplier      float64       `json:"farmPlotCostMultiplier"`
	MaxPlots                    uint8         `json:"maxPlots"`
	WaterCooldown               time.Duration `json:"waterCooldown"`
	WaterCropTimeReductionHours time.Duration `json:"waterCropTimeReductionHours"`
	CropsPreishAfter            time.Duration `json:"cropsPreishAfter"`
}

type colors struct {
	Success int `json:"success"`
	Failure int `json:"failure"`
	Neutral int `json:"neutral"`
}

type emojis struct {
	ComponentEmojiNames componentEmojiNames `json:"componentEmojiNames"`
	Bank                string              `json:"bank"`
	Wallet              string              `json:"wallet"`
	Economy             string              `json:"economy"`
	NetWorth            string              `json:"netWorth"`
	Success             string              `json:"success"`
	Failure             string              `json:"failure"`
	PerishedCrop        string              `json:"perishedCrop"`
	EmptyPlot           string              `json:"emptyPlot"`
	Tools               string              `json:"tools"`
	MusicNotes          string              `json:"musicNotes"`
	MusicPlaying        string              `json:"musicPlaying"`
	MusicPaused         string              `json:"musicPaused"`
}

type componentEmojiNames struct {
	MoneyBag string `json:"moneyBag"`
	Help     string `json:"help"`
	Refresh  string `json:"refresh"`
}

// ReloadConfig is a wrapper function for reloading the config. For clarity
func ReloadConfig() error {
	return readConfig()
}

// readConfig will read the config file
func readConfig() error {

	file, err := os.Open("./config.json")
	if err != nil {
		return err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(file)
	file.Close()

	if err = json.Unmarshal(buf.Bytes(), &CONFIG); err != nil {
		return err
	}

	return nil
}

// createConfig creates the default config file
func createConfig() error {

	// Default config settings
	configStruct := configStruct{
		Token:               "",
		BotPrefix:           ",",
		OwnerID:             "",
		BoundChannels:       []string{},
		AllowDirectMessages: true,
		BotInfo: botInfo{
			AppID:      "",
			Permission: 139690691648,
			VersionURL: "https://raw.githubusercontent.com/CarlFlo/DiscordMoneyBot/master/main.go",
			DepositURL: "https://github.com/CarlFlo/dootBot",
		},
		Music: music{
			YoutubeAPIKey:        "",
			MaxSongLengthMinutes: 60,
			MaxCacheAgeMin:       90,
		},
		MessageProcessing: messageProcessing{
			MessageLengthLimit:    1850, // The meximum length a send message can be before it will be split.
			MaxIncommingMsgLength: 0,    // Set to 0 for ignore
		},
		Database: database{
			FileName: "database.db",
		},
		Economy: economy{
			Name:          "credits",
			StartingMoney: 0,
		},
		Work: work{
			Cooldown:                6,
			MinMoney:                100,
			MaxMoney:                250,
			ToolBonus:               45,
			ToolBasePrice:           50,
			ToolBasePriceMultiplier: 1.25,
			MaxTools:                20,
			StreakOutput:            []string{":regional_indicator_b:", ":regional_indicator_o:", ":regional_indicator_n:", ":regional_indicator_u:", ":regional_indicator_s:"},
			StreakBonus:             350,
			StreakResetHours:        24,
		},
		Daily: daily{
			Cooldown:         24,
			MinMoney:         350,
			MaxMoney:         800,
			StreakOutput:     []string{":one:", ":two:", ":three:", ":four:", ":five:", ":six:", ":seven:"},
			StreakBonus:      2000,
			StreakResetHours: 24,
		},
		Farm: farm{
			DefaultOwnedFarmPlots:       1,
			CropSeedPrice:               50,
			FarmPlotPrice:               5000,
			FarmPlotCostMultiplier:      1.4,
			MaxPlots:                    12,
			WaterCooldown:               2,
			WaterCropTimeReductionHours: 1,
			CropsPreishAfter:            24,
		},
		Colors: colors{
			Success: 0x198754,
			Failure: 0xE9302A,
			Neutral: 0x006ED0,
		},
		Emojis: emojis{
			ComponentEmojiNames: componentEmojiNames{
				MoneyBag: "💰",
				Help:     "💡", // Alt: 💡, ❔
				Refresh:  "🔄",
			},
			Bank:         ":bank:",
			Wallet:       ":dollar:",
			Economy:      ":moneybag:",
			NetWorth:     ":bar_chart:",
			Success:      ":white_check_mark:",
			Failure:      ":x:",
			PerishedCrop: ":wilted_rose:",
			EmptyPlot:    ":brown_square:",
			Tools:        ":tools:",
			MusicNotes:   ":musical_note:",
			MusicPlaying: ":arrow_forward:",
			MusicPaused:  ":pause_button:",
		},
		Debug: debug{
			IgnoreWorkCooldown:  false,
			IgnoreDailyCooldown: false,
			IgnoreWaterCooldown: false,
		},
	}

	jsonData, _ := json.MarshalIndent(configStruct, "", "   ")
	err := os.WriteFile("config.json", jsonData, 0644)

	return err
}

// loadConfiguration loads the configuration file into memory
func loadConfiguration() error {

	if err := readConfig(); err != nil {
		if err = createConfig(); err != nil {
			return err
		}
		if err = readConfig(); err != nil {
			return err
		}
	}
	return nil
}

func Load() {
	if err := loadConfiguration(); err != nil {
		malm.Fatal("Error loading configuration: %s", err)
		return
	}

	requiredVariableCheck()

	malm.Info("Configuration loaded")
}

// Some variables are required for the bot to work
func requiredVariableCheck() {

	// This function checks if some important variables are set in the config file
	problem := false

	if len(CONFIG.Token) == 0 {
		malm.Error("No bot Token provided in the config file!")
		problem = true
	}

	if len(CONFIG.BotInfo.AppID) == 0 {
		malm.Error("No AppID provided in the config file! (The bot's Discord ID)")
		problem = true
	}

	if len(CONFIG.OwnerID) == 0 {
		malm.Error("No OwnerID provided in the config file! (This should be your Discord ID)")
		problem = true
	}

	if problem {
		malm.Fatal("There are at least one variable missing in the configuration file. Please fix the above errors!")
	}
}
