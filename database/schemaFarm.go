package database

import (
	"fmt"
	"math"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"gorm.io/gorm"
)

type Farm struct {
	Model
	Plots         []*FarmPlot
	OwnedPlots    uint8
	LastWateredAt time.Time // Last time the user watered the farm plots

	PlotsChanged    bool `gorm:"-"` // Ignored by the database
	HarvestEarnings int  `gorm:"-"` // If 0 then no earnings
}

/*
	One farm per user, but you can have many plots
	The farm plot keeps track of which farm it belongs to

	Limit the crops the user can plant
	They unlock better crops by planting the basic first
	Planting crop ID 1 will then unlock ID 2 and so on
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

	// Updates/saves the plots as well
	if f.PlotsChanged {
		for _, plot := range f.Plots {
			plot.Save()
		}
	}

	DB.Save(&f)
}

// Queries the database for the farm data with the given user object.
// Needs to be called first
func (f *Farm) QueryUserFarmData(u *User) {
	DB.Raw("SELECT * FROM userFarms WHERE userFarms.ID = ?", u.ID).First(&f)
	if f.ID == 0 { // Meaning there is no data, so we initialize it
		f.ID = u.ID // The farms index is the same as the user
		f.OwnedPlots = config.CONFIG.Farm.DefaultOwnedFarmPlots

		f.Save()
	}
}

// Updates the object to contain all the farmplots
func (f *Farm) QueryFarmPlots() {
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

// Returns true if anything is planted on the farm
// Run QueryFarmPlots() first
func (f *Farm) HasPlantedPlots() bool {
	return len(f.Plots) > 0
}

// Returns true if the user can water their crops
func (f *Farm) CanWater() bool {
	since := time.Since(f.LastWateredAt).Hours()
	return config.CONFIG.Debug.IgnoreWorkCooldown || since > float64(config.CONFIG.Farm.WaterCooldown)
}

// Returns the time the user can water their crops as a formatted discord string
// https://hammertime.cyou/
func (f *Farm) CanWaterAt() string {
	nextTime := f.LastWateredAt.Add(time.Hour * config.CONFIG.Farm.WaterCooldown).Unix()
	return fmt.Sprintf("<t:%d:R>", nextTime)
}

// Returns true if the user can harvest any of their crops
// unifinished
func (f *Farm) CanHarvest() bool {
	return true
}

// Functions waters every single plot
// Meaning it will update the plantedAt time
// Run QueryFarmPlots() before running this function
func (f *Farm) WaterPlots() {

	// Update last watered at
	f.LastWateredAt = time.Now()
	// A change was made so it needs to be saved when farm Save function is called
	f.PlotsChanged = true

	for _, plot := range f.Plots {
		plot.Water()
	}
}

type harvestResult struct {
	Name    string
	Emoji   string
	Earning int
}

// Returns an array containing the crop object that was harvested
// Money earned is saved in f.HarvestEarnings. Remember to add it to the user's balance
// Run QueryFarmPlots() before running this function
func (f *Farm) HarvestPlots() []harvestResult {

	var result []harvestResult

	for _, plot := range f.Plots {

		plot.QueryCropInfo()

		if !plot.HasFullyGrown() {
			continue // Not fully grown, so skip
		}

		result = append(result,
			harvestResult{
				Name:    plot.Crop.Name,
				Emoji:   plot.Crop.Emoji,
				Earning: plot.Crop.HarvestReward,
			})

		f.HarvestEarnings += plot.Crop.HarvestReward

		// Delete from the database
		defer f.DeletePlot(plot) // We cannot delete it at once, because we are iterating over it
	}

	return result
}

func (f *Farm) SuccessfulHarvest() bool {
	return f.HarvestEarnings > 0
}

// Will return the amount of crops that have perished from plots
// Will also remove perished crop plots from the database
// Remember to have call GetFarmPlots() before calling this function
func (f *Farm) CropsPerishedCheck() []string {

	// Checks if the user has watered their crops in the last x hours
	if float64(config.CONFIG.Farm.CropsPreishAfter) > time.Since(f.LastWateredAt).Hours() {
		return []string{}
	}

	// User has not watered in the last x hours. Missing the set deadline
	// Crops not fully grown (ready to be harvested) will perish

	var perishedCrops []string

	for _, plot := range f.Plots {

		plot.QueryCropInfo()

		if plot.HasFullyGrown() {
			// Crop has fully grown so we need to check if the user didn't
			// just wait the whole growth duration without watering it

			crop := &plot.Crop
			// Calc the time when the crop will be fully grown
			fullyGrownAt := plot.PlantedAt.Add(crop.DurationToGrow)

			// Calc the time when the crop would have had to be watered
			waterDeadline := fullyGrownAt.Add(time.Hour * config.CONFIG.Farm.CropsPreishAfter * -1)

			// If the last time the plant was watered is after the deadline
			// Then the crop is ok, lse it will perish
			if f.LastWateredAt.After(waterDeadline) {
				// The crop is ok
				continue
			}
		}

		perishedCrops = append(perishedCrops, plot.Crop.Name)
		defer f.DeletePlot(plot) // We cannot delete it at once, because we are iterating over it
	}

	return perishedCrops
}

func (f *Farm) DeletePlot(plot *FarmPlot) {

	f.PlotsChanged = true
	// Remove plot from f.Plots
	for i, p := range f.Plots {
		if p.ID == plot.ID {
			f.Plots = append(f.Plots[:i], f.Plots[i+1:]...)
			break
		}
	}

	plot.DeleteFromDB()
}

// Returns the cost of what buying a new plot would cost for the user
func (f *Farm) CalcFarmPlotPrice() int {

	// floor (Base price * multiplier ^ (number of plots - 1))

	return int(math.Floor(
		float64(config.CONFIG.Farm.FarmPlotPrice) * math.Pow(
			config.CONFIG.Farm.FarmPlotCostMultiplier,
			float64(f.OwnedPlots-1))))
}
