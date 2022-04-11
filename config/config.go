package config

import (
	"encoding/json"
	"io/ioutil"
	"time"

	"github.com/CarlFlo/malm"
)

// Redo this so this isnt required https://www.youtube.com/watch?v=y_eIBmt3JdY

// CONFIG holds all the config data
var CONFIG *configStruct

type configStruct struct {
	Token               string            `json:"token"`
	BotPrefix           string            `json:"botPrefix"`
	Version             string            `json:"version"`
	OwnerID             string            `json:"ownerID"`
	DispConfOnStart     bool              `json:"dispConfOnStart"`
	BoundChannels       []string          `json:"boundChannels"`
	AllowDirectMessages bool              `json:"allowDirectMessages"`
	BotInfo             botInfo           `json:"botInfo"`
	MessageProcessing   messageProcessing `json:"messageProcessing"`
	Database            database          `json:"database"`
	Work                work              `json:"work"`
	Daily               daily             `json:"daily"`
}

type botInfo struct {
	ClientID   string `json:"clientID"`
	Permission uint64 `json:"permission"`
}

type messageProcessing struct {
	MessageLengthLimit    int `json:"messageLengthLimit"`
	MaxIncommingMsgLength int `json:"maxIncommingMsgLength"`
}

type database struct {
	FileName string `json:"fileName"`
}

type work struct {
	// Cooldown in hours
	Cooldown         time.Duration `json:"cooldown"`
	MinMoney         int           `json:"minMoney"`
	MaxMoney         int           `json:"maxMoney"`
	ToolBonus        int           `json:"toolBonus"`
	StreakLength     uint16        `json:"streakLength"`
	StreakBonus      int           `json:"streakBonus"`
	StreakResetHours int           `json:"streakResetHours"`
}

type daily struct {
	// Cooldown in hours
	Cooldown         time.Duration `json:"cooldown"`
	MinMoney         int           `json:"minMoney"`
	MaxMoney         int           `json:"maxMoney"`
	StreakLength     uint16        `json:"streakLength"`
	StreakBonus      int           `json:"streakBonus"`
	StreakResetHours int           `json:"streakResetHours"`
}

// ReloadConfig is a wrapper function for reloading the config. For clarity
func ReloadConfig() error {
	return readConfig()
}

// readConfig will read the config file
func readConfig() error {

	file, err := ioutil.ReadFile("./config.json")

	if err != nil {
		return err
	}

	if err = json.Unmarshal(file, &CONFIG); err != nil {
		return err
	}

	if CONFIG.DispConfOnStart {
		malm.Debug("Config:\n%s", string(file))
	}

	return nil
}

// createConfig creates the default config file
func createConfig() error {

	// Default config settings
	configStruct := configStruct{
		Token:               "",
		BotPrefix:           ",",
		Version:             "2022-04-10",
		OwnerID:             "",
		DispConfOnStart:     false,
		BoundChannels:       []string{},
		AllowDirectMessages: true,
		BotInfo: botInfo{
			ClientID:   "",
			Permission: 2184694848,
		},
		MessageProcessing: messageProcessing{
			MessageLengthLimit:    1850, // The meximum length a send message can be before it will be split.
			MaxIncommingMsgLength: 0,    // Set to 0 for ignore
		}, Database: database{
			FileName: "database.db",
		}, Work: work{
			Cooldown:         6,
			MinMoney:         100,
			MaxMoney:         250,
			ToolBonus:        100,
			StreakLength:     5,
			StreakBonus:      1000,
			StreakResetHours: 24,
		},
		Daily: daily{
			Cooldown:         24,
			MinMoney:         1000,
			MaxMoney:         2500,
			StreakLength:     7,
			StreakBonus:      10000,
			StreakResetHours: 48,
		},
	}

	jsonData, _ := json.MarshalIndent(configStruct, "", "   ")
	err := ioutil.WriteFile("config.json", jsonData, 0644)

	return err
}

// LoadConfiguration loads the configuration file into memory
func LoadConfiguration() error {

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
