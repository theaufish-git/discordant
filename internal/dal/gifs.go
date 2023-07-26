package dal

import (
	"context"

	"github.com/rleszilm/genms/service"
)

type Gif interface {
	service.Service
	Fetch(context.Context, string) (string, error)
}
