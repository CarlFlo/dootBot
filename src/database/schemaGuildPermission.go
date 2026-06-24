package database

import (
	"errors"

	"gorm.io/gorm"
)

var errGuildPermissionNotFound = errors.New("guild permission not found")

type GuildPermissionAssignment struct {
	Model
	GuildID   string `gorm:"uniqueIndex:idx_guild_permission_assignment;not null"`
	DiscordID string `gorm:"uniqueIndex:idx_guild_permission_assignment;not null"`
	Level     uint8  `gorm:"not null"`
}

func (GuildPermissionAssignment) TableName() string {
	return "guild_permission_assignments"
}

type GuildPermissionRole struct {
	Model
	GuildID string `gorm:"uniqueIndex:idx_guild_permission_role;not null"`
	RoleID  string `gorm:"uniqueIndex:idx_guild_permission_role;not null"`
	Level   uint8  `gorm:"not null"`
}

func (GuildPermissionRole) TableName() string {
	return "guild_permission_roles"
}

func SetGuildUserPermission(guildID, discordID string, level uint8) error {
	assignment := &GuildPermissionAssignment{
		GuildID:   guildID,
		DiscordID: discordID,
	}

	return DB.Where(GuildPermissionAssignment{
		GuildID:   guildID,
		DiscordID: discordID,
	}).Assign(GuildPermissionAssignment{Level: level}).FirstOrCreate(assignment).Error
}

func RemoveGuildUserPermission(guildID, discordID string) error {
	result := DB.Where("guild_id = ? AND discord_id = ?", guildID, discordID).Delete(&GuildPermissionAssignment{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errGuildPermissionNotFound
	}
	return nil
}

func ListGuildUserPermissions(guildID string) ([]GuildPermissionAssignment, error) {
	assignments := []GuildPermissionAssignment{}
	if err := DB.Where("guild_id = ?", guildID).Order("level desc, discord_id asc").Find(&assignments).Error; err != nil {
		return nil, err
	}
	return assignments, nil
}

func SetGuildRolePermission(guildID, roleID string, level uint8) error {
	link := &GuildPermissionRole{
		GuildID: guildID,
		RoleID:  roleID,
	}

	return DB.Where(GuildPermissionRole{
		GuildID: guildID,
		RoleID:  roleID,
	}).Assign(GuildPermissionRole{Level: level}).FirstOrCreate(link).Error
}

func RemoveGuildRolePermission(guildID, roleID string) error {
	result := DB.Where("guild_id = ? AND role_id = ?", guildID, roleID).Delete(&GuildPermissionRole{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errGuildPermissionNotFound
	}
	return nil
}

func ListGuildRolePermissions(guildID string) ([]GuildPermissionRole, error) {
	links := []GuildPermissionRole{}
	if err := DB.Where("guild_id = ?", guildID).Order("level desc, role_id asc").Find(&links).Error; err != nil {
		return nil, err
	}
	return links, nil
}

func ResolveGuildPermissionLevel(guildID, discordID string, roleIDs []string) (uint8, uint8, error) {
	var assignment GuildPermissionAssignment
	userLevel := uint8(0)
	if err := DB.Where("guild_id = ? AND discord_id = ?", guildID, discordID).First(&assignment).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, 0, err
		}
	} else {
		userLevel = assignment.Level
	}

	roleLevel := uint8(0)
	if len(roleIDs) == 0 {
		return userLevel, roleLevel, nil
	}

	var maxRoleLevel struct {
		Level uint8
	}
	if err := DB.Model(&GuildPermissionRole{}).
		Select("COALESCE(MAX(level), 0) AS level").
		Where("guild_id = ? AND role_id IN ?", guildID, roleIDs).
		Scan(&maxRoleLevel).Error; err != nil {
		return 0, 0, err
	}

	return userLevel, maxRoleLevel.Level, nil
}

func IsGuildPermissionNotFound(err error) bool {
	return errors.Is(err, errGuildPermissionNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
