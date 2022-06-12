package database

import (
	"fmt"

	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/malm"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

const resetDatabaseOnStart = true

func Connect() {
	if err := connectToDB(); err != nil {
		malm.Fatal("Database initialization error: %s", err)
		return
	}
	malm.Info("Connected to database")
}

func connectToDB() error {

	var err error
	DB, err = gorm.Open(sqlite.Open(config.CONFIG.Database.FileName), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return err
	}

	/*
		var modelList = []interface{}{
			&User{},
			&Work{},
			&Daily{},
			&Farm{},
			&FarmPlot{},
			&FarmCrop{},
			&Debug{},
		}
	*/

	if resetDatabaseOnStart {

		malm.Info("Resetting database...")

		// Populates the database with the default values
		defer PopulateDatabase()

		tableList := []string{
			(&User{}).TableName(),
			(&Work{}).TableName(),
			(&Daily{}).TableName(),
			(&Farm{}).TableName(),
			(&FarmPlot{}).TableName(),
			(&FarmCrop{}).TableName(),
			(&Debug{}).TableName()}
		/*
			if len(tableList) != len(modelList) {
				return fmt.Errorf("Table list and model list are not the same length")
			}
		*/
		for _, name := range tableList {
			DB.Exec(fmt.Sprintf("DROP TABLE %s", name))
		}
	}

	// Remeber to add new tables to the tableList and not just here!
	return DB.AutoMigrate(&User{},
		&Work{},
		&Daily{},
		&Farm{},
		&FarmPlot{},
		&FarmCrop{},
		&Debug{})
}
