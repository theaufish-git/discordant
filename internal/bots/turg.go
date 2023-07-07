package bots

import (
	"context"
	"fmt"
	"io"
	"log"
	"math/rand"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"github.com/theaufish-git/discordant/internal/dal"
	"github.com/theaufish-git/discordant/internal/gifs"
	"github.com/theaufish-git/discordant/internal/period"
)

const (
	turgCfgFile = "turg-config"
)

var (
	turgGifs = []string{
		"bless",
		"blessed",
		"you are blessed",
	}

	turgisms = []string{
		"did you remember you’re blessed?",
		"*you're blessed!*",
		"bless your attack rolls ya goofs",
		"4 level bless, starry form dragon",
	}
)

type Turg struct {
	Generic

	period *Signal[period.Period]
	pause  *Signal[bool]
	tmpHP  *Signal[int64]

	cfg *config.Turg
	tdb dal.Turg
	gdb gifs.Giffer
}

func NewTurg(tdb dal.Turg, gdb gifs.Giffer, cfg *config.Turg) (*Turg, error) {
	return &Turg{
		Generic: Generic{
			bot:          cfg.Bot,
			target:       cfg.Target,
			permissions:  cfg.Permissions,
			guild:        cfg.Guild,
			allowMembers: map[string]struct{}{},
			allowRoles:   map[string]struct{}{},
			handlers:     map[string]Command{},
		},
		cfg: cfg,
		tdb: tdb,
		gdb: gdb,
	}, nil
}

func (t *Turg) ID() string {
	return "turg"
}

func (t *Turg) Initialize(ctx context.Context) error {
	if err := t.Generic.Initialize(ctx); err != nil {
		return err
	}

	cfg, err := t.tdb.Load(ctx, turgCfgFile)
	if err != nil {
		return err
	}

	if cfg != nil {
		t.cfg = cfg
	}

	t.period = NewSignal[period.Period]()
	t.period.Value = period.NewRandom(t.cfg.Period.Min, t.cfg.Period.Max)
	t.pause = NewSignal[bool]()
	t.pause.Value = t.cfg.Pause
	t.tmpHP = NewSignal[int64]()
	t.tmpHP.Value = t.cfg.TempHP

	// create command
	_, err = t.ApplicationCommandCreate(t.State.User.ID, t.GID(), &discordgo.ApplicationCommand{
		Name:        "turg",
		Description: "Turg-o-tron commands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pause",
				Description: "Turg-o-tron should either pause or resume speaking.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Description: "Whether to keep speaking.",
						Name:        "pause",
						Required:    true,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set-ism-period",
				Description: "Turg-o-tron should speak once randomly between min and max.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Description: "Minimum amount of time between speaking.",
						Name:        "min-period",
						Required:    true,
					},
					{
						Type:        discordgo.ApplicationCommandOptionString,
						Description: "Maximum amount of time between speaking.",
						Name:        "max-period",
						Required:    false,
					},
				},
			},
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "set-tmp-hp",
				Description: "Turg-o-tron should say there is this much tmp hp every minute.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Description: "The current tmp hp value.",
						Name:        "tmp-hp",
						Required:    true,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	t.Generic.WithHandler("pause", t.handlePause)
	t.Generic.WithHandler("set-ism-period", t.handleSetIsmPeriod)
	t.Generic.WithHandler("set-tmp-hp", t.handleSetTmpHP)
	return nil
}

func (t *Turg) Shutdown(ctx context.Context) error {
	t.period.Close()
	t.pause.Close()
	t.tmpHP.Close()

	return t.Generic.Shutdown(ctx)
}

func (t *Turg) Run(ctx context.Context) error {
	var needsSave bool
	cfgTicker := time.NewTicker(5 * time.Minute)
	ismTicker := time.NewTicker(t.period.Value.Period())
	defer ismTicker.Stop()
	tmpHPTicker := time.NewTicker(t.period.Value.Period())
	defer tmpHPTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case period := <-t.period.C():
			ismTicker.Reset(period.Period())
			tmpHPTicker.Reset(period.Period())

			needsSave = true
			t.cfg.Period.Max = period.Max()
			t.cfg.Period.Min = period.Min()
			t.period.Value = period
		case pause := <-t.pause.C():
			if t.pause.Value != pause {
				if pause {
					t.ChannelMessageSend(t.CID(), "Turg-o-tron is taking a break.")
				} else {
					t.ChannelMessageSend(t.CID(), "Turg-o-tron activated. Form of starry dragon!")
				}
			} else {
				continue
			}

			needsSave = true
			t.cfg.Pause = pause
			t.pause.Value = pause
		case tmpHP := <-t.tmpHP.C():
			needsSave = true
			t.cfg.TempHP = tmpHP
			t.tmpHP.Value = tmpHP
		case <-cfgTicker.C:
			if !needsSave {
				continue
			}

			if err := t.tdb.Save(ctx, turgCfgFile, t.cfg); err != nil {
				log.Println("could not save config:", err)
			}
			needsSave = false
		case <-ismTicker.C:
			ismTicker.Reset(t.period.Value.Period())
			if t.pause.Value {
				continue
			}

			x := rand.Intn(10)
			switch x {
			case 0, 1, 2:
				gifURL, err := t.gdb.Gif(turgGifs[x])
				if err != nil {
					log.Println("cannot find gif:", err)
					continue
				}
				t.ChannelMessageSend(t.CID(), gifURL)
			default:
				ism := turgisms[rand.Intn(len(turgisms))]
				t.ChannelMessageSend(t.CID(), ism)
			}
		case <-tmpHPTicker.C:
			tmpHPTicker.Reset(t.period.Value.Period())
			if t.pause.Value || t.tmpHP.Value == 0 {
				continue
			}

			t.ChannelMessageSend(t.CID(), fmt.Sprintf("dont forget your temp hp! (%v)", t.tmpHP.Value))
		}
	}
}

func (t *Turg) handlePause(sess *discordgo.Session, subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	p := findOption(subcmd.Options, "pause")
	pause := p.BoolValue()

	t.pause.C() <- pause

	resp.Write([]byte(fmt.Sprintf("pause set to: %v", pause)))
	return nil
}

func (t *Turg) handleSetIsmPeriod(sess *discordgo.Session, subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	max := findOption(subcmd.Options, "max-period")
	min := findOption(subcmd.Options, "min-period")

	if max == nil {
		max = min
	}

	maxDur, err := time.ParseDuration(max.StringValue())
	if err != nil {
		return err
	}

	minDur, err := time.ParseDuration(min.StringValue())
	if err != nil {
		return err
	}

	t.period.C() <- period.NewRandom(minDur, maxDur)

	resp.Write([]byte(fmt.Sprintf("speak periods set to: %v - %v", minDur, maxDur)))
	return nil
}

func (t *Turg) handleSetTmpHP(sess *discordgo.Session, subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	tmpHP := findOption(subcmd.Options, "tmp-hp").IntValue()

	t.tmpHP.C() <- tmpHP

	resp.Write([]byte(fmt.Sprintf("tmp hp set to: %v", tmpHP)))
	return nil
}
