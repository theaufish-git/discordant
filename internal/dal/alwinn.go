package dal

import (
	"context"

	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type Alwinn interface {
	Save(context.Context, string, *config.Alwinn) error
	Load(context.Context, string) (*config.Alwinn, error)
}
