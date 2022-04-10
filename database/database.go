package database

import (
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func connectToDB() error {

	DB, err := gorm.Open(sqlite.Open(config.CONFIG.Database.FileName), &gorm.Config{})
	if err != nil {
		return err
	}

	DB.AutoMigrate(&User{})

	return nil
}

func SetupDatabase() error {

	err := connectToDB()
	if err != nil {
		return err
	}

	return nil
}
