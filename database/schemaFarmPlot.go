package database

import (
	"time"

	"gorm.io/gorm"
)

type FarmPlot struct {
	gorm.Model
	Farm    Farm      `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // The farm this plot belongs to
	Crop    FarmCrop  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"` // The planted crop
	Planted time.Time // When the user planted the crop
}

func (FarmPlot) TableName() string {
	return "userFarmPlotData"
}
