package bot

import (
	"unicode"

	"github.com/CarlFlo/dootBot/src/bot/commands"
	"github.com/CarlFlo/dootBot/src/bot/commands/daily"
	"github.com/CarlFlo/dootBot/src/bot/commands/farming"
	"github.com/CarlFlo/dootBot/src/bot/commands/work"
	"github.com/CarlFlo/dootBot/src/bot/music"
	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/permissions"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

const (
	enumUser permissions.Level = iota
	enumRequester
	enumController
	enumAdmin
)

// Command type for sorting similar commands together
const (
	typeGeneral uint8 = iota // General commands (All admin commands should be considered general)
	typeUser                 // Commands for the users
	typeEconomy              // Commands for the economy
	typeMusic                //	Commands for the music
	typeMisc                 // Miscellaneous commands
)

type command struct {
	function           func(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput)
	requiredPermission permissions.Level
	helpSyntax         string
	commandType        uint8
}

// Add variable to specify that command only can be run in a guild, not in a directmessage

var validCommands = make(map[string]command)

// mapValidCommands will initialize a map
// with all the valid functions that can be run
func mapValidCommands() {

	/* all keys MUST be lowercase */
	// Admin commands
	validCommands["reload"] = command{
		function:           commands.Reload,
		requiredPermission: enumAdmin,
		commandType:        typeGeneral}

	/*
		validCommands["debug"] = command{
			function:           commands.Debug,
			requiredPermission: enumAdmin,
			commandType:        typeGeneral}
	*/
	/*
		validCommands["presence"] = command{
			function:           commands.Presence,
			requiredPermission: enumAdmin,
			helpSyntax:         "[v verbose, d dump]",
			commandType:        typeGeneral}
	*/
	/*
		validCommands["gleave"] = command{
			function:           commands.LeaveServer,
			requiredPermission: enumAdmin,
			helpSyntax:         "[server/guild ID]",
			commandType:        typeGeneral}
	*/
	validCommands["music"] = command{
		function:           commands.MusicAdmin,
		requiredPermission: enumAdmin,
		helpSyntax:         "[permissions|channels]",
		commandType:        typeGeneral}

	// Perm User - General commands
	validCommands["help"] = command{
		function:           help,
		requiredPermission: enumUser,
		commandType:        typeGeneral}

	validCommands["profile"] = command{
		function:           commands.Profile,
		requiredPermission: enumUser,
		commandType:        typeGeneral}

	/*
		validCommands["mine"] = command{
			function:           mine.Dwarvenkeep,
			requiredPermission: enumUser,
			commandType:        typeGeneral}
	*/

	// Music
	validCommands["play"] = command{
		function:           music.PlayMusic,
		requiredPermission: enumRequester,
		helpSyntax:         "[youtube url/search query]",
		commandType:        typeMusic}

	validCommands["pause"] = command{
		function:           music.PauseMusic,
		requiredPermission: enumController,
		commandType:        typeMusic}

	validCommands["stop"] = command{ // Will also leave the voice channel
		function:           music.StopMusic,
		requiredPermission: enumController,
		commandType:        typeMusic}

	validCommands["skip"] = command{ // Same as next
		function:           music.SkipMusic,
		requiredPermission: enumController,
		commandType:        typeMusic}

	validCommands["next"] = command{ // Same as skip
		function:           music.SkipMusic,
		requiredPermission: enumController,
		commandType:        typeMusic}

	validCommands["loop"] = command{
		function:           music.LoopMusic,
		requiredPermission: enumController,
		commandType:        typeMusic}

	validCommands["previous"] = command{ // Goes back one song or restarts the current one
		function:           music.MusicPrevious,
		requiredPermission: enumController,
		commandType:        typeMusic}

	validCommands["restart"] = command{
		function:           music.MusicPrevious,
		requiredPermission: enumController,
		commandType:        typeMusic}

	validCommands["clearqueue"] = command{ // Clears the queue
		function:           music.ClearQueueMusic,
		requiredPermission: enumController,
		commandType:        typeMusic}

	// Perm User - Economy commands
	validCommands["balance"] = command{
		function:           commands.Balance,
		requiredPermission: enumUser,
		commandType:        typeEconomy}

	validCommands["farm"] = command{
		function:           farming.Farming,
		requiredPermission: enumUser,
		commandType:        typeEconomy}

	validCommands["work"] = command{
		function:           work.Work,
		requiredPermission: enumUser,
		commandType:        typeEconomy}

	validCommands["daily"] = command{
		function:           daily.Daily,
		requiredPermission: enumUser,
		commandType:        typeEconomy}

	/*
		// Perm User - Misc commands
		validCommands["ping"] = command{
			function:           commands.Ping,
			requiredPermission: enumUser,
			commandType:        typeMisc}

		validCommands["botinvite"] = command{
			function:           commands.BotInvite,
			requiredPermission: enumUser,
			commandType:        typeMisc}
	*/
	// Validates the keys so no-one is uppercase
	validateKeys()
	generateHelp()
}

// Validates that all the keys are lowercase
func validateKeys() {
	for key := range validCommands {
		for _, char := range key {
			if !unicode.IsLower(char) {
				malm.Fatal("key: '%s' contains one or more non lowercase characters: '%c'", key, char)
			}
		}
	}
}
