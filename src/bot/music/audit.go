package music

import (
	"fmt"

	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/malm"
)

const (
	auditActionPlay     = "played"
	auditActionAdd      = "added"
	auditActionPause    = "paused"
	auditActionResume   = "resumed"
	auditActionSkip     = "skipped"
	auditActionStop     = "stopped"
	auditActionClear    = "cleared queue"
	auditActionLoopOn   = "enabled loop"
	auditActionLoopOff  = "disabled loop"
	auditActionRestart  = "restarted"
	auditActionPrevious = "previous"
)

func logMusicAudit(guildID, userID, action, description string, song *Song) {
	if guildID == "" || userID == "" {
		return
	}

	var auditSong *database.MusicAuditSong
	if song != nil {
		auditSong = &database.MusicAuditSong{
			Title:  song.Title,
			URL:    song.URL,
			Author: song.ChannelName,
		}
	}

	if err := database.CreateMusicAuditLog(guildID, userID, action, description, auditSong); err != nil {
		malm.Error("unable to create music audit log: %s", err)
	}
}

func currentSongForAudit(vi *VoiceInstance) *Song {
	if vi == nil {
		return nil
	}

	song, err := vi.GetFirstInQueue()
	if err != nil {
		return nil
	}

	return song
}

func queuedSongsAfterCurrent(vi *VoiceInstance) int {
	if vi == nil {
		return 0
	}

	remaining := vi.GetQueueLength() - vi.GetQueueIndex() - 1
	if remaining < 0 {
		return 0
	}

	return remaining
}

func previousActionLabel(vi *VoiceInstance) string {
	if vi != nil && vi.shouldGoToPreviousSong() {
		return auditActionPrevious
	}

	return auditActionRestart
}

func clearQueueDescription(removed int) string {
	if removed <= 0 {
		return "No queued songs were removed"
	}
	if removed == 1 {
		return "Removed 1 queued song"
	}

	return fmt.Sprintf("Removed %d queued songs", removed)
}

func stopDescription(removed int) string {
	if removed <= 0 {
		return "Stopped playback"
	}
	if removed == 1 {
		return "Stopped playback and canceled 1 queued song"
	}

	return fmt.Sprintf("Stopped playback and canceled %d queued songs", removed)
}
