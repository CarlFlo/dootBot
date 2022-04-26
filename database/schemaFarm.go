package database

import (
	"fmt"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"gorm.io/gorm"
)

type Farm struct {
	gorm.Model
	Plots       []FarmPlot
	OwnedPlots  uint8
	LastWatered time.Time // Last time the user watered the farm plots
}

/*
	One farm per user, but you can have many plots
	The farm plot keeps track of which farm it belongs to
*/

func (Farm) TableName() string {
	return "userFarms"
}

func (f *Farm) BeforeCreate(tx *gorm.DB) error {

	// January 1st 1970
	f.LastWatered = time.Unix(0, 0).UTC()
	return nil
}

// Saves the data to the database
func (f *Farm) Save() {
	DB.Save(&f)
}

// Queries the database for the farm data with the given user object.
// Needs to be called first
func (f *Farm) GetUserFarmData(u *User) {
	DB.Raw("SELECT * FROM userFarms WHERE userFarms.ID = ?", u.ID).First(&f)
	if f.ID == 0 { // Meaning there is not data so we initialize it
		f.ID = u.ID // The farms index is the same as the user
		f.OwnedPlots = config.CONFIG.Farm.DefaultOwnedFarmPlots

		f.Save()
	}
}

// Updates the object to contain all the farmplots
func (f *Farm) GetFarmPlots() {
	DB.Raw("SELECT * FROM userFarmPlots WHERE userFarmPlots.Farm_ID = ? LIMIT ?", f.ID, f.OwnedPlots).Find(&f.Plots)
}

// Rember to run GetFarmPlots() before running this function
func (f *Farm) GetUnusedPlots() int {
	return int(f.OwnedPlots) - len(f.Plots)
}

// Returns true if the user has a free (unused) farm plot
func (f *Farm) HasFreePlot() bool {
	return f.GetUnusedPlots() > 0
}

// Returns true if the user can water their crops
func (f *Farm) CanWater() bool {

	malm.Debug("%v > %v", time.Since(f.LastWatered).Hours(), float64(config.CONFIG.Farm.WaterCooldown))

	since := time.Since(f.LastWatered).Hours()
	return config.CONFIG.Debug.IgnoreWorkCooldown || since > float64(config.CONFIG.Farm.WaterCooldown)
}

// Returns the time the user can water their crops as a formatted discord string
// https://hammertime.cyou/
func (f *Farm) CanWaterAt() string {
	nextTime := f.LastWatered.Add(time.Hour * config.CONFIG.Farm.WaterCooldown).Unix()
	return fmt.Sprintf("<t:%d:R>", nextTime)
}

func (f *Farm) WaterPlots() {

	for _, plot := range f.Plots {
		plot.Water()
		plot.Save()
	}

}
