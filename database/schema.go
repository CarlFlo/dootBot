package database

import (
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	UserID string `gorm:"uniqueIndex"`
	Money  uint64
}
