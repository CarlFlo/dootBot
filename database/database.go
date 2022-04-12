package database

import (
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

const resetDatabaseOnStart = true

func connectToDB() error {

	var err error
	DB, err = gorm.Open(sqlite.Open(config.CONFIG.Database.FileName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	if resetDatabaseOnStart {
		malm.Info("Resetting database...")
		DB.Exec("DROP TABLE users")
		DB.Exec("DROP TABLE works")
		DB.Exec("DROP TABLE dailies")
	}

	DB.AutoMigrate(&User{}, &Work{}, &Daily{})

	return nil
}

func SetupDatabase() error {

	err := connectToDB()
	if err != nil {
		return err
	}

	return nil
}
