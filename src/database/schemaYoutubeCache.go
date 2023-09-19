package database

import (
	"fmt"
	"net/url"
	"strconv"
	"time"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/malm"
	"gorm.io/gorm"
)

type YoutubeCache struct {
	Model
	VideoID        string `gorm:"uniqueIndex"`
	Title          string `gorm:"not null"`
	Thumbnail      string `gorm:"not null"`
	ChannelName    string `gorm:"not null"`
	Duration       string `gorm:"not null"`
	URLCache       string
	URLCacheExpire time.Time
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
func (c *YoutubeCache) Check(videoID string, title, thumbnail, channelName, duration, streamURL *string) bool {
	if err := DB.Table("youtubeCache").Where("video_id = ?", videoID).First(c).Error; err != nil {
		// Not found, or error.
		return false
	}

	// Has cache URL expired?
	// Adds the max lenght min time to ensure the link does not expire when playing
	expireTime := c.URLCacheExpire.Add(time.Minute * config.CONFIG.Music.MaxSongLengthMinutes)

	// Expire time is not zero and now has not passed the expireTime
	if !c.URLCacheExpire.IsZero() && !time.Now().After(expireTime) {
		if len(c.URLCache) != 0 {
			*streamURL = c.URLCache
		}
	} else {
		// Cache is invalid or missing. Remove the old cache, or wait for it to be overwritten?
		// Will be overwritten automatically when the song is being fetched
	}

	*title = c.Title
	*thumbnail = c.Thumbnail
	*channelName = c.ChannelName
	*duration = c.Duration

	return true
}

func (c *YoutubeCache) Cache(videoID, title, thumbnail, channelName, duration string) {

	c = &YoutubeCache{
		VideoID:        videoID,
		Title:          title,
		Thumbnail:      thumbnail,
		ChannelName:    channelName,
		Duration:       duration,
		URLCacheExpire: time.Time{}, // CacheURL is added when the song object fetches the URL
	}

	if err := DB.Create(c).Error; err != nil {
		malm.Error("%s", err)
	}
}

func (c *YoutubeCache) UpdateStreamURL(videoID, streamURL string) error {

	expiresAt, err := getSongExpireTime(streamURL)

	if err != nil {
		return err
	}

	updates := struct {
		URLCache       string
		URLCacheExpire time.Time
	}{
		URLCache:       streamURL,
		URLCacheExpire: expiresAt,
	}

	// Update the table
	if err := DB.Table("youtubeCache").
		Where("video_id = ?", videoID).
		Updates(updates).Error; err != nil {
		return err
	}

	return nil
}

// Extracts when the URL will expire from the URL
// Check for nil with: '.IsZero()'
func getSongExpireTime(streamURL string) (time.Time, error) {
	parsedURL, err := url.Parse(streamURL)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse url. Reason: '%s'", err)
	}

	// Extract the "expire" query parameter
	expireValue := parsedURL.Query().Get("expire")

	if len(expireValue) == 0 {
		return time.Time{}, fmt.Errorf("unable to extract 'expire' value from URL: '%s'", streamURL)
	}

	expireValueNum, err := strconv.ParseInt(expireValue, 10, 64)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to convert 'expire' string to int64. Reason: '%s'", err)
	}

	timestamp := time.Unix(expireValueNum, 0)
	//formattedTime := timestamp.Format("2006-01-02 15:04:05 UTC")

	if timestamp.IsZero() {
		return time.Time{}, fmt.Errorf("parsed timestamp is zero value")
	}

	return timestamp, nil
}
