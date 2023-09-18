package database

import (
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/bwmarrin/discordgo"
	"gorm.io/gorm"
)

type Farm struct {
	Model
	Plots                   []*FarmPlot
	OwnedPlots              uint8
	LastWateredAt           time.Time // Last time the user watered the farm plots
	HighestPlantedCropIndex uint8

	PlotsChanged    bool `gorm:"-"` // Ignored by the database
	HarvestEarnings int  `gorm:"-"` // If 0 then no earnings
}

/*
	One farm per user, but you can have many plots
	The farm plot keeps track of which farm it belongs to

	Limit the crops the user can plant
	They unlock better crops by planting the basic first
	Planting crop ID 1 will then unlock ID 2 and so on
*/

func (Farm) TableName() string {
	return "userFarms"
}

// Does not trigger normaly.
func (f *Farm) BeforeCreate(tx *gorm.DB) error {

	f.HighestPlantedCropIndex = 1
	f.ResetLastWatered()
	return nil
}

// ResetLastWatered updates last watered to so that the user can water their plats again
func (f *Farm) ResetLastWatered() {
	f.LastWateredAt = time.Now().Add(time.Hour * config.CONFIG.Farm.WaterCooldown * -1)
}

// Saves the data to the database
func (f *Farm) Save() {

	// Updates/saves the plots as well
	if f.PlotsChanged {
		for _, plot := range f.Plots {
			plot.Save()
		}
	}

	DB.Save(&f)
}

// Queries the database for the farm data with the given user object.
// Needs to be called first
func (f *Farm) QueryUserFarmData(u *User) {
	DB.Raw("SELECT * FROM userFarms WHERE userFarms.ID = ?", u.ID).First(&f)
	if f.ID == 0 { // Meaning there is no data, so we initialize it
		f.ID = u.ID // The farms index is the same as the user
		f.OwnedPlots = config.CONFIG.Farm.DefaultOwnedFarmPlots
		f.HighestPlantedCropIndex = 1
		f.ResetLastWatered()

		f.Save()
	}
}

// Updates the object to contain all the farmplots
func (f *Farm) QueryFarmPlots() {
	DB.Raw("SELECT * FROM userFarmPlots WHERE userFarmPlots.Farm_ID = ? LIMIT ?", f.ID, f.OwnedPlots).Find(&f.Plots)
}

// TODO

// Peek looks at the farm and updates it to reflect the current state
// Always run before doing anything with the farm
// Checks if crops have perished
// Returns true if any crop perished
func (f *Farm) Peek() bool {

	anyCropsPerished := false

	// Checks if the user has watered their crops in the last x hours
	if float64(config.CONFIG.Farm.CropsPreishAfter) > time.Since(f.LastWateredAt).Hours() {
		return anyCropsPerished
	}

	// User has not watered in the last n hours. Thus missing the set deadline
	// Crops not fully grown (ready to be harvested) will perish
	for _, plot := range f.Plots {

		plot.QueryCropInfo()

		if plot.HasFullyGrown() {
			// Crop has fully grown so we need to check if the user didn't
			// just wait the whole growth duration without watering it

			crop := &plot.Crop
			// Calc the time when the crop will be fully grown
			fullyGrownAt := plot.PlantedAt.Add(crop.DurationToGrow)

			// Calc the time when the crop would have had to be watered
			waterDeadline := fullyGrownAt.Add(time.Hour * config.CONFIG.Farm.CropsPreishAfter * -1)

			// If the last time the plant was watered is after the deadline
			// Then the crop is ok, lse it will perish
			if f.LastWateredAt.After(waterDeadline) {
				// The crop is ok
				continue
			}
		}

		f.PlotsChanged = true
		anyCropsPerished = true
		plot.Perish()
	}

	f.Save()
	return anyCropsPerished
}

func (f *Farm) UpdateInteractionOverview(discordUser *discordgo.User, me *discordgo.MessageEdit) {

	f.overviewCreateEmbed(&me.Embeds, discordUser)

	var user User
	user.QueryUserByDiscordID(discordUser.ID)

	// Handle message components
	f.overviewCreateButtons(&me.Components, &user)
	f.overviewCreateCropMenu(&me.Components, &user)
}

