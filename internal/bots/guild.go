package bots

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrNoGuild = errors.New("no guild")
)

type GuildIDs map[string]string

func (g GuildIDs) GuildID(guild string) (string, error) {
	gid, ok := g[guild]
	if !ok {
		return "", ErrNoGuild
	}
	return gid, nil
}

func NewGuildIDs(sess *discordgo.Session) (GuildIDs, error) {
	res := GuildIDs{}
	var after string
	for {
		guilds, err := sess.UserGuilds(100, "", after)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrNoGuild, err)
		}

		if len(guilds) == 0 {
			return res, nil
		}

		for _, g := range guilds {
			res[g.Name] = g.ID
			after = g.ID
		}
	}
}
