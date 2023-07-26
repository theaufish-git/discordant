package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Discordant struct {
	DAL    DAL
	Alwinn Alwinn
	Turg   Turg
}

type DAL struct {
	ConfigDriver string `split_words:"true" required:"true"`
	GifDriver    string `split_words:"true" required:"true"`
}

type StaticStorage struct {
	Dir string `required:"true"`
}

type GoogleCloudStorage struct {
	Bucket string `required:"true"`
	Key    string `required:"true"`
}

type Gif struct {
	Token string `split_words:"true" required:"true"`
}

type Bot struct {
	Token string `required:"true" json:"-"`
}

type IsmBot struct {
	Bot

	Period      Period
	Pause       bool                `default:"true"`
	Subscribers map[string]struct{} `default:""`
}

type Permissions struct {
	Roles   []string `required:"true"`
	Members []string `required:"true"`
}

type Period struct {
	Min      time.Duration `default:"1m"`
	MinLimit time.Duration `split_words:"true" default:"1m"`
	Max      time.Duration `default:"5m"`
	MaxLimit time.Duration `split_words:"true" default:"5m"`
}

type Alwinn struct {
	IsmBot
	InspirationDie int64 `split_words:"true" default:"4"`
}

type Turg struct {
	IsmBot
	TempHP int64 `split_words:"true" default:"0"`
}

func NewFromEnv[T any](prefix string) (*T, error) {
	var obj T
	if err := envconfig.Process(prefix, &obj); err != nil {
		return nil, err
	}

	return &obj, nil
}