// CreateFarmOverview creates the message that will be sent to the user
func (f *Farm) CreateFarmOverview(msg *discordgo.MessageSend, m *discordgo.MessageCreate, user *User) {

	f.QueryFarmPlots()

	// Takes a look at the farm, updating it to reflect the current state
	f.Peek()

	// Handel embeds
	f.overviewCreateEmbed(&msg.Embeds, m.Author)

	// Handle message components
	f.overviewCreateButtons(&msg.Components, user)
	f.overviewCreateCropMenu(&msg.Components, user) // Throwing an error currently
}

func (f *Farm) overviewCreateEmbed(embeds *[]*discordgo.MessageEmbed, discordUser *discordgo.User) {

	*embeds = append(*embeds, &discordgo.MessageEmbed{
		Type:        discordgo.EmbedTypeRich,
		Color:       config.CONFIG.Colors.Neutral,
		Title:       fmt.Sprintf("%s's Farm", discordUser.Username),
		Description: f.CreateEmbedDescription(),
		Fields:      f.CreateEmbedFields(),
		Footer: &discordgo.MessageEmbedFooter{
			Text: fmt.Sprintf("Crops will perish if not watered every day!\nYou can own up to %d farm plots!", config.CONFIG.Farm.MaxPlots),
		},
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: fmt.Sprintf("%s#%s", discordUser.AvatarURL("256"), discordUser.ID),
		},
	})
}

func (f *Farm) overviewCreateCropMenu(msgCompondents *[]discordgo.MessageComponent, user *User) {

	// User can't afford to plant so no need to create the menu
	if !user.CanAfford(uint64(config.CONFIG.Farm.CropSeedPrice)) {
		return
	} else if !f.HasFreePlot() {
		return
	}

	menuComponent := []discordgo.MessageComponent{
		&discordgo.SelectMenu{
			MenuType:    discordgo.StringSelectMenu,
			CustomID:    "FPC", // 'FPC' is code for 'Farm Plant Crop'
			Placeholder: fmt.Sprintf("Select a crop to plant (Cost %s %s)", utils.HumanReadableNumber(config.CONFIG.Farm.CropSeedPrice), config.CONFIG.Economy.Name),
			MaxValues:   1,
			Options:     f.createCropOptions(),
		},
	}

	*msgCompondents = append(*msgCompondents, discordgo.ActionsRow{
		Components: menuComponent,
	})

}

func (f *Farm) createCropOptions() []discordgo.SelectMenuOption {

	options := []discordgo.SelectMenuOption{}

	var crops []FarmCrop
	DB.Where("id <= ?", f.HighestPlantedCropIndex).Order("id desc").Limit(int(f.HighestPlantedCropIndex)).Find(&crops)
	//DB.Raw("SELECT * FROM farmCrops WHERE id <= ? ORDER BY id DESC LIMIT ?", f.HighestPlantedCropIndex, f.HighestPlantedCropIndex).Scan(&crops)

	for _, crop := range crops {

		options = append(options, discordgo.SelectMenuOption{
			Label: fmt.Sprintf("%s | %s | %s %s", crop.Name, crop.GetDuration(), utils.HumanReadableNumber(crop.HarvestReward), config.CONFIG.Economy.Name),
			Value: crop.Name,
			Emoji: discordgo.ComponentEmoji{
				Name: crop.Emoji,
			},
		})
	}

	return options
}

