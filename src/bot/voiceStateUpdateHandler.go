package bot

import (
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func voiceStateUpdateHandler(s *discordgo.Session, vs *discordgo.VoiceStateUpdate) {

	malm.Info("[%s : %s] VoiceStateUpdate event for '%s'", vs.GuildID, vs.ChannelID, vs.UserID)

}
