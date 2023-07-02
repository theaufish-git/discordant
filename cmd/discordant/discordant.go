package main

import (
	"context"
	"log"
	"sync"

	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"github.com/theaufish-git/discordant/internal/bots"
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
	}

	bs := []bots.Bot{
		mustReturn(bots.NewAlwinn(gdb, &cfg.Alwinn)),
		mustReturn(bots.NewTurg(gdb, &cfg.Turg)),
	}

	ctx, cancel := context.WithCancel(context.Background())

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
