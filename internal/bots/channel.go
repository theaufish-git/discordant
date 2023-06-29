package bots

import (
	"errors"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

var (
	ErrNoChannel = errors.New("no channel")
)

type ChannelIDs map[string]map[string]string

func (c ChannelIDs) ChannelID(category string, channel string) (string, error) {
	if _, ok := c[category]; !ok {
		return "", ErrNoChannel
	} else if cid, ok := c[category][channel]; !ok {
		return "", ErrNoChannel
	} else {
		return cid, nil
	}
}

func NewChannelIDs(sess *discordgo.Session, guildID string) (ChannelIDs, error) {
	channels, err := sess.GuildChannels(guildID)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNoChannel, err)
	}

	categories := map[string]string{}
	for _, channel := range channels {
		switch channel.Type {
		case discordgo.ChannelTypeGuildCategory:
			categories[channel.ID] = channel.Name
		}
	}

	groupedChannels := ChannelIDs{}
	for _, channel := range channels {
		switch channel.Type {
		case discordgo.ChannelTypeGuildText:
			if _, ok := groupedChannels[categories[channel.ParentID]]; !ok {
				groupedChannels[categories[channel.ParentID]] = map[string]string{}
			}
			groupedChannels[categories[channel.ParentID]][channel.Name] = channel.ID
		}
	}

	return groupedChannels, nil
}
