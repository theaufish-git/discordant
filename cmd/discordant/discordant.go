package main

import (
	"context"
	"log"
	"sync"

	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"github.com/theaufish-git/discordant/internal/bots"
	"github.com/theaufish-git/discordant/internal/dal"
	"github.com/theaufish-git/discordant/internal/dal/googlestorage"
	"github.com/theaufish-git/discordant/internal/gifs"
)

func mustReturn[T any](x T, err error) T {
	if err != nil {
		log.Fatalln(err)
	}
	return x
}

func main() {
	cfg := mustReturn(config.NewDiscordantFromEnv("dsc"))

	var gdb gifs.Giffer
	switch cfg.Gifs.Driver {
	case "giphy":
		gdb = gifs.NewGiphy(cfg.Gifs.Token)
	case "tenor":
		gdb = gifs.NewTenor(cfg.Gifs.Token)
	default:
		log.Fatal("invalid giffer driver:", cfg.Gifs.Driver)
	}

	ctx, cancel := context.WithCancel(context.Background())

	var adb dal.Alwinn
	var tdb dal.Turg
	var err error
	switch cfg.DAL.Driver {
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
		log.Fatal("invalid dal driver:", cfg.DAL.Driver)
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
