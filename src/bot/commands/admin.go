package commands

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/permissions"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/bwmarrin/discordgo"
)

var discordIDPattern = regexp.MustCompile(`\d+`)

func MusicAdmin(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if m.GuildID == "" {
		utils.SendMessageFailure(m, "This command can only be used in a server")
		return
	}

	if !input.NumberOfArgsAreAtleast(1) {
		sendMusicHelp(m)
		return
	}

	switch input.GetArgsLowercase()[0] {
	case "permissions":
		handleMusicPermissions(s, m, input)
	case "channels":
		handleMusicChannels(s, m, input)
	default:
		sendMusicHelp(m)
	}
}

func handleMusicPermissions(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(2) {
		sendMusicPermissionsHelp(m)
		return
	}

	switch input.GetArgsLowercase()[1] {
	case "view":
		musicPermissionsView(s, m)
	case "linkrole":
		musicPermissionsLinkRole(s, m, input)
	case "unlinkrole":
		musicPermissionsUnlinkRole(s, m, input)
	case "openrequests":
		musicPermissionsOpenRequests(m, input)
	default:
		sendMusicPermissionsHelp(m)
	}
}

func musicPermissionsLinkRole(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(4) {
		utils.SendMessageFailure(m, "Usage: music permissions linkrole [requester|controller|admin] <@role>")
		return
	}

	level, err := permissions.ParseLevel(input.GetArgsLowercase()[2])
	if err != nil {
		utils.SendMessageFailure(m, err.Error())
		return
	}

	roleID := parseDiscordID(input.GetArgs()[3])
	if roleID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord role ID")
		return
	}

	if err := database.SetGuildRolePermission(m.GuildID, roleID, uint8(level)); err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to link role: %s", err))
		return
	}

	utils.SendMessageSuccess(m, fmt.Sprintf("Linked role %s to %s", formatRoleIdentity(s, m.GuildID, roleID), level.String()))
}

func musicPermissionsUnlinkRole(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(3) {
		utils.SendMessageFailure(m, "Usage: music permissions unlinkrole <@role>")
		return
	}

	roleID := parseDiscordID(input.GetArgs()[2])
	if roleID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord role ID")
		return
	}

	if err := database.RemoveGuildRolePermission(m.GuildID, roleID); err != nil {
		if database.IsGuildPermissionNotFound(err) {
			utils.SendMessageNeutral(m, fmt.Sprintf("%s is not linked", formatRoleIdentity(s, m.GuildID, roleID)))
			return
		}
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to unlink role: %s", err))
		return
	}

	utils.SendMessageSuccess(m, fmt.Sprintf("Unlinked role %s", formatRoleIdentity(s, m.GuildID, roleID)))
}

func musicPermissionsOpenRequests(m *discordgo.MessageCreate, input *structs.CmdInput) {
	settings, err := database.GetGuildPermissionSettings(m.GuildID)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load permission settings: %s", err))
		return
	}

	if !input.NumberOfArgsAreAtleast(3) {
		lines := []string{
			"**Music Permissions: Open Requests**",
			fmt.Sprintf("Status: **%s**", toggleStateLabel(settings.OpenRequestsEnabled)),
			"`music permissions openrequests [on|off]`",
		}
		utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
		return
	}

	enabled, ok := parseToggleState(input.GetArgsLowercase()[2])
	if !ok {
		utils.SendMessageFailure(m, "Usage: music permissions openrequests [on|off]")
		return
	}

	if err := database.SetGuildOpenRequestsEnabled(m.GuildID, enabled); err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to update open requests: %s", err))
		return
	}

	utils.SendMessageSuccess(m, fmt.Sprintf("Open requests are now **%s**", toggleStateLabel(enabled)))
}

func musicPermissionsView(s *discordgo.Session, m *discordgo.MessageCreate) {
	roleLinks, err := database.ListGuildRolePermissions(m.GuildID)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load role permissions: %s", err))
		return
	}

	settings, err := database.GetGuildPermissionSettings(m.GuildID)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load permission settings: %s", err))
		return
	}

	lines := []string{
		"**Music Permissions**",
		fmt.Sprintf("Open requests: **%s**", toggleStateLabel(settings.OpenRequestsEnabled)),
	}

	if len(roleLinks) == 0 {
		lines = append(lines, "Role links: none")
	} else {
		lines = append(lines, "Role links:")
		for _, roleLink := range roleLinks {
			lines = append(lines, fmt.Sprintf("- %s: %s", formatRoleIdentity(s, m.GuildID, roleLink.RoleID), permissions.Level(roleLink.Level).String()))
		}
	}

	lines = append(lines, "`music permissions linkrole [requester|controller|admin] <@role>`")
	lines = append(lines, "`music permissions openrequests`")
	utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
}

