package database

import (
	"time"

	"gorm.io/gorm"
)

type FarmCrop struct {
	gorm.Model
	Name           string
	Emoji          string
	DurationToGrow time.Time
	HarvestReward  int
}

func (FarmCrop) TableName() string {
	return "farmCrops"
}
