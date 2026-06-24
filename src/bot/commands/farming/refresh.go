package farming

import (
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/bwmarrin/discordgo"
)

func RefreshInteraction(authorID string, author *discordgo.User, me *discordgo.MessageEdit) {
	var user database.User
	user.QueryUserByDiscordID(authorID)

	var farm database.Farm
	farm.QueryUserFarmData(&user)

	farm.UpdateInteractionOverview(author, me)
}
