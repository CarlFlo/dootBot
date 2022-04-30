package database

import "time"

// This is the basic model that should be used for all tables

type Model struct {
	ID        uint `gorm:"primarykey"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
