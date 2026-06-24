package commands

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/permissions"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/bwmarrin/discordgo"
)

var discordIDPattern = regexp.MustCompile(`\d+`)

const permissionIdentityCacheTTL = 7 * 24 * time.Hour

var permissionIdentityCache = struct {
	mu          sync.RWMutex
	namesByID   map[string]string
	refreshedAt time.Time
}{
	namesByID: map[string]string{},
}

func AdminAdd(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if m.GuildID == "" {
		utils.SendMessageFailure(m, "This command can only be used in a server")
		return
	}
	if !input.NumberOfArgsAreAtleast(2) {
		utils.SendMessageFailure(m, "Usage: promote [requester/controller/admin] [user ID or mention]")
		return
	}

	level, err := permissions.ParseLevel(input.GetArgsLowercase()[0])
	if err != nil {
		utils.SendMessageFailure(m, err.Error())
		return
	}

	discordID := parseDiscordID(input.GetArgs()[1])
	if discordID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord user ID")
		return
	}

	if err := database.SetGuildUserPermission(m.GuildID, discordID, uint8(level)); err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to set permission: %s", err))
		return
	}

	cachePermissionIdentity(s, discordID)
	utils.SendMessageSuccess(m, fmt.Sprintf("Granted %s to %s in this server", level.String(), formatPermissionIdentity(s, discordID)))
}

func AdminRemove(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if m.GuildID == "" {
		utils.SendMessageFailure(m, "This command can only be used in a server")
		return
	}
	if !input.NumberOfArgsAreAtleast(1) {
		utils.SendMessageFailure(m, "Usage: demote [user ID or mention]")
		return
	}

	discordID := parseDiscordID(input.GetArgs()[0])
	if discordID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord user ID")
		return
	}

	if err := database.RemoveGuildUserPermission(m.GuildID, discordID); err != nil {
		if database.IsGuildPermissionNotFound(err) {
			utils.SendMessageNeutral(m, fmt.Sprintf("%s has no explicit permission override in this server", formatPermissionIdentity(s, discordID)))
			return
		}
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to remove permission: %s", err))
		return
	}

	removeCachedPermissionIdentity(discordID)
	utils.SendMessageSuccess(m, fmt.Sprintf("Removed explicit permissions for %s in this server", formatPermissionIdentity(s, discordID)))
}

func AdminLinkRole(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if m.GuildID == "" {
		utils.SendMessageFailure(m, "This command can only be used in a server")
		return
	}
	if !input.NumberOfArgsAreAtleast(2) {
		utils.SendMessageFailure(m, "Usage: linkrole [requester/controller/admin] [role ID or mention]")
		return
	}

	level, err := permissions.ParseLevel(input.GetArgsLowercase()[0])
	if err != nil {
		utils.SendMessageFailure(m, err.Error())
		return
	}

	roleID := parseDiscordID(input.GetArgs()[1])
	if roleID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord role ID")
		return
	}

	if err := database.SetGuildRolePermission(m.GuildID, roleID, uint8(level)); err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to link role: %s", err))
		return
	}

	utils.SendMessageSuccess(m, fmt.Sprintf("Linked role %s to %s in this server", formatRoleIdentity(s, m.GuildID, roleID), level.String()))
}

func AdminUnlinkRole(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if m.GuildID == "" {
		utils.SendMessageFailure(m, "This command can only be used in a server")
		return
	}
	if !input.NumberOfArgsAreAtleast(1) {
		utils.SendMessageFailure(m, "Usage: unlinkrole [role ID or mention]")
		return
	}

	roleID := parseDiscordID(input.GetArgs()[0])
	if roleID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord role ID")
		return
	}

	if err := database.RemoveGuildRolePermission(m.GuildID, roleID); err != nil {
		if database.IsGuildPermissionNotFound(err) {
			utils.SendMessageNeutral(m, fmt.Sprintf("%s is not linked to a DootBot permission in this server", formatRoleIdentity(s, m.GuildID, roleID)))
			return
		}
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to unlink role: %s", err))
		return
	}

	utils.SendMessageSuccess(m, fmt.Sprintf("Unlinked role %s from DootBot permissions in this server", formatRoleIdentity(s, m.GuildID, roleID)))
}

func AdminList(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if m.GuildID == "" {
		utils.SendMessageFailure(m, "This command can only be used in a server")
		return
	}

	assignments, err := database.ListGuildUserPermissions(m.GuildID)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load user permissions: %s", err))
		return
	}

	roleLinks, err := database.ListGuildRolePermissions(m.GuildID)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load role permissions: %s", err))
		return
	}

	userIDs := make([]string, 0, len(assignments))
	for _, assignment := range assignments {
		userIDs = append(userIDs, assignment.DiscordID)
	}
	ensurePermissionIdentityCache(s, userIDs)

	guild, _ := s.Guild(m.GuildID)
	lines := []string{
		"**DootBot permissions for this server**",
	}

	if guild != nil {
		lines = append(lines, fmt.Sprintf("Owner: %s", formatPermissionIdentity(s, guild.OwnerID)))
	}
	lines = append(lines, "Members with Discord Administrator or Manage Server are also treated as Admin")

	if len(assignments) == 0 {
		lines = append(lines, "User grants: none")
	} else {
		lines = append(lines, "User grants:")
		for _, assignment := range assignments {
			lines = append(lines, fmt.Sprintf("- %s: %s", formatPermissionIdentity(s, assignment.DiscordID), permissions.Level(assignment.Level).String()))
		}
	}

	if len(roleLinks) == 0 {
		lines = append(lines, "Role links: none")
	} else {
		lines = append(lines, "Role links:")
		for _, roleLink := range roleLinks {
			lines = append(lines, fmt.Sprintf("- %s: %s", formatRoleIdentity(s, m.GuildID, roleLink.RoleID), permissions.Level(roleLink.Level).String()))
		}
	}

	utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
}

