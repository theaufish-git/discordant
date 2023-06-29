package bots

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrNoMember = errors.New("no member")
)

type MemberIDs struct {
	ids   map[string]string
	roles map[string][]string
}

func (m MemberIDs) MemberID(member string) (string, error) {
	mid, ok := m.ids[member]
	if !ok {
		return "", ErrNoMember
	}
	return mid, nil
}

func (m MemberIDs) RoleIDs(memberID string) ([]string, error) {
	roles, ok := m.roles[memberID]
	if !ok {
		return nil, ErrNoMember
	}
	return roles, nil
}

func NewMemberIDs(sess *discordgo.Session, guildID string) (*MemberIDs, error) {
	res := &MemberIDs{
		ids:   map[string]string{},
		roles: map[string][]string{},
	}

	var after string
	for {
		members, err := sess.GuildMembers(guildID, after, 1000)
		if err != nil {
			return nil, fmt.Errorf("%w: %v", ErrNoMember, err)
		}

		if len(members) == 0 {
			break
		}

		for _, m := range members {
			res.ids[m.User.Username] = m.User.ID
			res.ids[m.Nick] = m.User.ID
			res.roles[m.User.ID] = m.Roles

			after = m.User.ID
		}
	}

	return res, nil
}
