package database

import (
	"time"

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
