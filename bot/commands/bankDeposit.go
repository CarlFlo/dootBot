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

	if len(input.GetArgs()) == 0 {
		s.ChannelMessageSend(m.ChannelID, "No amount specified!")
		return
	}

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

	oldUserMoney := user.PrettyPrintMoney()
	oldBankMoney := bank.PrettyPrintMoney()

	err = bank.Deposit(&user, uint64(amount))

	if err != nil {
		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Deposit failed! (%s)", err))
		return
	}

	footerText := fmt.Sprintf("While bank deposits are always instant, be advised that there is a %d %s withdrawal fee and that it can take upwards of %d hours to process the withdrawal!\nAccounts with a balance over %d %s will receive a daily interest rate of %.2f%%.", config.CONFIG.Bank.WithdrawFee, config.CONFIG.Economy.Name, config.CONFIG.Bank.MaxWithdrawWaitHours, config.CONFIG.Bank.MinAmountForInterest, config.CONFIG.Economy.Name, config.CONFIG.Bank.InterestRate*100)

	prettyAmount := utils.HumanReadableNumber(amount)

	complexMessage := &discordgo.MessageSend{Embeds: []*discordgo.MessageEmbed{
		&discordgo.MessageEmbed{
			Type:        discordgo.EmbedTypeRich,
			Color:       config.CONFIG.Colors.Success,
			Title:       fmt.Sprintf("%s %s %s", config.CONFIG.Emojis.Bank, config.CONFIG.Bank.Name, config.CONFIG.Emojis.Bank),
			Description: fmt.Sprintf("You have successfully deposited ``%s`` %s into your bank account", prettyAmount, config.CONFIG.Economy.Name),
			/*
				Author: &discordgo.MessageEmbedAuthor{
					Name:    fmt.Sprintf("%s#%s", m.Author.Username, m.Author.Discriminator),
					IconURL: m.Author.AvatarURL(""),
				},
			*/
			Fields: []*discordgo.MessageEmbedField{
				&discordgo.MessageEmbedField{
					Name:   "Wallet Funds",
					Value:  fmt.Sprintf("%s - %s (%s)", oldUserMoney, prettyAmount, user.PrettyPrintMoney()),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   fmt.Sprintf("%s %s %s", config.CONFIG.Emojis.Transfers, config.CONFIG.Emojis.Transfers, config.CONFIG.Emojis.Transfers),
					Value:  fmt.Sprintf("=(%s)=>", prettyAmount),
					Inline: true,
				},
				&discordgo.MessageEmbedField{
					Name:   "Bank Funds",
					Value:  fmt.Sprintf("%s + %s (%s)", oldBankMoney, prettyAmount, bank.PrettyPrintMoney()),
					Inline: true,
				},
			},
			Thumbnail: &discordgo.MessageEmbedThumbnail{
				URL: m.Author.AvatarURL("256"),
			},
			Footer: &discordgo.MessageEmbedFooter{
				Text: footerText,
			},
		},
	}}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! %s", err)
		return
	}
}
