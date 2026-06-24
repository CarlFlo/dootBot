package commands

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/config"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/bwmarrin/discordgo"
)

var discordIDPattern = regexp.MustCompile(`\d+`)

const adminIdentityCacheTTL = 7 * 24 * time.Hour // 1 week

var adminIdentityCache = struct {
	mu          sync.RWMutex
	namesByID   map[string]string
	refreshedAt time.Time
}{
	namesByID: map[string]string{},
}

func AdminAdd(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(1) {
		utils.SendMessageFailure(m, "No Discord user ID provided")
		return
	}

	discordID := parseDiscordID(input.GetArgs()[0])
	if discordID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord user ID")
		return
	}

	if discordID == config.CONFIG.OwnerID {
		utils.SendMessageNeutral(m, "The configured owner already has permanent admin access")
		return
	}

	if database.IsStoredAdmin(discordID) {
		utils.SendMessageNeutral(m, fmt.Sprintf("%s is already an admin", formatAdminIdentity(s, discordID)))
		return
	}

	if err := database.AddAdmin(discordID); err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to add admin: %s", err))
		return
	}

	cacheAdminIdentity(s, discordID)
	utils.SendMessageSuccess(m, fmt.Sprintf("Added '%s' as an admin", formatAdminIdentity(s, discordID)))
}

func AdminRemove(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(1) {
		utils.SendMessageFailure(m, "No Discord user ID provided")
		return
	}

	discordID := parseDiscordID(input.GetArgs()[0])
	if discordID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord user ID")
		return
	}

	if discordID == config.CONFIG.OwnerID {
		utils.SendMessageFailure(m, "The configured owner cannot have admin access removed")
		return
	}

	if err := database.RemoveAdmin(discordID); err != nil {
		if database.IsAdminNotFound(err) {
			utils.SendMessageNeutral(m, fmt.Sprintf("%s is not an additional admin", formatAdminIdentity(s, discordID)))
			return
		}
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to remove admin: %s", err))
		return
	}

	removeCachedAdminIdentity(discordID)
	utils.SendMessageSuccess(m, fmt.Sprintf("Removed '%s' from admins", formatAdminIdentity(s, discordID)))
}

func AdminList(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	admins, err := database.GetAdmins()
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load admin list: %s", err))
		return
	}

	adminIDs := make([]string, 0, len(admins)+1)
	adminIDs = append(adminIDs, config.CONFIG.OwnerID)
	for _, admin := range admins {
		adminIDs = append(adminIDs, admin.DiscordID)
	}
	ensureAdminIdentityCache(s, adminIDs)

	lines := []string{}

	lines = append(lines, "**DootBot admins**")
	lines = append(lines, fmt.Sprintf("Owner: %s", formatAdminIdentity(s, config.CONFIG.OwnerID)))

	if len(admins) == 0 {
		lines = append(lines, "Admins: none")
	} else {
		extraAdmins := make([]string, 0, len(admins))
		for _, admin := range admins {
			extraAdmins = append(extraAdmins, formatAdminIdentity(s, admin.DiscordID))
		}
		lines = append(lines, fmt.Sprintf("Admins: %s", strings.Join(extraAdmins, ", ")))
	}

	utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
}

func parseDiscordID(input string) string {
	return discordIDPattern.FindString(input)
}

func formatAdminIdentity(s *discordgo.Session, discordID string) string {
	ensureAdminIdentityCache(s, []string{discordID})

	adminIdentityCache.mu.RLock()
	cachedName := adminIdentityCache.namesByID[discordID]
	adminIdentityCache.mu.RUnlock()
	if cachedName != "" {
		return cachedName
	}

	name := resolveDiscordUsername(s, discordID)
	cacheResolvedAdminIdentity(discordID, name)
	return name
}

func ensureAdminIdentityCache(s *discordgo.Session, discordIDs []string) {
	if len(discordIDs) == 0 {
		return
	}

	adminIdentityCache.mu.RLock()
	cacheEmpty := len(adminIdentityCache.namesByID) == 0
	cacheExpired := !adminIdentityCache.refreshedAt.IsZero() && time.Since(adminIdentityCache.refreshedAt) >= adminIdentityCacheTTL
	adminIdentityCache.mu.RUnlock()

	if cacheEmpty || cacheExpired {
		refreshAdminIdentityCache(s, discordIDs)
		return
	}

	missingIDs := make([]string, 0)
	adminIdentityCache.mu.RLock()
	for _, discordID := range discordIDs {
		if _, ok := adminIdentityCache.namesByID[discordID]; !ok {
			missingIDs = append(missingIDs, discordID)
		}
	}
	adminIdentityCache.mu.RUnlock()

	if len(missingIDs) > 0 {
		refreshMissingAdminIdentities(s, missingIDs)
	}
}

func refreshAdminIdentityCache(s *discordgo.Session, discordIDs []string) {
	namesByID := make(map[string]string, len(discordIDs))
	for _, discordID := range discordIDs {
		namesByID[discordID] = resolveDiscordUsername(s, discordID)
	}

	adminIdentityCache.mu.Lock()
	adminIdentityCache.namesByID = namesByID
	adminIdentityCache.refreshedAt = time.Now()
	adminIdentityCache.mu.Unlock()
}

func refreshMissingAdminIdentities(s *discordgo.Session, discordIDs []string) {
	adminIdentityCache.mu.Lock()
	defer adminIdentityCache.mu.Unlock()

	for _, discordID := range discordIDs {
		if _, ok := adminIdentityCache.namesByID[discordID]; ok {
			continue
		}
		adminIdentityCache.namesByID[discordID] = resolveDiscordUsername(s, discordID)
	}
}

func cacheAdminIdentity(s *discordgo.Session, discordID string) {
	cacheResolvedAdminIdentity(discordID, resolveDiscordUsername(s, discordID))
}

func cacheResolvedAdminIdentity(discordID, name string) {
	adminIdentityCache.mu.Lock()
	adminIdentityCache.namesByID[discordID] = name
	if adminIdentityCache.refreshedAt.IsZero() {
		adminIdentityCache.refreshedAt = time.Now()
	}
	adminIdentityCache.mu.Unlock()
}

func removeCachedAdminIdentity(discordID string) {
	adminIdentityCache.mu.Lock()
	delete(adminIdentityCache.namesByID, discordID)
	adminIdentityCache.mu.Unlock()
}

func resolveDiscordUsername(s *discordgo.Session, discordID string) string {
	user, err := s.User(discordID)
	if err != nil || user == nil {
		return fmt.Sprintf("`%s`", discordID)
	}

	return formatDiscordUsername(user)
}

func formatDiscordUsername(user *discordgo.User) string {
	if user.Discriminator != "" && user.Discriminator != "0" {
		return fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	}

	return user.Username
}
