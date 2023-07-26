package static

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path"

	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type JSON[T any] struct {
	service.UnimplementedService
	cfg *config.StaticStorage
}

func NewJSON[T any](ctx context.Context, cfg *config.StaticStorage) (*JSON[T], error) {
	return &JSON[T]{
		cfg: cfg,
	}, nil
}

func (j *JSON[T]) Initialize(_ context.Context) error {
	return nil
}

func (j *JSON[T]) Shutdown(_ context.Context) error {
	return nil
}

func (j *JSON[T]) String() string {
	return "dal-json-" + j.cfg.Dir
}

func (j *JSON[T]) Load(ctx context.Context, id string) (*T, error) {
	f, err := os.Open(path.Join(j.cfg.Dir, id))
	if err != nil && errors.Is(err, os.ErrNotExist) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}
	defer f.Close()

	var obj T
	if err := json.NewDecoder(f).Decode(&obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

func (j *JSON[T]) Save(ctx context.Context, id string, cfg *T) error {
	f, err := os.Create(path.Join(j.cfg.Dir, id))
	if err != nil {
		return err
	}
	defer f.Close()

	return json.NewEncoder(f).Encode(cfg)
}
