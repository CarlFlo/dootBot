package bot

import (
	"github.com/CarlFlo/dootBot/src/config"
)

func isOwner(discordID string) bool {
	return discordID == config.CONFIG.OwnerID
}
