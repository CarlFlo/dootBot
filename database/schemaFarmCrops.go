package database

import (
	"time"

	"gorm.io/gorm"
)

type FarmCrop struct {
	gorm.Model
	Name           string
	Emoji          string
	DurationToGrow time.Duration
	HarvestReward  int
}

func (FarmCrop) TableName() string {
	return "farmCrops"
}