func parseDiscordID(input string) string {
	return discordIDPattern.FindString(input)
}

func formatPermissionIdentity(s *discordgo.Session, discordID string) string {
	ensurePermissionIdentityCache(s, []string{discordID})

	permissionIdentityCache.mu.RLock()
	cachedName := permissionIdentityCache.namesByID[discordID]
	permissionIdentityCache.mu.RUnlock()
	if cachedName != "" {
		return cachedName
	}

	name := resolveDiscordUsername(s, discordID)
	cacheResolvedPermissionIdentity(discordID, name)
	return name
}

func ensurePermissionIdentityCache(s *discordgo.Session, discordIDs []string) {
	if len(discordIDs) == 0 {
		return
	}

	permissionIdentityCache.mu.RLock()
	cacheEmpty := len(permissionIdentityCache.namesByID) == 0
	cacheExpired := !permissionIdentityCache.refreshedAt.IsZero() && time.Since(permissionIdentityCache.refreshedAt) >= permissionIdentityCacheTTL
	permissionIdentityCache.mu.RUnlock()

	if cacheEmpty || cacheExpired {
		refreshPermissionIdentityCache(s, discordIDs)
		return
	}

	missingIDs := make([]string, 0)
	permissionIdentityCache.mu.RLock()
	for _, discordID := range discordIDs {
		if _, ok := permissionIdentityCache.namesByID[discordID]; !ok {
			missingIDs = append(missingIDs, discordID)
		}
	}
	permissionIdentityCache.mu.RUnlock()

	if len(missingIDs) > 0 {
		refreshMissingPermissionIdentities(s, missingIDs)
	}
}

func refreshPermissionIdentityCache(s *discordgo.Session, discordIDs []string) {
	namesByID := make(map[string]string, len(discordIDs))
	for _, discordID := range discordIDs {
		namesByID[discordID] = resolveDiscordUsername(s, discordID)
	}

	permissionIdentityCache.mu.Lock()
	permissionIdentityCache.namesByID = namesByID
	permissionIdentityCache.refreshedAt = time.Now()
	permissionIdentityCache.mu.Unlock()
}

func refreshMissingPermissionIdentities(s *discordgo.Session, discordIDs []string) {
	permissionIdentityCache.mu.Lock()
	defer permissionIdentityCache.mu.Unlock()

	for _, discordID := range discordIDs {
		if _, ok := permissionIdentityCache.namesByID[discordID]; ok {
			continue
		}
		permissionIdentityCache.namesByID[discordID] = resolveDiscordUsername(s, discordID)
	}
}

func cachePermissionIdentity(s *discordgo.Session, discordID string) {
	cacheResolvedPermissionIdentity(discordID, resolveDiscordUsername(s, discordID))
}

func cacheResolvedPermissionIdentity(discordID, name string) {
	permissionIdentityCache.mu.Lock()
	permissionIdentityCache.namesByID[discordID] = name
	if permissionIdentityCache.refreshedAt.IsZero() {
		permissionIdentityCache.refreshedAt = time.Now()
	}
	permissionIdentityCache.mu.Unlock()
}

func removeCachedPermissionIdentity(discordID string) {
	permissionIdentityCache.mu.Lock()
	delete(permissionIdentityCache.namesByID, discordID)
	permissionIdentityCache.mu.Unlock()
}

func resolveDiscordUsername(s *discordgo.Session, discordID string) string {
	user, err := s.User(discordID)
	if err != nil || user == nil {
		return fmt.Sprintf("`%s`", discordID)
	}

	return formatDiscordUsername(user)
}

func resolveRoleName(s *discordgo.Session, guildID, roleID string) string {
	guild, err := s.State.Guild(guildID)
	if err != nil || guild == nil {
		guild, err = s.Guild(guildID)
	}
	if err != nil || guild == nil {
		return fmt.Sprintf("<@&%s>", roleID)
	}

	for _, role := range guild.Roles {
		if role.ID == roleID {
			return fmt.Sprintf("@%s", role.Name)
		}
	}

	return fmt.Sprintf("<@&%s>", roleID)
}

func formatRoleIdentity(s *discordgo.Session, guildID, roleID string) string {
	return resolveRoleName(s, guildID, roleID)
}

func formatDiscordUsername(user *discordgo.User) string {
	if user.Discriminator != "" && user.Discriminator != "0" {
		return fmt.Sprintf("%s#%s", user.Username, user.Discriminator)
	}

	return user.Username
}
