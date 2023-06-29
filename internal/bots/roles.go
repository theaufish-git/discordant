package bots

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrNoRole = errors.New("no role")
)

type RoleIDs map[string]string

func (r RoleIDs) RoleID(role string) (string, error) {
	rid, ok := r[role]
	if !ok {
		return "", ErrNoRole
	}
	return rid, nil
}

func NewRoleIDs(sess *discordgo.Session, guildID string) (map[string]string, error) {
	roles, err := sess.GuildRoles(guildID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoRole, err)
	}

	res := map[string]string{}
	for _, r := range roles {
		res[r.Name] = r.ID
	}
	return res, nil
}
