package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Discordant struct {
	DAL    DAL
	Gifs   Gifs
	Alwinn Alwinn
	Turg   Turg
}

type Gifs struct {
	Driver string `required:"true"`
	Token  string `split_words:"true" required:"true"`
}

type DAL struct {
	Driver        string `default:"googlestorage"`
	GoogleStorage GoogleStorage
}

type GoogleStorage struct {
	Bucket string `default:"discordant"`
	Key    string `default:"/tmp/discordant-storage-rw.json"`
}

type Bot struct {
	Token       string `required:"true"`
	Guild       string `required:"true"`
	Target      Target
	Permissions Permissions
}

type IsmBot struct {
	Bot

	Period Period
	Pause  bool `default:"true"`
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

type Alwinn struct {
	IsmBot
	InspirationDie int64 `split_words:"true" default:"4"`
}

type Turg struct {
	IsmBot
	TempHP int64 `split_words:"true" default:"0"`
}

func NewDiscordantFromEnv(prefix string) (*Discordant, error) {
	obj := &Discordant{}
	if err := envconfig.Process(prefix, obj); err != nil {
		return nil, err
	}

	return obj, nil
}
