package database

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	DiscordID string `gorm:"uniqueIndex"`
	Money     uint64
	Work      Work  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Daily     Daily `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

type Work struct {
	gorm.Model
	LastWorkedAt time.Time
	Streak       uint16
	Tools        uint8
}

type Daily struct {
	gorm.Model
	LastDailyAt time.Time
	Streak      uint16
}
