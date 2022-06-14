package database

import "time"

type Notify struct {
	Model
	User     User      `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Message  string    `gorm:"not null"`
	NotifyAt time.Time `gorm:"not null"`
}

func (Notify) TableName() string {
	return "notify"
}
