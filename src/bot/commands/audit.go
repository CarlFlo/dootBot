package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/CarlFlo/dootBot/src/bot/structs"
	"github.com/CarlFlo/dootBot/src/database"
	"github.com/CarlFlo/dootBot/src/utils"
	"github.com/bwmarrin/discordgo"
)

const musicAuditPageSize = 10

func Audit(s *discordgo.Session, m *discordgo.MessageCreate, input *structs.CmdInput) {
	if m.GuildID == "" {
		utils.SendMessageFailure(m, "This command can only be used in a server")
		return
	}

	page, ok := parseAuditPage(input.GetArgs(), input.GetArgsLowercase())
	if !ok {
		utils.SendMessageFailure(m, "Usage: audit [page number]")
		return
	}

	logs, total, err := database.ListMusicAuditLogs(m.GuildID, page, musicAuditPageSize)
	if err != nil {
		utils.SendMessageFailure(m, fmt.Sprintf("Failed to load audit log: %s", err))
		return
	}

	totalPages := database.MusicAuditTotalPages(total, musicAuditPageSize)
	if total == 0 {
		utils.SendMessageNeutral(m, "**Music Audit Log**\nNo music audit entries have been recorded for this server yet.")
		return
	}

	if page > totalPages {
		utils.SendMessageFailure(m, fmt.Sprintf("Page %d does not exist. Available pages: 1-%d", page, totalPages))
		return
	}

	lines := []string{
		"**Music Audit Log**",
		fmt.Sprintf("Page **%d/%d** | %d entries total | newest first", page, totalPages, total),
		"",
	}

	for _, entry := range logs {
		lines = append(lines, formatAuditEntry(s, m.GuildID, entry))
		lines = append(lines, "")
	}

	if page < totalPages {
		lines = append(lines, database.FormatMusicAuditPageUsage(page+1))
	}

	utils.SendMessageNeutral(m, strings.TrimSpace(strings.Join(lines, "\n")))
}

func parseAuditPage(args, argsLower []string) (int, bool) {
	if len(args) == 0 {
		return 1, true
	}

	candidate := strings.TrimSpace(args[0])
	if len(argsLower) > 1 && argsLower[0] == "page" {
		candidate = strings.TrimSpace(args[1])
	}

	page, err := strconv.Atoi(candidate)
	if err != nil || page < 1 {
		return 0, false
	}

	return page, true
}

func formatAuditEntry(s *discordgo.Session, guildID string, entry database.MusicAuditLog) string {
	name := resolveAuditUserName(s, guildID, entry.UserID)
	timestamp := entry.CreatedAt.Format("2006-01-02 15:04:05")
	details := auditDetails(entry)

	return fmt.Sprintf("`%s` **%s** %s%s", timestamp, name, entry.Action, details)
}

func auditDetails(entry database.MusicAuditLog) string {
	parts := []string{}

	if entry.SongTitle != "" {
		if entry.SongAuthor != "" {
			parts = append(parts, fmt.Sprintf("**%s** by **%s**", entry.SongTitle, entry.SongAuthor))
		} else {
			parts = append(parts, fmt.Sprintf("**%s**", entry.SongTitle))
		}
	}

	if entry.Description != "" {
		parts = append(parts, entry.Description)
	}

	if len(parts) == 0 {
		return ""
	}

	return " | " + strings.Join(parts, " | ")
}

func resolveAuditUserName(s *discordgo.Session, guildID, userID string) string {
	if s == nil || guildID == "" || userID == "" {
		return fallbackAuditUserName(userID)
	}

	if member, err := s.State.Member(guildID, userID); err == nil && member != nil {
		if name := preferredMemberName(member); name != "" {
			return name
		}
	}

	if member, err := s.GuildMember(guildID, userID); err == nil && member != nil {
		if name := preferredMemberName(member); name != "" {
			return name
		}
	}

	return fallbackAuditUserName(userID)
}

func preferredMemberName(member *discordgo.Member) string {
	if member == nil {
		return ""
	}

	if member.Nick != "" {
		return member.Nick
	}
	if member.User == nil {
		return ""
	}
	if member.User.GlobalName != "" {
		return member.User.GlobalName
	}
	return member.User.Username
}

func fallbackAuditUserName(userID string) string {
	if userID == "" {
		return "Unknown user"
	}

	return fmt.Sprintf("Unknown user (%s)", userID)
}
