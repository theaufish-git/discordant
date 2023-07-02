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
	"github.com/theaufish-git/discordant/internal/gifs"
	"github.com/theaufish-git/discordant/internal/period"
)

var (
	dies = map[int64]struct{}{
		4:   {},
		6:   {},
		8:   {},
		10:  {},
		12:  {},
		20:  {},
		100: {},
	}

	alwinnGifs = []string{
		"silverybarbs",
		"counterspell",
		"inspiration",
		"you're inspired",
	}

	alwinnisms = []string{
		"advantage because ~fairy fire~!",
		"and I will inspire uh...",
	}
)

type Alwinn struct {
	Generic

	period         *Signal[period.Period]
	pause          *Signal[bool]
	inspirationDie *Signal[int64]

	cfg *config.Alwinn
	gdb gifs.Giffer
}

func NewAlwinn(gdb gifs.Giffer, cfg *config.Alwinn) (*Alwinn, error) {
	return &Alwinn{
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
		gdb: gdb,
	}, nil
}

func (a *Alwinn) ID() string {
	return "alwinn"
}

func (a *Alwinn) Initialize(ctx context.Context) error {
	if err := a.Generic.Initialize(ctx); err != nil {
		return err
	}

	a.period = NewSignal[period.Period]()
	a.period.Value = period.NewRandom(a.cfg.Period.Min, a.cfg.Period.Max)
	a.pause = NewSignal[bool]()
	a.inspirationDie = NewSignal[int64]()
	a.inspirationDie.Value = a.cfg.InspirationDie

	// create command
	_, err := a.ApplicationCommandCreate(a.State.User.ID, a.GID(), &discordgo.ApplicationCommand{
		Name:        "alwinn",
		Description: "Alwinn-ator commands",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionSubCommand,
				Name:        "pause",
				Description: "Alwinn-ator should either pause or resume speaking.",
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
				Description: "Alwinn-ator should speak once randomly between min and max.",
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
				Name:        "set-inspiration-die",
				Description: "Alwinn-ator should say this is the inspiration die value.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionInteger,
						Description: "The current inspiration die value.",
						Name:        "inspiration-die",
						Required:    true,
					},
				},
			},
		},
	})
	if err != nil {
		return err
	}

	a.Generic.WithHandler("pause", a.handlePause)
	a.Generic.WithHandler("set-ism-period", a.handleSetIsmPeriod)
	a.Generic.WithHandler("set-inspiration-die", a.handleSetInspirationDie)
	return nil
}

func (a *Alwinn) Shutdown(ctx context.Context) error {
	a.period.Close()
	a.pause.Close()
	a.inspirationDie.Close()

	return a.Generic.Shutdown(ctx)
}

func (a *Alwinn) Run(ctx context.Context) error {
	ismTicker := time.NewTicker(a.period.Value.Period())
	defer ismTicker.Stop()
	inspirationDieTicker := time.NewTicker(a.period.Value.Period())
	defer inspirationDieTicker.Stop()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case period := <-a.period.C():
			a.period.Value = period
			ismTicker.Reset(a.period.Value.Period())
			inspirationDieTicker.Reset(a.period.Value.Period())
		case pause := <-a.pause.C():
			if a.pause.Value != pause {
				if pause {
					a.ChannelMessageSend(a.CID(), "Alwinn-ator is taking a nap.")
				} else {
					a.ChannelMessageSend(a.CID(), "Alwinn-ator activated. Theres a barb that needs silvering!")
				}
			}
			a.pause.Value = pause
		case inspirationDie := <-a.inspirationDie.C():
			a.inspirationDie.Value = inspirationDie
		case <-ismTicker.C:
			ismTicker.Reset(a.period.Value.Period())
			if a.pause.Value {
				continue
			}

			x := rand.Intn(3)
			switch x {
			case 0, 1:
				gifURL, err := a.gdb.Gif(alwinnGifs[x])
				if err != nil {
					log.Println("cannot find gif:", err)
					continue
				}
				a.ChannelMessageSend(a.CID(), gifURL)
			default:
				ism := alwinnisms[rand.Intn(len(alwinnisms))]
				a.ChannelMessageSend(a.CID(), ism)
			}
		case <-inspirationDieTicker.C:
			inspirationDieTicker.Reset(a.period.Value.Period())
			if a.pause.Value {
				continue
			}

			a.ChannelMessageSend(a.CID(), fmt.Sprintf("do you want to use your inspiration? it's a d%d", a.inspirationDie.Value))
		}
	}
}

func (a *Alwinn) handlePause(sess *discordgo.Session, subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	p := findOption(subcmd.Options, "pause")
	pause := p.BoolValue()

	a.pause.C() <- pause

	resp.Write([]byte(fmt.Sprintf("pause set to: %v", pause)))
	return nil
}

func (a *Alwinn) handleSetIsmPeriod(sess *discordgo.Session, subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
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

	a.period.C() <- period.NewRandom(minDur, maxDur)

	resp.Write([]byte(fmt.Sprintf("speak periods set to: %v - %v", minDur, maxDur)))
	return nil
}

func (a *Alwinn) handleSetInspirationDie(sess *discordgo.Session, subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	inspirationDie := findOption(subcmd.Options, "inspiration-die").IntValue()
	if _, ok := dies[inspirationDie]; !ok {
		return fmt.Errorf("not a vaild die: %d", inspirationDie)
	}

	a.inspirationDie.C() <- inspirationDie

	resp.Write([]byte(fmt.Sprintf("inspiration die set to: %v", inspirationDie)))
	return nil
}
