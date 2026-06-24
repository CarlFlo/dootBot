package database

import (
	"errors"

	"gorm.io/gorm"
)

var errAdminNotFound = errors.New("admin not found")

type Admin struct {
	Model
	DiscordID string `gorm:"uniqueIndex;not null"`
}

func (Admin) TableName() string {
	return "admins"
}

func (a *Admin) Save() error {
	return DB.Save(a).Error
}

func (a *Admin) Exists(discordID string) bool {
	return DB.Where("discord_id = ?", discordID).First(a).Error == nil
}

func (a *Admin) DeleteByDiscordID(discordID string) error {
	result := DB.Where("discord_id = ?", discordID).Delete(&Admin{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errAdminNotFound
	}
	return nil
}

func GetAdmins() ([]Admin, error) {
	var admins []Admin
	if err := DB.Order("discord_id asc").Find(&admins).Error; err != nil {
		return nil, err
	}
	return admins, nil
}

func AddAdmin(discordID string) error {
	admin := &Admin{DiscordID: discordID}
	return DB.Where(Admin{DiscordID: discordID}).FirstOrCreate(admin).Error
}

func IsStoredAdmin(discordID string) bool {
	return DB.Where("discord_id = ?", discordID).First(&Admin{}).Error == nil
}

func RemoveAdmin(discordID string) error {
	result := DB.Where("discord_id = ?", discordID).Delete(&Admin{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errAdminNotFound
	}
	return nil
}

func IsAdminNotFound(err error) bool {
	return errors.Is(err, errAdminNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
