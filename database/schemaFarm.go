package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/malm"
	"gorm.io/gorm"
)

type Farm struct {
	gorm.Model
	OwnedPlots  uint8
	LastWatered time.Time // Last time the user watered the farm plots
}

/*
	One farm per user, but you can have many plots
	The farm plot keeps track of which farm it belongs to
*/

func (Farm) TableName() string {
	return "userFarmData"
}

func (f *Farm) AfterCreate(tx *gorm.DB) error {

	// initialize the plots

	var crop FarmCrop
	crop.GetCropByName("Banana")

	for i := 0; i < int(config.CONFIG.Farm.DefaultOwnedFarmPlots); i++ {

		malm.Debug("Creating plot entry for '%s'", crop.Name)

		// Create a new plot
		var plot FarmPlot

		plot.Farm = *f
		plot.Crop = crop
		plot.Planted = time.Now().UTC()

		plot.Save()
	}

	// January 1st 1970
	f.LastWatered = time.Unix(0, 0).UTC()
	return nil
}

// Saves the data to the database
func (f *Farm) Save() {
	DB.Save(&f)
}

// Queries the database for the farm data with the given user object.
func (f *Farm) GetFarmInfo(u *User) {
	DB.Raw("SELECT * FROM userFarmData WHERE userFarmData.ID = ?", u.ID).First(&f)
	if f.ID == 0 { // Meaning there is not data so we initialize it
		f.ID = u.ID
		f.OwnedPlots = config.CONFIG.Farm.DefaultOwnedFarmPlots // Create entry in database
	}
}

func (f *Farm) GetFarmPlots() *[]FarmPlot {
	var plots []FarmPlot
	DB.Raw("SELECT * FROM userFarmPlotData WHERE userFarmPlotData.Farm = ? LIMIT ?", f.ID, f.OwnedPlots).Find(&plots)
	return &plots
}
