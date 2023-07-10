package googlestorage

import (
	"bytes"
	"context"
	"encoding/json"

	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	"google.golang.org/api/storage/v1"
)

type JSON[T any] struct {
	svc *storage.Service
	cfg *config.GoogleStorage
}

func NewJSON[T any](ctx context.Context, cfg *config.GoogleStorage) (*JSON[T], error) {
	svc, err := storage.NewService(ctx, option.WithCredentialsFile(cfg.Key))
	if err != nil {
		return nil, err
	}

	return &JSON[T]{
		svc: svc,
		cfg: cfg,
	}, nil
}

func (j *JSON[T]) Load(ctx context.Context, id string) (*T, error) {
	resp, err := j.svc.Objects.Get(j.cfg.Bucket, id).Download()
	if err != nil {
		if apiErr, ok := err.(*googleapi.Error); ok && apiErr.Code == 404 {
			return nil, nil
		}
		return nil, err
	}
	defer resp.Body.Close()

	var obj T
	if err := json.NewDecoder(resp.Body).Decode(&obj); err != nil {
		return nil, err
	}

	return &obj, nil
}

func (j *JSON[T]) Save(ctx context.Context, id string, cfg *T) error {
	media := bytes.Buffer{}
	if err := json.NewEncoder(&media).Encode(cfg); err != nil {
		return err
	}

	call := j.svc.Objects.Insert(j.cfg.Bucket, &storage.Object{Name: id}).Media(&media)

	if _, err := call.Do(); err != nil {
		return err
	}
	return nil
}
