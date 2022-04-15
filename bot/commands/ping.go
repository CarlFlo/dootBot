package commands

import (
	"fmt"
	"time"

	"github.com/CarlFlo/DiscordMoneyBot/bot/structs"
	"github.com/CarlFlo/DiscordMoneyBot/utils"
	"github.com/CarlFlo/malm"

	"github.com/bwmarrin/discordgo"
)

var cache = make(map[string]time.Time)

// Ping - Sends back a 'Pong' message
func Ping(s *discordgo.Session, m *discordgo.MessageCreate, input structs.CmdInput) {

	// Send ping
	pingMsg, err := utils.SendDirectMessage(s, m, "Pinging...")
	if err != nil {
		malm.Error("Error: %s", err)
		return
	}

	time, err := discordgo.SnowflakeTimestamp(pingMsg.ID)
	if err != nil {
		malm.Error("Error: %s", err)
		return
	}

	// Caches the info that will be used to calculate the ping
	cache[pingMsg.ID] = time

	// Update message
	s.ChannelMessageEdit(pingMsg.ChannelID, pingMsg.ID, "Pinging... :bar_chart:")
}

// Pong updates the ping message with the ping duration
// It parses the time and calculates the diff between the cached time and the new time
// The difference in miliseconds is edited on the message
// The id is cleared from the cache
func Pong(s *discordgo.Session, mu *discordgo.MessageUpdate) {

	if cachedTime, ok := cache[mu.ID]; ok {
		// Removes that message id from the cache
		delete(cache, mu.ID)

		// Parses the time and ignores the error
		//newTime, _ := mu.EditedTimestamp.Parse()
		newTime := mu.EditedTimestamp
		diff := newTime.Sub(cachedTime)

		s.ChannelMessageEdit(mu.ChannelID, mu.ID, fmt.Sprintf("The ping is %v", diff))
	}
}