func (f *Farm) overviewCreateButtons(msgCompondents *[]discordgo.MessageComponent, user *User) {

	// Create the buttons
	btnComponents := []discordgo.MessageComponent{}

	btnComponents = append(btnComponents, &discordgo.Button{
		Label:    "Harvest",
		Disabled: !(f.CanHarvest() && f.HasPlantedPlots()),
		CustomID: "FH", // 'FH' is code for 'Farm Harvest'
	})
	btnComponents = append(btnComponents, &discordgo.Button{
		Label:    "Water",
		Disabled: !(f.CanWater() && f.HasPlantedPlots()), // Disable if nothing is planted
		CustomID: "FW",                                   // 'FW' is code for 'Farm Water'
	})

	// For buying an additional plot (only if they haven't reached the limit)
	if !f.HasMaxAmountOfPlots() {
		plotPrice := f.CalcFarmPlotPrice()

		canAffordPlot := user.Money >= uint64(plotPrice)

		// Add limit to the number of plots a user can buy

		btnComponents = append(btnComponents, &discordgo.Button{
			Label:    fmt.Sprintf("Buy Farm Plot (%s)", utils.HumanReadableNumber(plotPrice)),
			Style:    3, // Green color style
			Disabled: !canAffordPlot,
			Emoji: discordgo.ComponentEmoji{
				Name: config.CONFIG.Emojis.ComponentEmojiNames.MoneyBag,
			},
			CustomID: "BFP", // 'BFP' is code for 'Buy Farm Plot'
		})
	}

	btnComponents = append(btnComponents, &discordgo.Button{
		Label:    "Help",
		Style:    2, // Gray color style
		Disabled: false,
		Emoji: discordgo.ComponentEmoji{
			Name: config.CONFIG.Emojis.ComponentEmojiNames.Help,
		},
		CustomID: "FHELP", // 'FHELP' is code for 'Farm Help'; Provies commands and information regarding farming
	})

	*msgCompondents = append(*msgCompondents, discordgo.ActionsRow{
		Components: btnComponents,
	})
}

func (f *Farm) HasUserUnlocked(fc *FarmCrop) bool {

	return fc.ID <= uint(f.HighestPlantedCropIndex)
}

// Rember to run GetFarmPlots() before running this function
func (f *Farm) GetUnusedPlots() int {
	return int(f.OwnedPlots) - len(f.Plots)
}

// Returns true if the user has a free (unused) farm plot
func (f *Farm) HasFreePlot() bool {
	return f.GetUnusedPlots() > 0
}

// Returns true if anything is planted on the farm
// Run QueryFarmPlots() first
func (f *Farm) HasPlantedPlots() bool {
	return len(f.Plots) > 0
}

// Returns true if the user can water their crops
func (f *Farm) CanWater() bool {
	since := time.Since(f.LastWateredAt).Hours()
	return config.CONFIG.Debug.IgnoreWorkCooldown || since > float64(config.CONFIG.Farm.WaterCooldown)
}

// Returns the time the user can water their crops as a formatted discord string
// https://hammertime.cyou/
func (f *Farm) CanWaterAt() string {
	nextTime := f.LastWateredAt.Add(time.Hour * config.CONFIG.Farm.WaterCooldown).Unix()
	return fmt.Sprintf("<t:%d:R>", nextTime)
}

// Returns true if the user can harvest any of their crops
func (f *Farm) CanHarvest() bool {

	for _, plot := range f.Plots {
		plot.QueryCropInfo()

		if plot.HasFullyGrown() || plot.HasPerished() {
			return true
		}
	}
	return false
}

// Functions waters every single plot
// Meaning it will update the plantedAt time
// Run QueryFarmPlots() before running this function
func (f *Farm) WaterPlots() {

	// Update last watered at
	f.LastWateredAt = time.Now()
	// A change was made so it needs to be saved when farm Save function is called
	f.PlotsChanged = true

	for _, plot := range f.Plots {
		plot.Water()
	}
}

type harvestResult struct {
	Name    string
	Emoji   string
	Earning int
}

// Returns an array containing the crop object that was harvested
// Money earned is saved in f.HarvestEarnings. Remember to add it to the user's balance
// Run QueryFarmPlots() before running this function
func (f *Farm) HarvestPlots() []harvestResult {

	var result []harvestResult

	for _, plot := range f.Plots {

		plot.QueryCropInfo()

		if !plot.HasFullyGrown() && !plot.HasPerished() {
			continue // Not fully grown, so skip. Do not skip for perished plants
		}

		if !plot.HasPerished() {
			result = append(result,
				harvestResult{
					Name:    plot.Crop.Name,
					Emoji:   plot.Crop.Emoji,
					Earning: plot.Crop.HarvestReward,
				})

			f.HarvestEarnings += plot.Crop.HarvestReward
		}

		// Delete from the database
		defer f.DeletePlot(plot) // We cannot delete it at once, because we are iterating over it
	}

	return result
}

