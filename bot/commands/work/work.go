package work

import (
	"fmt"
	"math/rand"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/config"
	"github.com/CarlFlo/DiscordMoneyBot/database"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

/* TODO
Make this and Daily functions into a wrapper
where you pass in the database object as pointers
if everything went ok then the returned error is nil
therefore we save the database object.

Which will look cleaner and be easier to understand
*/

func Work(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	var user database.User
	user.QueryUserByDiscordID(m.Author.ID)

	var work database.Work
	work.GetWorkInfo(&user)

	canWork := work.CanDoWork()

	// Reset streak if user hasn't worked in a specified amount of time (set in config)
	work.StreakPreMsgAction()

	// TODO: Rework this like how Daily was reworked

	complexMessage := &discordgo.MessageSend{
		Embeds: []*discordgo.MessageEmbed{
			{
				Title:       createWorkMessageTitle(&work, canWork),
				Description: createWorkMessageDescription(&user, &work, canWork),
				Color:       createWorkMessageColor(&work, canWork),
				Fields:      createWorkMessageFields(&work, canWork),
				Footer:      createWorkMessageFooter(&work, canWork),
				Thumbnail: &discordgo.MessageEmbedThumbnail{
					URL: fmt.Sprintf("%s#%s", m.Author.AvatarURL("256"), m.Author.ID),
				},
			},
		},
	}
	if components := createButtonComponent(&work); components != nil {
		complexMessage.Components = components
	}

	// Sends the message
	if _, err := s.ChannelMessageSendComplex(m.ChannelID, complexMessage); err != nil {
		malm.Error("Could not send message! (Data not saved) %s", err)
		return
	}

	work.StreakPostMsgAction()

	user.Save()
	work.Save()
}

// Returns the work title string
func createWorkMessageTitle(work *database.Work, canDoWork bool) string {

	if canDoWork {
		return "Pay Check"
	}
	return fmt.Sprintf("%s Slow down!", config.CONFIG.Emojis.Failure)
}

// generates the work description message
// Will also give the user money if they can work
func createWorkMessageDescription(user *database.User, work *database.Work, canDoWork bool) string {

	toolsTooltip := generateToolTooltip(work)

	var description string

	if canDoWork {

		// Calculates the income
		moneyEarned := generateWorkIncome(work)
		user.AddMoney(uint64(moneyEarned))

		moneyEarnedString := utils.HumanReadableNumber(moneyEarned)

		description = fmt.Sprintf("%sYou earned ``%s`` %s! Your new balance is ``%s`` %s!\nYou will be able to work again %s\nCurrent streak: ``%d``\n\n%s",
			config.CONFIG.Emojis.Economy,
			moneyEarnedString,
			config.CONFIG.Economy.Name,
			user.PrettyPrintMoney(),
			config.CONFIG.Economy.Name,
			work.CanDoWorkAt(),
			work.ConsecutiveStreaks,
			toolsTooltip)

	} else {
		description = fmt.Sprintf("You can work again %s\n\n%s", work.CanDoWorkAt(), toolsTooltip)
	}

	return description
}

func createWorkMessageColor(work *database.Work, canDoWork bool) int {

	if canDoWork {
		return config.CONFIG.Colors.Success
	}
	return config.CONFIG.Colors.Failure
}
func createWorkMessageFields(work *database.Work, canDoWork bool) []*discordgo.MessageEmbedField {

	extraRewardValue, percentage := generateWorkStreakMessage(work.Streak, canDoWork)

	return []*discordgo.MessageEmbedField{
		{
			Name:  fmt.Sprintf("Extra Reward Progress (%s)", percentage),
			Value: extraRewardValue,
		},
	}
}

func createWorkMessageFooter(work *database.Work, canDoWork bool) *discordgo.MessageEmbedFooter {

	footerText := fmt.Sprintf("You can work once every %d hours!", int(config.CONFIG.Work.Cooldown))

	if canDoWork {
		footerText = fmt.Sprintf("The streak resets after %d hours of inactivity and will reward %d %s on completion!\nEach tool you buy will earn you an additional %s %s when you work! (Max %d)",
			config.CONFIG.Work.StreakResetHours,
			config.CONFIG.Work.StreakBonus,
			config.CONFIG.Economy.Name,
			utils.HumanReadableNumber(config.CONFIG.Work.ToolBonus),
			config.CONFIG.Economy.Name,
			config.CONFIG.Work.MaxTools)
	}

	return &discordgo.MessageEmbedFooter{
		Text: footerText,
	}
}

func generateToolTooltip(work *database.Work) string {

	numOfBoughtTools := int(work.Tools)

	if numOfBoughtTools == 0 {
		return fmt.Sprintf("%s You have 0 tools", config.CONFIG.Emojis.Tools)
	}

	wordFormat := "tools"
	if numOfBoughtTools == 1 {
		wordFormat = "tool"
	}

	bonus := utils.HumanReadableNumber(numOfBoughtTools * config.CONFIG.Work.ToolBonus)
	return fmt.Sprintf("%s You have %d %s, giving you an additional **%s** %s",
		config.CONFIG.Emojis.Tools,
		numOfBoughtTools,
		wordFormat,
		bonus,
		config.CONFIG.Economy.Name)
}

func generateWorkStreakMessage(streak uint16, addStreakMessage bool) (string, string) {

	percentage := float64(streak) / float64(len(config.CONFIG.Work.StreakOutput))
	upTo := int(float64(len(config.CONFIG.Work.StreakOutput)) * percentage)

	// Append to a string values in config.CONFIG.Work.StreakOutput up to the index of upTo
	var visualStreakProgress string

	for i := 0; i < upTo; i++ {
		visualStreakProgress += fmt.Sprintf("%s ", config.CONFIG.Work.StreakOutput[i])
	}
	for i := upTo; i < len(config.CONFIG.Work.StreakOutput); i++ {
		visualStreakProgress += "- "
	}

	percentageText := fmt.Sprintf("%d%%", int(percentage*100))

	var streakMessage string
	if addStreakMessage && streak == uint16(len(config.CONFIG.Work.StreakOutput)) {
		streakMessage = fmt.Sprintf("An additional ``%s`` %s were added to your earnings!", utils.HumanReadableNumber(config.CONFIG.Work.StreakBonus), config.CONFIG.Economy.Name)
	}

	return fmt.Sprintf("%s %s", visualStreakProgress, streakMessage), percentageText
}

func generateWorkIncome(work *database.Work) int {

	// Generate a random int between config.CONFIG.Work.MinMoney and config.CONFIG.Work.MaxMoney
	moneyEarned := rand.Intn(config.CONFIG.Work.MaxMoney-config.CONFIG.Work.MinMoney) + config.CONFIG.Work.MinMoney

	// Adds the streak bonus to the amount
	if work.Streak == uint16(len(config.CONFIG.Work.StreakOutput)) {
		moneyEarned += config.CONFIG.Work.StreakBonus
	}

	moneyEarned += int(work.Tools) * config.CONFIG.Work.ToolBonus

	return moneyEarned
}

func createButtonComponent(work *database.Work) []discordgo.MessageComponent {

	components := []discordgo.MessageComponent{}

	_, priceString := work.CalcBuyToolPrice()

	if !work.HasHitMaxToolLimit() {

		// Adds each tool present in the config file
		components = append(components, &discordgo.Button{
			Label:    fmt.Sprintf("Buy Tool (%s)", priceString),
			Style:    3, // Green color style
			Disabled: false,
			Emoji: discordgo.ComponentEmoji{
				Name: config.CONFIG.Emojis.ComponentEmojiNames.MoneyBag,
			},
			CustomID: "BWT", // 'BWT' is code for 'Buy Work Tool'
		})
	}

	if len(components) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
}
