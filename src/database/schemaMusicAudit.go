package database

import "fmt"

const defaultMusicAuditPageSize = 10

type MusicAuditLog struct {
	Model
	GuildID     string `gorm:"index:idx_music_audit_guild_created;not null"`
	UserID      string `gorm:"index;not null"`
	Action      string `gorm:"size:64;not null"`
	SongTitle   string `gorm:"size:512"`
	SongURL     string `gorm:"size:1024"`
	SongAuthor  string `gorm:"size:255"`
	Description string `gorm:"size:1024"`
}

func (MusicAuditLog) TableName() string {
	return "music_audit_logs"
}

func CreateMusicAuditLog(guildID, userID, action, description string, song *MusicAuditSong) error {
	entry := MusicAuditLog{
		GuildID:     guildID,
		UserID:      userID,
		Action:      action,
		Description: description,
	}

	if song != nil {
		entry.SongTitle = song.Title
		entry.SongURL = song.URL
		entry.SongAuthor = song.Author
	}

	return DB.Create(&entry).Error
}

type MusicAuditSong struct {
	Title  string
	URL    string
	Author string
}

func ListMusicAuditLogs(guildID string, page, pageSize int) ([]MusicAuditLog, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = defaultMusicAuditPageSize
	}

	var total int64
	if err := DB.Model(&MusicAuditLog{}).Where("guild_id = ?", guildID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	logs := []MusicAuditLog{}
	offset := (page - 1) * pageSize
	if err := DB.
		Where("guild_id = ?", guildID).
		Order("created_at desc").
		Order("id desc").
		Limit(pageSize).
		Offset(offset).
		Find(&logs).Error; err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func MusicAuditTotalPages(total int64, pageSize int) int {
	if pageSize < 1 {
		pageSize = defaultMusicAuditPageSize
	}
	if total == 0 {
		return 1
	}

	return int((total + int64(pageSize) - 1) / int64(pageSize))
}

func FormatMusicAuditPageUsage(page int) string {
	return fmt.Sprintf("Use `audit %d` for the next page", page)
}
