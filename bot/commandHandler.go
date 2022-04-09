package bot

import (
	"unicode"

	"github.com/CarlFlo/DiscordMoneyBot/bot/commands"
	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

const (
	enumUser uint8 = iota
	enumAdmin
)

// Command type for sorting similar commands together
const (
	typeGeneral uint8 = iota // General commands (All admin commands should be considered general)
	typeUser                 // Commands for the users
	typeMisc                 // Miscellaneous commands
)

type command struct {
	function           func(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput)
	requiredPermission uint8
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

	validCommands["debug"] = command{
		function:           commands.Debug,
		requiredPermission: enumAdmin,
		commandType:        typeGeneral}

	validCommands["presence"] = command{
		function:           commands.Presence,
		requiredPermission: enumAdmin,
		helpSyntax:         "[v verbose, d dump]",
		commandType:        typeGeneral}

	// Perm User - General commands
	validCommands["help"] = command{
		function:           help,
		requiredPermission: enumUser,
		commandType:        typeGeneral}

	validCommands["work"] = command{
		function:           commands.Work,
		requiredPermission: enumUser,
		commandType:        typeGeneral}

	// Perm User - Misc commands
	validCommands["ping"] = command{
		function:           commands.Ping,
		requiredPermission: enumUser,
		commandType:        typeMisc}

	validCommands["botinvite"] = command{
		function:           commands.BotInvite,
		requiredPermission: enumUser,
		commandType:        typeMisc}

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
