package database

import (
	"time"

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

func (fp *FarmPlot) UpdatePlotSlot(i int, crop *FarmCrop) {
	fp.Crop = *crop
	fp.PlantedAt = time.Now().UTC()
	fp.Save()
}
