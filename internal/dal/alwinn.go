package dal

import (
	"context"

	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type Alwinn interface {
	service.Service
	Save(context.Context, string, *config.Alwinn) error
	Load(context.Context, string) (*config.Alwinn, error)
}
