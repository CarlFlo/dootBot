package bot

import (
	"github.com/CarlFlo/dootBot/src/bot/music"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func voiceServerUpdateHandler(s *discordgo.Session, vs *discordgo.VoiceServerUpdate) {
	if vs == nil {
		return
	}

	malm.Info("[%s] VoiceServerUpdate event", vs.GuildID)
	music.ForwardVoiceServerUpdate(vs)
}
