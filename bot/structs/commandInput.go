package structs

import (
	"strings"

	"github.com/CarlFlo/DiscordMoneyBot/config"
)

// CmdInput holds a command
type CmdInput struct {
	command         string
	args            []string
	argsLowercase   []string
	adminPermission bool
}

// ParseInput parses a input string from the user
// It creates the CmdInput struct with the required data
func (I *CmdInput) ParseInput(input string, adminPerm bool) {
	// Remove prefix
	prefixLen := len(config.CONFIG.BotPrefix)
	input = input[prefixLen:]

	// Make lowercase
	inputLower := strings.ToLower(input)

	// The string is split
	args := strings.Split(input, " ")
	argsLowercase := strings.Split(inputLower, " ")

	// Saves the data in the struct
	I.command = args[0]
	I.args = args[1:]
	I.argsLowercase = argsLowercase[1:]
	I.adminPermission = adminPerm
}

func (I *CmdInput) NumberOfArgsAreAtleast(n int) bool {
	return len(I.argsLowercase) >= n
}

func (I *CmdInput) NumberOfArgsAre(n int) bool {
	return len(I.argsLowercase) == n
}

// GetCommand returns the command
func (I *CmdInput) GetCommand() string {
	return I.command
}

// GetArgs returns the args
func (I *CmdInput) GetArgs() []string {
	return I.args
}

// GetArgsLowercase returns the args
func (I *CmdInput) GetArgsLowercase() []string {
	return I.argsLowercase
}

// IsAdmin returns true of command issuer is an admin to the bot
func (I *CmdInput) IsAdmin() bool {
	return I.adminPermission
}

// ArgsContains looks for and returns a bool depending on if the
// args in the command has a specific string in it from a slice
func (I *CmdInput) ArgsContains(query []string) bool {

	for _, args := range I.argsLowercase {
		for _, lookingFor := range query {
			if args == lookingFor {
				return true
			}
		}
	}
	return false
}