func (f *Farm) SuccessfulHarvest() bool {
	return f.HarvestEarnings > 0
}

func (f *Farm) MissedWaterDeadline() bool {
	return time.Since(f.LastWateredAt).Hours() > float64(config.CONFIG.Farm.CropsPreishAfter)
}

// Will return the amount of crops that have perished from plots
// Will also remove perished crop plots from the database
// Remember to have call GetFarmPlots() before calling this function
func (f *Farm) CropsPerishedCheck() []string {

	// Checks if the user has watered their crops in the last x hours
	if !f.MissedWaterDeadline() {
		return []string{}
	}

	// User has not watered in the last n hours. Thus missing the set deadline
	// Crops not fully grown (ready to be harvested) will perish

	var perishedCrops []string

	for _, plot := range f.Plots {

		plot.QueryCropInfo()

		if plot.HasFullyGrown() {
			// Crop has fully grown so we need to check if the user didn't
			// just wait the whole growth duration without watering it

			crop := &plot.Crop
			// Calc the time when the crop will be fully grown
			fullyGrownAt := plot.PlantedAt.Add(crop.DurationToGrow)

			// Calc the time when the crop would have had to be watered
			waterDeadline := fullyGrownAt.Add(time.Hour * config.CONFIG.Farm.CropsPreishAfter * -1)

			// If the last time the plant was watered is after the deadline
			// Then the crop is ok, lse it will perish
			if f.LastWateredAt.After(waterDeadline) {
				// The crop is ok
				continue
			}
		}

		perishedCrops = append(perishedCrops, plot.Crop.Name)
		defer f.DeletePlot(plot) // We cannot delete it at once, because we are iterating over it the list
	}

	return perishedCrops
}

func (f *Farm) DeletePlot(plot *FarmPlot) {

	f.PlotsChanged = true
	// Remove plot from f.Plots
	for i, p := range f.Plots {
		if p.ID == plot.ID {
			f.Plots = append(f.Plots[:i], f.Plots[i+1:]...)
			break
		}
	}

	plot.DeleteFromDB()
}

// Returns the cost of what buying a new plot would cost for the user
func (f *Farm) CalcFarmPlotPrice() int {

	// floor (Base price * multiplier ^ (number of plots - 1))

	return int(math.Floor(
		float64(config.CONFIG.Farm.FarmPlotPrice) * math.Pow(
			config.CONFIG.Farm.FarmPlotCostMultiplier,
			float64(f.OwnedPlots-1))))
}

func (f *Farm) CreateEmbedDescription() string {

	description := fmt.Sprintf("You currently own %d plot", f.OwnedPlots)

	if f.OwnedPlots > 1 {
		description += "s"
	}

	return description
}

func (f *Farm) CreateEmbedFields() []*discordgo.MessageEmbedField {
	var fields []*discordgo.MessageEmbedField

	f.QueryFarmPlots()

	for i, p := range f.Plots {

		p.QueryCropInfo()

		fields = append(fields, &discordgo.MessageEmbedField{
			Name:   fmt.Sprintf("%d) %s %s", i+1, p.Crop.Emoji, p.Crop.Name),
			Value:  p.HarvestableAt(),
			Inline: true,
		})
	}

	unusedPlots := f.OwnedPlots - uint8(len(f.Plots))

	emptyPlotValue := strings.Repeat(config.CONFIG.Emojis.EmptyPlot, 5)

	for i := 0; i < int(unusedPlots); i++ {
		fields = append(fields, &discordgo.MessageEmbedField{
			Name: fmt.Sprintf("%d) Empty Plot ", i+1+len(f.Plots)),
			//Value:  "⠀",
			Value:  emptyPlotValue,
			Inline: true,
		})
	}

	return fields
}

func (f *Farm) HasMaxAmountOfPlots() bool {
	return f.OwnedPlots >= config.CONFIG.Farm.MaxPlots
}
