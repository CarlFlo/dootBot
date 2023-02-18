package bot

import (
	"github.com/CarlFlo/malm"
	"github.com/bwmarrin/discordgo"
)

// https://github.com/bwmarrin/discordgo/blob/master/examples/slash_commands/main.go

var defaultMemberPermissions int64 = discordgo.PermissionManageServer

func initializeSlashCommands(session *discordgo.Session) {

	commands := []*discordgo.ApplicationCommand{}

	createCommands(commands)

	registeredCommands := make([]*discordgo.ApplicationCommand, len(commands))
	for i, v := range commands {
		cmd, err := session.ApplicationCommandCreate(session.State.User.ID, "", v)
		if err != nil {
			malm.Fatal("Cannot create '%v' command: %v", v.Name, err)
		}
		registeredCommands[i] = cmd
	}

}

func createCommands(commands []*discordgo.ApplicationCommand) {

	commands = append(commands, &discordgo.ApplicationCommand{
		Name:        "The Name",
		Description: "The Description",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Name:        "Option 1",
				Description: "Description 1",
			},
		},
	})

}
