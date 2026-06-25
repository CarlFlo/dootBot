package database

import (
	"errors"

	"gorm.io/gorm"
)

var errGuildPermissionNotFound = errors.New("guild permission not found")

// Legacy user-grant table retained only so existing databases can be migrated away from it.
type GuildPermissionAssignment struct {
	Model
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

type GuildPermissionSettings struct {
	Model
	GuildID             string `gorm:"uniqueIndex;not null"`
	OpenRequestsEnabled bool   `gorm:"not null;default:false"`
}

func (GuildPermissionSettings) TableName() string {
	return "guild_permission_settings"
}

type GuildMusicChannel struct {
	Model
	GuildID   string `gorm:"uniqueIndex:idx_guild_music_channel;not null"`
	ChannelID string `gorm:"uniqueIndex:idx_guild_music_channel;not null"`
}

func (GuildMusicChannel) TableName() string {
	return "guild_music_channels"
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

func ResolveGuildPermissionLevel(guildID string, roleIDs []string) (uint8, bool, error) {
	roleLevel := uint8(0)
	if len(roleIDs) > 0 {
		var maxRoleLevel struct {
			Level uint8
		}
		if err := DB.Model(&GuildPermissionRole{}).
			Select("COALESCE(MAX(level), 0) AS level").
			Where("guild_id = ? AND role_id IN ?", guildID, roleIDs).
			Scan(&maxRoleLevel).Error; err != nil {
			return 0, false, err
		}
		roleLevel = maxRoleLevel.Level
	}

	settings := GuildPermissionSettings{}
	if err := DB.Where("guild_id = ?", guildID).First(&settings).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, false, err
		}
		return roleLevel, false, nil
	}

	return roleLevel, settings.OpenRequestsEnabled, nil
}

func GetGuildPermissionSettings(guildID string) (GuildPermissionSettings, error) {
	settings := GuildPermissionSettings{}
	if err := DB.Where("guild_id = ?", guildID).First(&settings).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return GuildPermissionSettings{GuildID: guildID}, nil
		}
		return GuildPermissionSettings{}, err
	}
	return settings, nil
}

func SetGuildOpenRequestsEnabled(guildID string, enabled bool) error {
	settings := &GuildPermissionSettings{
		GuildID: guildID,
	}

	return DB.Where(GuildPermissionSettings{GuildID: guildID}).
		Assign(GuildPermissionSettings{OpenRequestsEnabled: enabled}).
		FirstOrCreate(settings).Error
}

func ToggleGuildOpenRequestsEnabled(guildID string) (bool, error) {
	settings, err := GetGuildPermissionSettings(guildID)
	if err != nil {
		return false, err
	}

	next := !settings.OpenRequestsEnabled
	if err := SetGuildOpenRequestsEnabled(guildID, next); err != nil {
		return false, err
	}

	return next, nil
}

func BindGuildMusicChannel(guildID, channelID string) error {
	binding := &GuildMusicChannel{
		GuildID:   guildID,
		ChannelID: channelID,
	}

	return DB.Where(GuildMusicChannel{
		GuildID:   guildID,
		ChannelID: channelID,
	}).FirstOrCreate(binding).Error
}

func UnbindGuildMusicChannel(guildID, channelID string) error {
	result := DB.Where("guild_id = ? AND channel_id = ?", guildID, channelID).Delete(&GuildMusicChannel{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return errGuildPermissionNotFound
	}
	return nil
}

func ClearGuildMusicChannels(guildID string) error {
	return DB.Where("guild_id = ?", guildID).Delete(&GuildMusicChannel{}).Error
}

func ListGuildMusicChannels(guildID string) ([]GuildMusicChannel, error) {
	channels := []GuildMusicChannel{}
	if err := DB.Where("guild_id = ?", guildID).Order("channel_id asc").Find(&channels).Error; err != nil {
		return nil, err
	}
	return channels, nil
}

func IsGuildMusicChannelAllowed(guildID, channelID string) (bool, error) {
	var count int64
	if err := DB.Model(&GuildMusicChannel{}).Where("guild_id = ?", guildID).Count(&count).Error; err != nil {
		return false, err
	}
	if count == 0 {
		return true, nil
	}

	var matchCount int64
	if err := DB.Model(&GuildMusicChannel{}).
		Where("guild_id = ? AND channel_id = ?", guildID, channelID).
		Count(&matchCount).Error; err != nil {
		return false, err
	}

	return matchCount > 0, nil
}

func IsGuildPermissionNotFound(err error) bool {
	return errors.Is(err, errGuildPermissionNotFound) || errors.Is(err, gorm.ErrRecordNotFound)
}
