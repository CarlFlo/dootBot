package database

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func connectToDB() error {

	DB, err := gorm.Open(sqlite.Open("database.db"), &gorm.Config{})
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
