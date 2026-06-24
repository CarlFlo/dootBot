package bot

import (
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
)

func isOwner(discordID string) bool {
	return discordID == config.CONFIG.OwnerID
}

func isAdmin(discordID string) bool {
	return isOwner(discordID) || database.IsStoredAdmin(discordID)
}
