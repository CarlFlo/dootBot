package music

import (
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/bwmarrin/discordgo"
)

// GuildID is the key
var instances = map[string]*voiceInstance{}

/*
	Play songs in a voice channel
	Commands:
	play (plays a song or adds the song to the queue if something is playing), resume, skip, stop, pause, playlist (ability to create a personal playlist, adds songs with buttons etc)

	playlist: dropdown menu with selections of playlists in the guild

	Save stats in DB for songs played, skiped
	Only save:

	https://www.youtube.com/watch?v=5qap5aO4i9A -> 5qap5aO4i9A
	To save storage, in DB
*/

func PlayMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	if input.NumberOfArgsAreAtleast(1) {
		// Nothing to play
		return
	}

	//guildID := utils.GetGuild(s, m)
	//instance := voiceInstances[guildID]

}

func ResumeMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}

func StopMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}

func SkipMusic(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

}
