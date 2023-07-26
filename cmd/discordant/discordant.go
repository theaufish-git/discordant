package main

import (
	"context"

	"github.com/rleszilm/genms/logging"
	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"github.com/theaufish-git/discordant/internal/bots"
	"github.com/theaufish-git/discordant/internal/dal"
	"github.com/theaufish-git/discordant/internal/dal/gif"
	"github.com/theaufish-git/discordant/internal/dal/googlestorage"
	"github.com/theaufish-git/discordant/internal/dal/static"
)

var (
	logs = logging.NewChannel("discordant")
)

func main() {
	logs.Print("starting discordant")
	logs.Print("parsing config")
	cfg, err := config.NewFromEnv[config.Discordant]("dsc")
	if err != nil {
		logs.Fatal("cannot parse config:", err)
	}

	manager := service.NewManager()

	logs.Print("creating gif search driver")
	var gdb dal.Gif
	switch cfg.DAL.GifDriver {
	case "giphy":
		giphy, err := config.NewFromEnv[config.Gif]("dsc_dal_giphy")
		if err != nil {
			logs.Fatal("cannot parse config:", err)
		}

		gdb = gif.NewGiphy(giphy)
	case "tenor":
		tenor, err := config.NewFromEnv[config.Gif]("dsc_dal_tenor")
		if err != nil {
			logs.Fatal("cannot parse config:", err)
		}

		gdb = gif.NewTenor(tenor)
	default:
		logs.Fatal("invalid gif driver:", cfg.DAL.GifDriver)
	}
	manager.Register(gdb)

	ctx := context.Background()

	logs.Print("creating bot storage drivers")
	var adb dal.Alwinn
	var tdb dal.Turg
	switch cfg.DAL.ConfigDriver {
	case "staticstorage":
		ss, err := config.NewFromEnv[config.StaticStorage]("dsc_dal_static_storage")
		if err != nil {
			logs.Fatal("cannot parse config:", err)
		}

		adb, err = static.NewJSON[config.Alwinn](ctx, ss)
		if err != nil {
			logs.Fatal("cannot create google storage for alwinn db:", err)
		}

		tdb, err = static.NewJSON[config.Turg](ctx, ss)
		if err != nil {
			logs.Fatal("cannot create google storage for turg db:", err)
		}
	case "googlestorage":
		gs, err := config.NewFromEnv[config.GoogleCloudStorage]("dsc_dal_google_storage")
		if err != nil {
			logs.Fatal("cannot parse config:", err)
		}

		adb, err = googlestorage.NewJSON[config.Alwinn](ctx, gs)
		if err != nil {
			logs.Fatal("cannot create google storage for alwinn db:", err)
		}

		tdb, err = googlestorage.NewJSON[config.Turg](ctx, gs)
		if err != nil {
			logs.Fatal("cannot create google storage for turg db:", err)
		}
	default:
		logs.Fatal("invalid dal driver:", cfg.DAL.ConfigDriver)
	}
	manager.Register(adb, tdb)

	logs.Print("creating alwinn")
	alBot, err := bots.NewAlwinn(adb, gdb, &cfg.Alwinn)
	if err != nil {
		logs.Fatal("could not create Alwinn:", err)
	}

	logs.Print("creating turg")
	tuBot, err := bots.NewTurg(tdb, gdb, &cfg.Turg)
	if err != nil {
		logs.Fatal("could not create Turg:", err)
	}

	logs.Print("starting bots")
	manager.Register(alBot, tuBot)
	if err := manager.Initialize(ctx); err != nil {
		logs.Fatal("could not initialize service:", err)
	}

	logs.Print("waiting")
	manager.Wait()
	logs.Print("shutting down bots")
	if err := manager.Shutdown(ctx); err != nil {
		logs.Fatal("could not shutdown service:", err)
	}
}
