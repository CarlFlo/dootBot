package bot

import (
	"github.com/CarlFlo/discordBotTemplate/config"
)

func isOwner(discordID string) bool {
	return discordID == config.CONFIG.OwnerID
}
