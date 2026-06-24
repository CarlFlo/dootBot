package bot

import (
	"github.com/CarlFlo/dootBot/src/bot/music"
	"github.com/bwmarrin/discordgo"
)

func voiceStateUpdateHandler(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {
	if vs == nil || vs.VoiceState == nil {
		return
	}

	//malm.Info("[%s : %s] VoiceStateUpdate event for '%s'", vs.GuildID, vs.ChannelID, vs.UserID)
	music.ForwardVoiceStateUpdate(vs)
}
