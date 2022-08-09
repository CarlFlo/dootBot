package bot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

var helpString string
var helpStringAdmin string

type typeHolder struct {
	group    string
	commands []string
}

// This function automatically generates the output for the help command
// It will generate two seperate strings, one for the users and one for the admins.
// The admin contains command only admins should know and only they can use
// The implementation is not super efficient, but it works
func generateHelp() {

	// The key is the 'commandType', followed by the commands for that type
	helpUserMap := make(map[string][]string)
	// Admin commands are all grouped together
	helpAdmin := []string{}

	// Populate the helpUserMap map with the commands and group them together
	// Hashmaps are inherently unpredictable in their order
	for cmd, data := range validCommands {
		// Adds the extra syntax if the commands has it
		if len(data.helpSyntax) > 0 {
			cmd += fmt.Sprintf(" %s", data.helpSyntax)
		}

		if data.requiredPermission == enumAdmin {
			helpAdmin = append(helpAdmin, cmd)
		} else {
			// Adds the command to the correct groups string slice
			commandType := commandTypeToString(data.commandType)
			helpUserMap[commandType] = append(helpUserMap[commandType], cmd)
		}
	}

	// Sorting the lists so the commands will be in alphabetical order
	sort.Strings(helpAdmin)
	for _, list := range helpUserMap {
		sort.Strings(list)
	}

	// Transfer the helpUserMap data to a slice so that the categories can be sorted
	th := []typeHolder{}
	for group, list := range helpUserMap {
		th = append(th, typeHolder{group, list})
	}

	// Sorts the command categories so they will be in alphabetical order
	sort.SliceStable(th, func(i, j int) bool {
		return th[i].group < th[j].group
	})

	// Create the help strings and caching the result
	for _, e := range th {
		// Saving the result for the users
		helpString += fmt.Sprintf("[%s]\n%s\n", e.group, strings.Join(e.commands[:], ", "))
	}

	// Saves the results for the admins
	helpStringAdmin = fmt.Sprintf("[%s]\n%s\n", "Admin", strings.Join(helpAdmin[:], ", "))

}

func commandTypeToString(key uint8) string {
	cmdType := ""
	switch key {
	case typeGeneral:
		cmdType = "General"
	case typeUser:
		cmdType = "User"
	case typeEconomy:
		cmdType = "Economy"
	case typeMisc:
		cmdType = "Misc"
	default:
		cmdType = "Unknown"
		malm.Warn("A command group type is unknown: %d", key)
	}
	return cmdType
}

// Automatically generate help for the user
func help(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	start := "```ini\n"
	end := "\n[Note]\nCommands are not case sensitive.\n```"

	// Admins will get additional help
	if input.IsAdmin() {
		s.ChannelMessageSend(m.ChannelID, start+helpStringAdmin+helpString+end)
	} else {
		s.ChannelMessageSend(m.ChannelID, start+helpString+end)
	}

}
