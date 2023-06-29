package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Discordant struct {
	Gifs Gifs
	Turg Turg
}

type Gifs struct {
	Driver     string `default:"giphy"`
	GiphyToken string `split_words:"true" required:"true"`
}

type Bot struct {
	Token string `required:"true"`
}

type Target struct {
	Category string `required:"true"`
	Channel  string `required:"true"`
}

type Permissions struct {
	Roles   []string `required:"true"`
	Members []string `required:"true"`
}

type Period struct {
	Min time.Duration `default:"1m"`
	Max time.Duration `default:"5m"`
}

type Turg struct {
	Bot         Bot
	Guild       string `required:"true"`
	Target      Target
	Permissions Permissions
	Period      Period
}

func NewDiscordantFromEnv(prefix string) (*Discordant, error) {
	obj := &Discordant{}
	if err := envconfig.Process(prefix, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
