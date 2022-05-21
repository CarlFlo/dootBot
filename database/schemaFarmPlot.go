package database

import (
	"fmt"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"gorm.io/gorm"
)

type FarmPlot struct {
	Model
	FarmID    uint `gorm:"index"`
	Farm      Farm `gorm:"references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // The farm this plot belongs to
	CropID    int
	Crop      FarmCrop  `gorm:"references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // The planted crop
	PlantedAt time.Time // When the user planted the crop
	Perished  bool      // Perished crops wont yeild any money
}

func (FarmPlot) TableName() string {
	return "userFarmPlots"
}

func (fp *FarmPlot) BeforeCreate(tx *gorm.DB) error {

	fp.PlantedAt = time.Now()
	return nil
}

func (fp *FarmPlot) AfterCreate(tx *gorm.DB) error {
	// Save to debug
	return nil
}

// Saves the data to the database
func (fp *FarmPlot) Save() {
	DB.Save(&fp)
}

// Removes the entry from the database
func (fp *FarmPlot) DeleteFromDB() {

	DB.Delete(&fp)
}

// Will mark the crop as perished and save it to the database
func (fp *FarmPlot) Perish() {
	fp.Perished = true
}

func (fp *FarmPlot) QueryCropInfo() {

	DB.Raw("SELECT * FROM farmCrops WHERE farmCrops.ID = ?", fp.CropID).First(&fp.Crop)
}

// Check if the crop has fully grown by comparing the duration to grow
// with the planted at time and the current time
// Call QueryCropInfo() first
func (fp *FarmPlot) HasFullyGrown() bool {
	return time.Now().After(fp.PlantedAt.Add(fp.Crop.DurationToGrow))
}

func (fp *FarmPlot) HasPerished() bool {
	return fp.Perished
}

// Wateres the plot by updating the PlantedAt time
func (fp *FarmPlot) Water() {

	// The plantedAt time is moved back - Not moving back the time the set amount for some reason
	fp.PlantedAt = fp.PlantedAt.Add(time.Hour * config.CONFIG.Farm.WaterCropTimeReductionHours * -1)
}

// Returns a discord formatted string showing when the crop will be harvestable
func (fp *FarmPlot) HarvestableAt() string {

	if fp.Perished {
		return "``Plant has perished``"
	}

	fullyGrown := fp.PlantedAt.Add(fp.Crop.DurationToGrow)

	// time after
	if time.Now().After(fullyGrown) {
		return "Now!"
	}

	return fmt.Sprintf("<t:%d:R>", fullyGrown.Unix())
}
