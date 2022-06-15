package commands

import (
	"bufio"
	"bytes"
	"fmt"
	"os"

	"github.com/CarlFlo/DiscordMoneyBot/src/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
	"github.com/CarlFlo/DiscordMoneyBot/src/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

// Presence - Reloads the configuration without restarting the application
func Presence(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {

	// Should extra information be provided?
	verbose := input.ArgsContains([]string{"v", "verbose"}) // Will output additional information
	dumpToFile := input.ArgsContains([]string{"d", "dump"}) // Will not output to discord, only dump to file

	var outputBuffer []string
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("Servers: %d\n", len(s.State.Guilds)))

	if verbose {
		b.WriteString("id | name | members\n")
	} else {
		b.WriteString("id\n")
	}

	for _, guild := range s.State.Guilds {
		var output string

		if verbose {
			// Each guild takes around 55-75 characters. Wich means ~30 servers per message
			output = fmt.Sprintf("%s | %s | %d\n", guild.ID, guild.Name, guild.MemberCount)
		} else {
			// 18 characters per servers. ~102 per message
			output = fmt.Sprintf("%s\n", guild.ID)
		}

		b.WriteString(output)

		// Discord messages has a 2000 character limit
		if b.Len() > config.CONFIG.MessageProcessing.MessageLengthLimit {
			// Appends the data and resets the buffer
			outputBuffer = append(outputBuffer, b.String())
			b.Reset()
		}
	}

	// Appends the final message
	if b.Len() > 0 {
		outputBuffer = append(outputBuffer, b.String())
	}

	if dumpToFile {
		// Output to file
		file, err := os.OpenFile("presenceOutput.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

		if err != nil {
			malm.Error("Could not create presence output file: %s", err)
			return
		}

		datawriter := bufio.NewWriter(file)
		for _, data := range outputBuffer {
			_, err = datawriter.WriteString(data + "\n")
			if err != nil {
				malm.Error("%s\n", err)
			}
		}

		datawriter.Flush()
		file.Close()

	} else {
		// Output to discord. Discord message limit is ~5 messages a second
		for _, line := range outputBuffer {
			utils.SendDirectMessage(m, fmt.Sprintf("```%s```", line))
		}
	}
}
