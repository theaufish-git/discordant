package dal

import (
	"context"

	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type Turg interface {
	service.Service
	Save(context.Context, string, *config.Turg) error
	Load(context.Context, string) (*config.Turg, error)
}
