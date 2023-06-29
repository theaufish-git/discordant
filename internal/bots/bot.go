package bots

import (
	"context"

	"github.com/bwmarrin/discordgo"
)

type Bot interface {
	ID() string
	Initialize(context.Context) error
	Run(context.Context) error
	Shutdown(context.Context) error
}

func findOption(options []*discordgo.ApplicationCommandInteractionDataOption, name string) *discordgo.ApplicationCommandInteractionDataOption {
	for _, option := range options {
		if option.Name == name {
			return option
		} else if len(option.Options) > 0 {
			if suboption := findOption(option.Options, name); suboption != nil {
				return suboption
			}
		}
	}
	return nil
}
