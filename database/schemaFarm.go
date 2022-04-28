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
	Plots         []FarmPlot
	OwnedPlots    uint8
	LastWateredAt time.Time // Last time the user watered the farm plots

	PlotsChanged bool `gorm:"-"` // Ignored by the database
}

/*
	One farm per user, but you can have many plots
	The farm plot keeps track of which farm it belongs to
*/

func (Farm) TableName() string {
	return "userFarms"
}

func (f *Farm) BeforeCreate(tx *gorm.DB) error {

	f.LastWateredAt = time.Now().Add(time.Hour * config.CONFIG.Farm.WaterCooldown * -1)
	return nil
}

// Saves the data to the database
func (f *Farm) Save() {
	DB.Save(&f)

	// Updates/saves the plots as well
	if f.PlotsChanged {
		for _, plot := range f.Plots {
			plot.Save()
		}
	}
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

	malm.Debug("%v > %v", time.Since(f.LastWateredAt).Hours(), float64(config.CONFIG.Farm.WaterCooldown))

	since := time.Since(f.LastWateredAt).Hours()
	return config.CONFIG.Debug.IgnoreWorkCooldown || since > float64(config.CONFIG.Farm.WaterCooldown)
}

// Returns the time the user can water their crops as a formatted discord string
// https://hammertime.cyou/
func (f *Farm) CanWaterAt() string {
	nextTime := f.LastWateredAt.Add(time.Hour * config.CONFIG.Farm.WaterCooldown).Unix()
	return fmt.Sprintf("<t:%d:R>", nextTime)
}

// Functions waters every single plot
// Meaning it will update the plantedAt time
func (f *Farm) WaterPlots() {

	// Update last watered at
	f.LastWateredAt = time.Now()
	// A change was made so it needs to be saved when farm Save function is called
	f.PlotsChanged = true

	for _, plot := range f.Plots {
		plot.Water()
	}
}

// Will return the amount of crops that have perished from plots
// Will also remove perished crop plots from the database
// Untested
func (f *Farm) CropsPerishedCheck() []string {

	// Checks if the user has watered their crops in the last x hours
	if float64(config.CONFIG.Farm.CropsPreishAfter) > time.Since(f.LastWateredAt).Hours() {
		return []string{}
	}

	// User has not watered in the last x hours. Missing the set deadline
	// Ungrown crops will perish

	var perishedCrops []string

	for _, plot := range f.Plots {

		plot.GetCropInfo()

		if plot.HasFullyGrown() {
			// Crop has fully grown so we need to check if the user didn't
			// just wait the whole growth duration without watering it

			crop := &plot.Crop
			fullyGrownAt := plot.PlantedAt.Add(crop.DurationToGrow)

			waterDeadline := fullyGrownAt.Add(time.Hour * config.CONFIG.Farm.CropsPreishAfter * -1)

			if !f.LastWateredAt.Before(waterDeadline) {
				continue
			}
		}

		perishedCrops = append(perishedCrops, plot.Crop.Name)
		plot.DeleteFromDB()
	}

	return perishedCrops
}
