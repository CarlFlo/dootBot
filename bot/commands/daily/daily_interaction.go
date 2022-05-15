package daily

import (
	"github.com/CarlFlo/DiscordMoneyBot/bot/commands"
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/bwmarrin/discordgo"
)

func DoDailyInteraction(authorID string, response *string, i *discordgo.Interaction, btnData *[]structs.ButtonData) {

	var user database.User
	user.QueryUserByDiscordID(authorID)

	var work database.Work
	work.GetWorkInfo(&user)

	var daily database.Daily
	daily.GetDailyInfo(&user)

	i.Message.Embeds[0].Fields = commands.GenerateProfileFields(&user, &work, &daily)

	*response = "Not implemented yet"

	*btnData = append(*btnData, structs.ButtonData{
		CustomID: "PD",
		Disabled: true,
	})
}
