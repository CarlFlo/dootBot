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
func CreateMusicOverviewMessage(channelID string, i interface{}) {

	guildID, err := utils.GetGuild(channelID)
	if err != nil {
		malm.Error("Error getting guild ID: %s", err.Error())
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
		msg.Embeds = createEmbeds(vi)
		messageComponents(vi, &msg.Components)

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
			Fields:      messageCreateFields(vi),
			Thumbnail:   messageThumbnail(vi),
			Author:      messageAuthor(vi),
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
			Style:    2, // Gray
		})

	} else {
		playOrPaused := "Pause"
		if !vi.IsPlaying() {
			playOrPaused = "Play"
		}

		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    playOrPaused,
			CustomID: "toggleSong",
			Style:    3, // Green
		})

		buttonRow.Components = append(buttonRow.Components, discordgo.Button{
			Label:    "Stop",
			CustomID: "stopSong",
			Style:    4, // Red
		})
	}

	if vi.GetQueueLength() > 1 {
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

	if vi.GetQueueLength() > 1 {
		description += "\nQueue:"
	}

	return title, description, song.GetYoutubeURL()
}

func messageCreateFields(vi *VoiceInstance) []*discordgo.MessageEmbedField {

	var fields []*discordgo.MessageEmbedField

	// Only show 3 song previews from the queue
	upTo := int(math.Min(float64(vi.GetQueueLength()), float64(4)))

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

	length := vi.GetQueueLength() - 1

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
