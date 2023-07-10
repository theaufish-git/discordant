package dal

import "context"

type Gif interface {
	Fetch(context.Context, string) (string, error)
}
