package music

import (
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/config"
)

type songStreamCacheWrapper struct {
	mu        sync.Mutex
	songCache map[string]songStreamCache
}

type songStreamCache struct {
	streamURL string
	expires   time.Time
}

var songCache = songStreamCacheWrapper{
	songCache: make(map[string]songStreamCache),
}

// Adding a duplicate will overwrite the old one
func (c *songStreamCacheWrapper) Add(song *Song) {

	c.mu.Lock()
	defer c.mu.Unlock()

	c.songCache[song.YoutubeVideoID] = songStreamCache{
		streamURL: song.StreamURL,
		expires:   time.Now().Add(time.Minute * config.CONFIG.Music.MaxCacheAgeMin), // Valid for 90 minutes, 1h 30 min
	}
}

func (c *songStreamCacheWrapper) Check(ytURL string) string {

	ssc := c.songCache[ytURL]
	if time.Now().After(ssc.expires) {
		// remove from map
		c.mu.Lock()
		defer c.mu.Unlock()
		delete(c.songCache, ytURL)
		return ""
	}
	return ssc.streamURL
}
