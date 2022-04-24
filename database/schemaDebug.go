package database

import (
	"gorm.io/gorm"
)

type Debug struct {
	gorm.Model
	DailyCount uint64
	WorkCount  uint64
}

func (Debug) TableName() string {
	return "debug"
}
