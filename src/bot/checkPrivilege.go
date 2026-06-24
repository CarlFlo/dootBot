package bot

import (
	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/permissions"
)

func hasCommandPermission(required permissions.Level, input *structs.CmdInput) bool {
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
