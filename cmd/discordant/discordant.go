package main

import (
	"context"
	"log"
	"sync"

	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"github.com/theaufish-git/discordant/internal/bots"
	"github.com/theaufish-git/discordant/internal/dal"
	"github.com/theaufish-git/discordant/internal/dal/gif"
	"github.com/theaufish-git/discordant/internal/dal/googlestorage"
)

func mustReturn[T any](x T, err error) T {
	if err != nil {
		log.Fatalln(err)
	}
	return x
}

func main() {
	cfg := mustReturn(config.NewDiscordantFromEnv("dsc"))

	var gdb dal.Gif
	switch cfg.DAL.GifDriver {
	case "giphy":
		gdb = gif.NewGiphy(cfg.DAL.Gif.Token)
	case "tenor":
		gdb = gif.NewTenor(cfg.DAL.Gif.Token)
	default:
		log.Fatal("invalid giffer driver:", cfg.DAL.GifDriver)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var adb dal.Alwinn
	var tdb dal.Turg
	var err error
	switch cfg.DAL.ConfigDriver {
	case "googlestorage":
		adb, err = googlestorage.NewJSON[config.Alwinn](ctx, &cfg.DAL.GoogleStorage)
		if err != nil {
			log.Fatal("cannot create google storage for alwinn db:", err)
		}

		tdb, err = googlestorage.NewJSON[config.Turg](ctx, &cfg.DAL.GoogleStorage)
		if err != nil {
			log.Fatal("cannot create google storage for turg db:", err)
		}
	default:
		log.Fatal("invalid dal driver:", cfg.DAL.ConfigDriver)
	}

	bs := []bots.Bot{
		mustReturn(bots.NewAlwinn(adb, gdb, &cfg.Alwinn)),
		mustReturn(bots.NewTurg(tdb, gdb, &cfg.Turg)),
	}

	wg := sync.WaitGroup{}
	for _, bot := range bs {
		if err := bot.Initialize(ctx); err != nil {
			log.Fatalf("err starting %s: %v\n", bot.ID(), err)
		}

		wg.Add(1)
		go func(bot bots.Bot) {
			defer wg.Done()
			defer cancel()

			if err := bot.Run(ctx); err != nil {
				log.Printf("completed %s: %v", bot.ID(), err)
			}
			bot.Shutdown(ctx)
		}(bot)
	}

	wg.Wait()
}
