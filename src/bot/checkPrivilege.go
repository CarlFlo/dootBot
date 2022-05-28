package bot

import (
	"github.com/CarlFlo/DiscordMoneyBot/src/config"
)

func isOwner(discordID string) bool {
	return discordID == config.CONFIG.OwnerID
}
