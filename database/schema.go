package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	DiscordID string `gorm:"uniqueIndex"`
	Money     uint64
	Work      Work `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Work struct {
	gorm.Model
	LastUpdated time.Time
	Streak      uint16
	Tools       uint8
}

/*
type Work struct {
	gorm.Model
	UserID      uint
	LastUpdated time.Time
	Streak      uint16
	Tools       uint8
}

*/
