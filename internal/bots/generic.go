package bots

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type Command func(*discordgo.Session, *discordgo.ApplicationCommandInteractionDataOption, *discordgo.InteractionCreate, io.Writer) error

type Generic struct {
	*discordgo.Session

	bot         config.Bot
	target      config.Target
	permissions config.Permissions

	guild        string
	gid          string
	cid          string
	allowMembers map[string]struct{}
	allowRoles   map[string]struct{}

	roles    RoleIDs
	channels ChannelIDs
	members  *MemberIDs

	handlers map[string]Command
}

func (g *Generic) Initialize(ctx context.Context) error {
	dg, err := discordgo.New("Bot " + g.bot.Token)
	if err != nil {
		return err
	}
	g.Session = dg

	err = dg.Open()
	if err != nil {
		return err
	}

	gids, err := NewGuildIDs(dg)
	if err != nil {
		return err
	}

	gid, err := gids.GuildID(g.guild)
	if err != nil {
		return err
	}
	g.gid = gid

	cids, err := NewChannelIDs(dg, gid)
	if err != nil {
		return err
	}
	g.channels = cids

	cid, err := cids.ChannelID(g.target.Category, g.target.Channel)
	if err != nil {
		return err
	}
	g.cid = cid

	mids, err := NewMemberIDs(dg, gid)
	if err != nil {
		return err
	}
	g.members = mids

	rids, err := NewRoleIDs(dg, gid)
	if err != nil {
		return err
	}
	g.roles = rids

	// setup allow sets
	for _, m := range g.permissions.Members {
		mid, err := g.members.MemberID(m)
		if err != nil {
			continue
		}
		g.allowMembers[mid] = struct{}{}
	}

	for _, r := range g.permissions.Roles {
		rid, err := g.roles.RoleID(r)
		if err != nil {
			continue
		}
		g.allowRoles[rid] = struct{}{}
	}

	g.AddHandler(g.onInteractionCreate)
	return nil
}

func (g *Generic) Shutdown(ctx context.Context) error {
	g.Close()
	return nil
}

func (g *Generic) CID() string {
	return g.cid
}

func (g *Generic) GID() string {
	return g.gid
}

func (g *Generic) WithHandler(trigger string, cmd Command) {
	g.handlers[trigger] = cmd
}

func (g *Generic) onInteractionCreate(sess *discordgo.Session, msg *discordgo.InteractionCreate) {
	if !g.allow(msg) {
		sess.InteractionRespond(msg.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "https://tenor.com/view/jurassic-park-ah-you-didnt-say-the-magic-word-say-please-gif-9628120",
				Flags:   discordgo.MessageFlagsEphemeral,
			},
		})
		return
	}

	switch msg.Type {
	case discordgo.InteractionApplicationCommand:
		data := msg.ApplicationCommandData()
		for _, subcmd := range data.Options {
			resp := &bytes.Buffer{}
			handler, ok := g.handlers[subcmd.Name]
			if !ok {
				sess.InteractionRespond(msg.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "not implemented",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				continue
			}

			if err := handler(sess, subcmd, msg, resp); err != nil {
				sess.InteractionRespond(msg.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("error: %+v", err),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				continue
			}

			sess.InteractionRespond(msg.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: resp.String(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	}
}

func (g *Generic) allow(msg *discordgo.InteractionCreate) bool {
	user := msg.User
	if user == nil {
		user = msg.Member.User
	}

	_, allow := g.allowMembers[user.ID]

	roles, _ := g.members.RoleIDs(user.ID)
	for _, role := range roles {
		_, ok := g.allowRoles[role]
		allow = allow || ok
	}

	return allow
}
