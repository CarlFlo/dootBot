package permissions

import (
	"errors"
	"fmt"
	"strings"

	"github.com/CarlFlo/dootBot/src/database"
	"github.com/bwmarrin/discordgo"
)

type Level uint8

const (
	LevelNone Level = iota
	LevelRequester
	LevelController
	LevelAdmin
)

type Context struct {
	GuildID         string
	UserID          string
	IsGuildOwner    bool
	HasDiscordAdmin bool
	ExplicitLevel   Level
	LinkedRoleLevel Level
	ResolvedLevel   Level
}

func (c Context) Has(level Level) bool {
	return c.ResolvedLevel >= level
}

func (c Context) IsAdmin() bool {
	return c.Has(LevelAdmin)
}

func (l Level) String() string {
	switch l {
	case LevelRequester:
		return "Requester"
	case LevelController:
		return "Controller"
	case LevelAdmin:
		return "Admin"
	default:
		return "None"
	}
}

func ParseLevel(input string) (Level, error) {
	switch strings.ToLower(strings.TrimSpace(input)) {
	case "requester", "request", "req":
		return LevelRequester, nil
	case "controller", "control":
		return LevelController, nil
	case "admin":
		return LevelAdmin, nil
	default:
		return LevelNone, fmt.Errorf("unknown permission level '%s'", input)
	}
}

func ResolveMessage(s *discordgo.Session, m *discordgo.MessageCreate) Context {
	userID := ""
	if m.Author != nil {
		userID = m.Author.ID
	}

	roleIDs := []string{}
	if m.Member != nil {
		roleIDs = append(roleIDs, m.Member.Roles...)
	}

	return resolve(s, m.GuildID, m.ChannelID, userID, roleIDs)
}

func ResolveInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) Context {
	userID := ""
	roleIDs := []string{}

	if i.Member != nil {
		roleIDs = append(roleIDs, i.Member.Roles...)
		if i.Member.User != nil {
			userID = i.Member.User.ID
		}
	}
	if userID == "" && i.User != nil {
		userID = i.User.ID
	}

	return resolve(s, i.GuildID, i.ChannelID, userID, roleIDs)
}

func resolve(s *discordgo.Session, guildID, channelID, userID string, roleIDs []string) Context {
	ctx := Context{
		GuildID: guildID,
		UserID:  userID,
	}

	if guildID == "" || userID == "" {
		return ctx
	}

	memberRoleIDs := roleIDs
	if len(memberRoleIDs) == 0 {
		member, err := memberForGuild(s, guildID, userID)
		if err == nil && member != nil {
			memberRoleIDs = append(memberRoleIDs, member.Roles...)
		}
	}

	userLevel, roleLevel, err := database.ResolveGuildPermissionLevel(guildID, userID, memberRoleIDs)
	if err == nil {
		ctx.ExplicitLevel = Level(userLevel)
		ctx.LinkedRoleLevel = Level(roleLevel)
		ctx.ResolvedLevel = maxLevel(ctx.ExplicitLevel, ctx.LinkedRoleLevel)
	}

	if guild, err := guildByID(s, guildID); err == nil && guild != nil && guild.OwnerID == userID {
		ctx.IsGuildOwner = true
		ctx.ResolvedLevel = maxLevel(ctx.ResolvedLevel, LevelAdmin)
	}

	if hasDiscordAdminPermission(s, userID, channelID) {
		ctx.HasDiscordAdmin = true
		ctx.ResolvedLevel = maxLevel(ctx.ResolvedLevel, LevelAdmin)
	}

	return ctx
}

func memberForGuild(s *discordgo.Session, guildID, userID string) (*discordgo.Member, error) {
	if s == nil {
		return nil, errors.New("discord session is nil")
	}

	if member, err := s.State.Member(guildID, userID); err == nil && member != nil {
		return member, nil
	}

	return s.GuildMember(guildID, userID)
}

func guildByID(s *discordgo.Session, guildID string) (*discordgo.Guild, error) {
	if s == nil {
		return nil, errors.New("discord session is nil")
	}

	if guild, err := s.State.Guild(guildID); err == nil && guild != nil {
		return guild, nil
	}

	return s.Guild(guildID)
}

func hasDiscordAdminPermission(s *discordgo.Session, userID, channelID string) bool {
	if s == nil || userID == "" || channelID == "" {
		return false
	}

	perms, err := s.UserChannelPermissions(userID, channelID)
	if err != nil {
		return false
	}

	adminMask := int64(discordgo.PermissionAdministrator | discordgo.PermissionManageGuild)
	return perms&adminMask != 0
}

func maxLevel(levels ...Level) Level {
	highest := LevelNone
	for _, level := range levels {
		if level > highest {
			highest = level
		}
	}
	return highest
}
