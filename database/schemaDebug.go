package database

import (
	"gorm.io/gorm"
)

type Debug struct {
	gorm.Model
	SeenMessagesCounter uint64
}

func (Debug) TableName() string {
	return "debug"
}

func (d *Debug) LogMessage() {
	d.SeenMessagesCounter += 1
}
