package commands

import (
	"fmt"
	"strconv"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

func BankDeposit(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	// convert string to int
	amount, err := strconv.Atoi(input.GetArgs()[0])
	if err != nil {
		// Invalid input
		s.ChannelMessageSend(m.ChannelID, "Invalid deposit amount!")
		return
	} else if amount < 1 {
		// Zero or negative value
		s.ChannelMessageSend(m.ChannelID, "Cannot deposit zero or negative amounts!")
		return
	}

	var user database.User
	user.GetUserByDiscordID(m.Author.ID)

	var bank database.Bank
	bank.GetBankInfo(&user)

	err = bank.Deposit(&user, uint64(amount))

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Deposit failed! (%s)", err))
		return
	}

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type:  discordgo.EmbedTypeRich,
			Title: fmt.Sprintf("Bank Deposit to %s", config.CONFIG.Bank.Name),
			Color: config.CONFIG.Colors.Success,
			Author: &discordgo.MessageEmbedAuthor{
				Name:    fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator),
				IconURL: m.Author.AvatarURL(""),
			},
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Wallet funds",
					Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, user.PrettyPrintMoney()),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "",
					Value:  fmt.Sprintf("==(%s)=>", utils.HumanReadableNumber(amount)),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Bank funds",
					Value:  "",
					Inline: true,
				},
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: "Depositing money in the bank will earn you interest.",
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}

/*
const lib = require('lib')({token: process.env.STDLIB_SECRET_TOKEN});

await lib.discord.channels['@0.3.0'].messages.create({
  "channel_id": `${context.params.event.channel_id}`,
  "content": "",
  "tts": false,
  "embeds": [
    {
      "type": "rich",
      "title": `Bank Deposit to (BankName)`,
      "description": `You successfully deposited X amount into your bank account!`,
      "color": 0x00FFFF,
      "fields": [
        {
          "name": `Wallet Funds `,
          "value": `350 - 200 (150)`,
          "inline": true
        },
        {
          "name": "\u200B",
          "value": `==(200)=>`,
          "inline": true
        },
        {
          "name": `Bank Funds`,
          "value": `4000 + 200 (4200)`,
          "inline": true
        }
      ],
      "footer": {
        "text": `While bank deposits are always instant, be advised that bank withdrawal can take upwards of 48 hours to process and show up in your wallet!\\nA convenience fee of 100 is also deducted when you withdraw any amount.`
      }
    }
  ]
});
*/
