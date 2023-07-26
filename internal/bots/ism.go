package bots

import (
	"context"
	"fmt"
	"time"

	"github.com/rleszilm/genms/logging"
	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"github.com/theaufish-git/discordant/internal/dal"
	"github.com/theaufish-git/discordant/internal/period"
)

type Ism struct {
	service.UnimplementedService

	sess    *Session
	logs    *logging.Channel
	id      string
	period  period.Period
	periods chan period.Period
	post    bool
	posts   chan bool
	halt    chan struct{}
	ra      *RandomAction[string, map[string]struct{}]

	gdb dal.Gif
}

func NewIsm(id string, bot *config.Bot, gdb dal.Gif, sigPeriod *Signal[period.Period], sigPause *Signal[bool]) *Ism {
	fmt.Println("sub periods")
	periods := make(chan period.Period, 1)
	sigPeriod.Subscribe(periods)

	fmt.Println("sub posts")

	posts := make(chan bool, 1)
	sigPause.Subscribe(posts)

	fmt.Println("create")

	i := &Ism{
		sess:    NewSession(bot),
		id:      id,
		period:  <-periods,
		periods: periods,
		post:    <-posts,
		posts:   posts,
		halt:    make(chan struct{}),
		ra:      &RandomAction[string, map[string]struct{}]{},

		gdb: gdb,
	}
	i.logs = logging.NewChannel(i.String())
	i.WithDependencies(i.sess)
	return i
}

func (i *Ism) Initialize(ctx context.Context) error {
	go i.Run(ctx)

	return nil
}

func (i *Ism) Shutdown(ctx context.Context) error {
	close(i.halt)

	return nil
}

func (i *Ism) String() string {
	return "ism-" + i.id
}

func (i *Ism) Run(ctx context.Context) error {
	i.logs.Print("ism running...")
	defer i.logs.Print("ism halting...")

	ismTicker := time.NewTicker(time.Minute)
	if i.period != nil {
		ismTicker = time.NewTicker(i.period.Period())
		defer ismTicker.Stop()
	}
	for {
		select {
		case <-i.halt:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		case p, more := <-i.periods:
			if !more {
				i.periods = nil
				continue
			}

			if i.period == nil {
				i.period = p
				ismTicker = time.NewTicker(i.period.Period())
				defer ismTicker.Stop()
			}

			ismTicker.Reset(i.period.Period())
		case p, more := <-i.posts:
			if !more {
				i.posts = nil
				continue
			}
			i.post = p

			if i.period == nil {
				continue
			}

			if !i.post {
				ismTicker.Reset(time.Hour)
			} else {
				ismTicker.Reset(i.period.Period())
			}
		case <-ismTicker.C:
			if !i.post {
				continue
			}

			if err := i.ra.Call(ctx); err != nil {
				i.logs.Print("error when calling random ism:", err)
			}

			ismTicker.Reset(i.period.Period())
		}
	}
}

func (i *Ism) WithBucket(action Action[string, map[string]struct{}], weight int64, state map[string]struct{}, values ...string) {
	i.ra.WithBucket(&RandomActionBucket[string, map[string]struct{}]{
		Action: action,
		Weight: weight,
		State:  state,
		Values: values,
	})
}

func (i *Ism) PostGif(ctx context.Context, gif string, subscribers map[string]struct{}) error {
	gifURL, err := i.gdb.Fetch(ctx, gif)
	if err != nil {
		i.logs.Print("cannot find gif:", err)
		return err
	}

	for cid := range subscribers {
		if _, err = i.sess.ChannelMessageSend(cid, gifURL); err != nil {
			i.logs.Print("could not post gif:", err)
		}
	}
	return nil
}

func (i *Ism) PostIsm(ctx context.Context, ism string, subscribers map[string]struct{}) error {
	for cid := range subscribers {
		if _, err := i.sess.ChannelMessageSend(cid, ism); err != nil {
			i.logs.Print("could not post ism:", err)
		}
	}
	return nil
}
