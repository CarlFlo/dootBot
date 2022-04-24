package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
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

func (f *Farm) AfterCreate(tx *gorm.DB) error {

	// initialize the plots

	/*
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
	*/
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
		f.ID = u.ID
		f.OwnedPlots = config.CONFIG.Farm.DefaultOwnedFarmPlots // Create entry in database
	}
}

func (f *Farm) GetFarmPlots() {
	DB.Raw("SELECT * FROM userFarmPlots WHERE userFarmPlots.Farm = ? LIMIT ?", f.ID, f.OwnedPlots).Find(f.Plots)
}

// Rember to run GetFarmPlots() before running this function
func (f *Farm) GetUnusedPlots() int {
	return int(f.OwnedPlots) - len(f.Plots)
}