func handleMusicChannels(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(2) {
		sendMusicChannelsHelp(m)
		return
	}

	switch input.GetArgsLowercase()[1] {
	case "view":
		musicChannelsView(s, m)
	case "bind":
		musicChannelsBind(s, m, input)
	case "unbind":
		musicChannelsUnbind(s, m, input)
	case "clear":
		musicChannelsClear(m)
	default:
		sendMusicChannelsHelp(m)
	}
}

func musicChannelsView(s *discordgo.Session, m *discordgo.MessageCreate) {
	channels, err := database.ListGuildMusicChannels(m.GuildID)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load music channels: %s", err))
		return
	}

	lines := []string{
		"**Music Channels**",
	}

	if len(channels) == 0 {
		lines = append(lines, "Bound channels: none")
		lines = append(lines, "Bot commands are allowed in **all channels**")
	} else {
		lines = append(lines, "Bound channels:")
		for _, channel := range channels {
			lines = append(lines, fmt.Sprintf("- %s", formatChannelIdentity(s, channel.ChannelID)))
		}
	}

	lines = append(lines, "`music channels bind <#channel>`")
	lines = append(lines, "`music channels clear`")
	utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
}

func musicChannelsBind(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(3) {
		utils.SendMessageFailure(m, "Usage: music channels bind <#channel>")
		return
	}

	channelID := parseDiscordID(input.GetArgs()[2])
	if channelID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord channel ID")
		return
	}

	if err := database.BindGuildMusicChannel(m.GuildID, channelID); err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to bind channel: %s", err))
		return
	}

	utils.SendMessageSuccess(m, fmt.Sprintf("Bound %s for bot commands", formatChannelIdentity(s, channelID)))
}

func musicChannelsUnbind(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if !input.NumberOfArgsAreAtleast(3) {
		utils.SendMessageFailure(m, "Usage: music channels unbind <#channel>")
		return
	}

	channelID := parseDiscordID(input.GetArgs()[2])
	if channelID == "" {
		utils.SendMessageFailure(m, "Could not parse the Discord channel ID")
		return
	}

	if err := database.UnbindGuildMusicChannel(m.GuildID, channelID); err != nil {
		if database.IsGuildPermissionNotFound(err) {
			utils.SendMessageNeutral(m, fmt.Sprintf("%s is not bound", formatChannelIdentity(s, channelID)))
			return
		}
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to unbind channel: %s", err))
		return
	}

	utils.SendMessageSuccess(m, fmt.Sprintf("Unbound %s", formatChannelIdentity(s, channelID)))
}

func musicChannelsClear(m *discordgo.MessageCreate) {
	if err := database.ClearGuildMusicChannels(m.GuildID); err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to clear music channels: %s", err))
		return
	}

	utils.SendMessageSuccess(m, "Cleared all music channel bindings")
}

func sendMusicHelp(m *discordgo.MessageCreate) {
	lines := []string{
		"**Music**",
		"`music permissions`",
		"`music permissions view`",
		"`music channels`",
		"`music channels view`",
	}
	utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
}

func sendMusicPermissionsHelp(m *discordgo.MessageCreate) {
	lines := []string{
		"**Music Permissions**",
		"`music permissions view`",
		"`music permissions linkrole requester <role ID or role mention>`",
		"`music permissions linkrole controller <role ID or role mention>`",
		"`music permissions linkrole admin <role ID or role mention>`",
		"`music permissions unlinkrole <role ID or role mention>`",
		"`music permissions openrequests`",
		"`music permissions openrequests [on|off]`",
	}
	utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
}

func sendMusicChannelsHelp(m *discordgo.MessageCreate) {
	lines := []string{
		"**Music Channels**",
		"`music channels view`",
		"`music channels bind #music`",
		"`music channels bind #bot-commands`",
		"`music channels unbind #music`",
		"`music channels clear`",
	}
	utils.SendMessageNeutral(m, strings.Join(lines, "\n"))
}

func parseDiscordID(input string) string {
	return discordIDPattern.FindString(input)
}

func formatRoleIdentity(s *discordgo.Session, guildID, roleID string) string {
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

func formatChannelIdentity(s *discordgo.Session, channelID string) string {
	if channel, err := s.State.Channel(channelID); err == nil && channel != nil {
		return fmt.Sprintf("#%s", channel.Name)
	}

	if channel, err := s.Channel(channelID); err == nil && channel != nil {
		return fmt.Sprintf("#%s", channel.Name)
	}

	return fmt.Sprintf("<#%s>", channelID)
}

func parseToggleState(input string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "on", "enable", "enabled", "true", "yes":
		return true, true
	case "off", "disable", "disabled", "false", "no":
		return false, true
	default:
		return false, false
	}
}

func toggleStateLabel(enabled bool) string {
	if enabled {
		return "Enabled"
	}
	return "Disabled"
}
