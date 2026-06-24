package bot

import (
	"fmt"
	"sort"
	"strings"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/permissions"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

type typeHolder struct {
	group    string
	commands []string
}

func generateHelp() {
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
	case typeMusic:
		cmdType = "Music"
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
	helpMap := make(map[string][]string)

	for cmd, data := range validCommands {
		if !helpVisible(input, data.requiredPermission) {
			continue
		}

		display := cmd
		if len(data.helpSyntax) > 0 {
			display += fmt.Sprintf(" %s", data.helpSyntax)
		}

		commandType := commandTypeToString(data.commandType)
		helpMap[commandType] = append(helpMap[commandType], display)
	}

	for _, list := range helpMap {
		sort.Strings(list)
	}

	th := []typeHolder{}
	for group, list := range helpMap {
		th = append(th, typeHolder{group, list})
	}

	sort.SliceStable(th, func(i, j int) bool {
		return th[i].group < th[j].group
	})

	output := strings.Builder{}
	for _, e := range th {
		output.WriteString(fmt.Sprintf("[%s]\n%s\n", e.group, strings.Join(e.commands, ", ")))
	}

	s.ChannelMessageSend(m.ChannelID, start+output.String()+end)
}

func helpVisible(input *structs.CmdInput, required permissions.Level) bool {
	switch required {
	case enumAdmin:
		return input.HasGuildPermission(permissions.LevelAdmin)
	case enumController:
		return input.HasGuildPermission(permissions.LevelController)
	case enumRequester:
		return input.HasGuildPermission(permissions.LevelRequester)
	default:
		return true
	}

}
