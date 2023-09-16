package database

import (
	"fmt"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

type User struct {
	Model
	DiscordID        string `gorm:"uniqueIndex"`
	Money            uint64
	LifetimeEarnings uint64
	Work             Work  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Daily            Daily `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	Farm             Farm  `gorm:"foreignKey:ID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}

func (User) TableName() string {
	return "users"
}

func (u *User) AfterCreate(tx *gorm.DB) error {
	// Log in debug DB maybe
	return nil
}

// Saves the data to the database
func (u *User) Save() {
	DB.Save(&u)
}

// Returns true if a user with that discord ID exists in the database
func (u *User) DoesUserExist(discordID string) bool {

	var count int
	DB.Raw("SELECT COUNT(*) FROM users WHERE discord_id = ?", discordID).Scan(&count)

	return count == 1
}

// Queries the database for the user with the given discord ID.
// The object which calls the method will be updated with the user's data
func (u *User) QueryUserByDiscordID(discordID string) {
	DB.Table("users").Where("discord_id = ?", discordID).First(&u)
}

func (u *User) PrettyPrintMoney() string {
	return utils.HumanReadableNumber(u.Money)
}

func (u *User) PrettyPrintLifetimeEarnings() string {
	return utils.HumanReadableNumber(u.LifetimeEarnings)
}

func (u *User) AddMoney(amount uint64) {
	u.Money += amount
	u.LifetimeEarnings += amount
}

func (u *User) DeductMoney(amount uint64) {
	u.Money -= amount
}

func (u *User) CanAfford(number uint64) bool {
	return u.Money >= number
}

func (u *User) CreateProfileEmbeds(du *discordgo.User, work *Work, daily *Daily, embeds *[]*discordgo.MessageEmbed) {

	*embeds = append(*embeds, &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Color:       config.CONFIG.Colors.Neutral,
		Title:       fmt.Sprintf("%s's profile", du.Username),
		Description: "",
		Fields:      u.createProfileFields(work, daily),
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("%s#%s", du.AvatarURL("256"), du.ID),
		},
	})
}

func (u *User) CreateProfileComponents(work *Work, daily *Daily) []discordgo.MessageComponent {

	components := []discordgo.MessageComponent{}

	// Only create if the daily can be done
	components = append(components, &discordgo.Button{
		Label:    "Collect Daily",
		Style:    1, // Default purple
		Disabled: false,
		CustomID: "PD", // 'PD' is code for 'Profile Daily'
	})

	// Only create if the work can be done
	components = append(components, &discordgo.Button{
		Label:    "Work",
		Style:    1, // Default purple
		Disabled: false,
		CustomID: "PW", // 'PW' is code for 'Profile Work'
	})

	components = append(components, &discordgo.Button{
		Label:    "",
		Style:    1, // Default purple
		Disabled: false,
		Emoji: discordgo.ComponentEmoji{
			Name: config.CONFIG.Emojis.ComponentEmojiNames.Refresh,
		},
		CustomID: "RP", // 'RP' is code for 'Refresh Profile'
	})

	if len(components) == 0 {
		return nil
	}

	return []discordgo.MessageComponent{discordgo.ActionsRow{Components: components}}
}

// CreateProfileFields generates the profile fields for message
func (u *User) createProfileFields(work *Work, daily *Daily) []*discordgo.MessageEmbedField {
	// The statuses on the cooldown's
	workStatus := config.CONFIG.Emojis.Success
	if !work.CanDoWork() {
		workStatus = fmt.Sprintf("%s Available %s", config.CONFIG.Emojis.Failure, work.CanDoWorkAt())
	}

	dailyStatus := config.CONFIG.Emojis.Success
	if !daily.CanDoDaily() {
		dailyStatus = fmt.Sprintf("%s Available %s", config.CONFIG.Emojis.Failure, daily.CanDoDailyAt())
	}

	fields := []*discordgo.MessageEmbedField{
		{
			Name:   fmt.Sprintf("Wallet %s", config.CONFIG.Emojis.Wallet),
			Value:  fmt.Sprintf("%s %s", config.CONFIG.Emojis.Economy, u.PrettyPrintMoney()),
			Inline: true,
		},
		{
			Name:   "Daily",
			Value:  dailyStatus,
			Inline: true,
		},
		{
			Name:   "Work",
			Value:  workStatus,
			Inline: true,
		},
	}

	return fields
}
