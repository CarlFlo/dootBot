package database

import (
	"github.com/CarlFlo/malm"
	"gorm.io/gorm"
)

type YoutubeCache struct {
	Model
	VideoID     string `gorm:"uniqueIndex"`
	Title       string `gorm:"not null"`
	Thumbnail   string `gorm:"not null"`
	ChannelName string `gorm:"not null"`
	Duration    string `gorm:"not null"`
}

func (YoutubeCache) TableName() string {
	return "youtubeCache"
}

func (c *YoutubeCache) AfterCreate(tx *gorm.DB) error {
	return nil
}

// Saves the data to the database
func (c *YoutubeCache) Save() {
	DB.Save(&c)
}

// Check checks if the videoID exists in the cache
// Populates the values if the video is found
// Returns true if it exists
func (c *YoutubeCache) Check(videoID string, title, thumbnail, channelName, duration *string) bool {
	if err := DB.Table("youtubeCache").Where("video_id = ?", videoID).First(c).Error; err != nil {
		// Not found, or error.
		return false
	}

	*title = c.Title
	*thumbnail = c.Thumbnail
	*channelName = c.ChannelName
	*duration = c.Duration

	return true
}

func (c *YoutubeCache) Cache(videoID, title, thumbnail, channelName, duration string) {

	c = &YoutubeCache{
		VideoID:     videoID,
		Title:       title,
		Thumbnail:   thumbnail,
		ChannelName: channelName,
		Duration:    duration,
	}

	if err := DB.Create(c).Error; err != nil {
		malm.Error("%s", err)
	}
}
