package music

import (
	"fmt"
	"math"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func CreateMusicOverviewMessage(channelID string, i interface{}) {
	guildID, err := utils.GetGuild(channelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err)
		return
	}

	vi := instances[guildID]
	if vi == nil {
		malm.Error("No music instance found for guild '%s' when creating a message", guildID)
		return
	}

	switch msg := i.(type) {
	case *discordgo.MessageSend:
		msg.Embeds = createEmbeds(vi)
		messageComponents(vi, &msg.Components)
	case *discordgo.MessageEdit:
		embeds := createEmbeds(vi)
		msg.Embeds = &embeds

		components := []discordgo.MessageComponent{}
		if msg.Components != nil {
			components = (*msg.Components)[:0]
		}
		messageComponents(vi, &components)
		msg.Components = &components
	default:
		malm.Fatal("Unknown message type when creating a message")
	}
}

func createEmbeds(vi *VoiceInstance) []*discordgo.MessageEmbed {
	title, description, url := messageTitleAndDescription(vi)

	return []*discordgo.MessageEmbed{
		{
			Title:       title,
			URL:         url,
			Description: description,
			Color:       config.CONFIG.Colors.Neutral,
			Thumbnail:   messageThumbnail(vi),
			Author:      messageAuthor(),
			Footer:      messageFooter(vi),
		},
	}
}

func messageComponents(vi *VoiceInstance, c *[]discordgo.MessageComponent) {
	buttonRow := discordgo.ActionsRow{}

	if vi.IsLoading() {
		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    "Loading...",
			CustomID: "-",
			Disabled: true,
			Style:    2,
		})
	} else {
		playOrPauseLabel := "Pause"
		if vi.IsPaused() {
			playOrPauseLabel = "Play"
		}

		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    playOrPauseLabel,
			CustomID: "toggleSong",
			Style:    3,
		})

		previousOrRestart := "Previous"
		if vi.IsStartOfQueue() {
			previousOrRestart = "Restart"
		}

		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    previousOrRestart,
			CustomID: "prevSong",
			Style:    2,
		})

		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    "Exit",
			CustomID: "stopSong",
			Style:    4,
		})

		loopLabel := "Loop (is off)"
		if vi.IsLooping() {
			loopLabel = "Loop (is on)"
		}

		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    loopLabel,
			CustomID: "loopSong",
			Style:    1,
		})
	}

	if vi.GetQueueLengthRelative() > 1 {
		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    "Next",
			CustomID: "nextSong",
			Style:    1,
		})
	}

	*c = append(*c, buttonRow)
}

func messageAuthor() *discordgo.MessageEmbedAuthor {
	return &discordgo.MessageEmbedAuthor{Name: "Music Player"}
}

func messageTitleAndDescription(vi *VoiceInstance) (string, string, string) {
	song, err := vi.GetFirstInQueue()
	if err != nil {
		return "Nothing to play", "", ""
	}

	description := ""
	upNextSongsToDisplay := 3

	if vi.GetQueueLengthRelative() > 1 {
		description = "**Up Next**"

		upTo := int(math.Min(
			float64(vi.GetQueueLength()),
			float64(vi.GetQueueIndex()+upNextSongsToDisplay+1),
		))

		for i := vi.GetQueueIndex() + 1; i < upTo; i++ {
			nextSong := vi.GetSongByIndex(i)
			if nextSong == nil {
				continue
			}
			description = fmt.Sprintf("%s\n%s", description, nextSong.Title)
		}
	}

	return song.Title, description, song.GetYoutubeURL()
}

func messageThumbnail(vi *VoiceInstance) *discordgo.MessageEmbedThumbnail {
	song, err := vi.GetFirstInQueue()
	if err != nil || song.Thumbnail == "" {
		return nil
	}

	return &discordgo.MessageEmbedThumbnail{URL: song.Thumbnail}
}

func messageFooter(vi *VoiceInstance) *discordgo.MessageEmbedFooter {
	length := vi.GetQueueLengthRelative() - 1
	text := ""

	if length > 0 {
		text = fmt.Sprintf("%d songs in the queue", length)
		if length == 1 {
			text = "1 song in the queue"
		}
		text += "\n"
	}

	song, err := vi.GetFirstInQueue()
	if err == nil {
		text += fmt.Sprintf("Current song duration: %s", song.GetDuration())
	}

	return &discordgo.MessageEmbedFooter{Text: text}
}
