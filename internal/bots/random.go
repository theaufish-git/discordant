package bots

import (
	"context"
	"math/rand"
)

type Action[T any, S any] func(context.Context, T, S) error

type RandomAction[T any, S any] struct {
	Weight  int64
	Buckets []*RandomActionBucket[T, S]
}

func (r *RandomAction[T, S]) WithBucket(b *RandomActionBucket[T, S]) {
	r.Buckets = append(r.Buckets, b)
	r.Weight += b.Weight
}

func (r *RandomAction[T, S]) Call(ctx context.Context) error {
	x := rand.Int63n(r.Weight)
	for _, b := range r.Buckets {
		x -= b.Weight
		if b.Weight <= 0 {
			return b.Call(ctx)
		}
	}
	return nil
}

type RandomActionBucket[T any, S any] struct {
	Action func(context.Context, T, S) error
	Weight int64
	State  S
	Values []T
}

func (r *RandomActionBucket[T, S]) Call(ctx context.Context) error {
	x := rand.Int63n(r.Weight)
	return r.Action(ctx, r.Values[x], r.State)
}
