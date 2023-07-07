package dal

import (
	"context"

	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type Turg interface {
	Save(context.Context, string, *config.Turg) error
	Load(context.Context, string) (*config.Turg, error)
}
