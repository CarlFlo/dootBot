package database

import (
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/config"
	"gorm.io/gorm"
)

type FarmPlot struct {
	gorm.Model
	FarmID    uint `gorm:"index"`
	Farm      Farm `gorm:"references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // The farm this plot belongs to
	CropID    int
	Crop      FarmCrop  `gorm:"references:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // The planted crop
	PlantedAt time.Time // When the user planted the crop
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

// Broken
func (fp *FarmPlot) GetCropInfo() FarmCrop {
	var fc FarmCrop
	DB.Raw("SELECT * FROM farmCrops WHERE farmCrops.ID = ?", fp.CropID).First(&fc)
	return fc
}

// Wateres the plot by updating the PlantedAt time
func (fp *FarmPlot) Water() {

	// The plantedAt time is moved back
	fp.PlantedAt = fp.PlantedAt.Add(time.Hour * config.CONFIG.Farm.WaterCropTimeReductionHours * -1)
}
