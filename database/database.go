package database

import (
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func connectToDB() error {

	var err error
	DB, err = gorm.Open(sqlite.Open(config.CONFIG.Database.FileName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	DB.AutoMigrate(&User{}, &Work{})

	return nil
}

func SetupDatabase() error {

	err := connectToDB()
	if err != nil {
		return err
	}

	return nil
}
