package bot

import (
	"github.com/CarlFlo/DiscordMoneyBot/config"
)

func isOwner(discordID string) bool {
	return discordID == config.CONFIG.OwnerID
}
