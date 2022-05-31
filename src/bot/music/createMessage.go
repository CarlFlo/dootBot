package music

import (
	"fmt"
	"math"

	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

/*	Message idea
Music Player
[Prev] > Heat Waves - Oliver Heldens
[Playing] > Colors - METAHESH
[Next] > Young RIght Now - Robin Schulz, Dennis Lloyd
[Next] > ...

> The name of the last user that pressed a button (So we can tell who paused or stopped)
 *Buttons* *Buttons* *Buttons*
*/

// CreateMusicOverviewMessage creates the music overview message
func CreateMusicOverviewMessage(s *discordgo.Session, m *discordgo.MessageCreate, ms *discordgo.MessageSend, me *discordgo.MessageEdit) {

	guildID := utils.GetGuild(s, m)
	vi := instances[guildID]

	if vi == nil {
		malm.Error("No music instance found for guild '%s' when creating a message", guildID)
		return
	}

	title, description, url := messageTitleAndDescription(vi)

	// Received a message send
	if ms != nil {

		ms.Embeds = []*discordgo.MessageEmbed{
			{
				Title:       title,
				URL:         url,
				Description: description,
				Color:       config.CONFIG.Colors.Neutral,
				Fields:      messageCreateFields(vi),
				Thumbnail:   messageThumbnail(vi),
				Author:      messageAuthor(vi),
				Footer:      messageFooter(vi),
			},
		}

		messageComponents(vi, &ms.Components)
		return
	}

	// Received a message edit
	me.Embeds = []*discordgo.MessageEmbed{
		{
			Title:       title,
			URL:         url,
			Description: description,
			Color:       config.CONFIG.Colors.Neutral,
			Fields:      messageCreateFields(vi),
			Thumbnail:   messageThumbnail(vi),
			Author:      messageAuthor(vi),
			Footer:      messageFooter(vi),
		},
	}
	messageComponents(vi, &me.Components)
}

func messageComponents(vi *VoiceInstance, c *[]discordgo.MessageComponent) {

	playOrPaused := "Pause"
	if !vi.IsPlaying() {
		playOrPaused = "Play"
	}

	buttonRow := discordgo.ActionsRow{
		Components: []discordgo.MessageComponent{
			discordgo.Button{
				Label:    playOrPaused,
				CustomID: "toggleSong",
				Style:    3, // Green
			},
			discordgo.Button{
				Label:    "Stop",
				CustomID: "stopSong",
				Style:    4, // Red
			},
		},
	}

	if vi.QueueLength() > 1 {
		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    "Clear queue",
			CustomID: "clearQueue",
			Style:    2, // Gray
		})

		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    "Next",
			CustomID: "nextSong",
			Style:    1, // Default 'blurple'
		})

	}

	// The buttons
	*c = append(*c, buttonRow)

	// #### Playlist menu ####
	//*c = append(*c, discordgo.ActionsRow{})
}

func messageAuthor(vi *VoiceInstance) *discordgo.MessageEmbedAuthor {

	/*
		var output bytes.Buffer

		output.WriteString(fmt.Sprintf("%s  ", config.CONFIG.Emojis.MusicNotes))

		if vi.IsPlaying() {
			output.WriteString(config.CONFIG.Emojis.MusicPlaying)
		} else {
			output.WriteString(config.CONFIG.Emojis.MusicPaused)
		}
	*/
	return &discordgo.MessageEmbedAuthor{
		Name: "Music Player",
	}
}

func messageTitleAndDescription(vi *VoiceInstance) (string, string, string) {

	var title string
	var description string

	song, err := vi.GetFirstInQueue()

	if err != nil {
		title = "Nothing to play"
		return title, description, ""
	}

	title = song.Title
	description = song.GetDuration()

	if vi.QueueLength() > 1 {
		description += "\nQueue:"
	}

	return title, description, song.GetYoutubeURL()
}

func messageCreateFields(vi *VoiceInstance) []*discordgo.MessageEmbedField {

	var fields []*discordgo.MessageEmbedField

	// Only show 3 song previews from the queue
	upTo := int(math.Min(float64(vi.QueueLength()), float64(4)))

	for i := 1; i < upTo; i++ {
		song := vi.GetSongByIndex(i)

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   song.Title,
			Value:  song.GetDuration(),
			Inline: true,
		})
	}
	return fields
}

func messageThumbnail(vi *VoiceInstance) *discordgo.MessageEmbedThumbnail {

	song, err := vi.GetFirstInQueue()

	if err != nil {
		return nil
	}

	return &discordgo.MessageEmbedThumbnail{
		URL: song.Thumbnail,
	}
}

func messageFooter(vi *VoiceInstance) *discordgo.MessageEmbedFooter {

	length := vi.QueueLength() - 1

	if length < 1 {
		return nil
	}

	text := fmt.Sprintf("%d songs in the queue", length)

	if length == 1 {
		text = fmt.Sprintf("%d song in the queue", length)
	}

	return &discordgo.MessageEmbedFooter{
		Text: text,
	}
}
