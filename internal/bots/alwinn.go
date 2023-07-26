package bots

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/rleszilm/genms/logging"
	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
	"github.com/theaufish-git/discordant/internal/dal"
	"github.com/theaufish-git/discordant/internal/period"
)

const (
	alwinnCfgFile = "alwinn-config"
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

	alwinnInspirtationDieFmts = []string{
		"do you want to use your inspiration? it's a d%d",
	}

	alwinnGifs = []string{
		"silverybarbs",
		"counterspell",
		"inspiration",
		"you're inspired",
	}

	alwinnIsms = []string{
		"advantage because ~fairy fire~!",
		"and I will inspire uh...",
	}
)

type Alwinn struct {
	service.UnimplementedService

	ism            *Ism
	logs           *logging.Channel
	period         *Signal[period.Period]
	pause          *Signal[bool]
	inspirationDie *Signal[int64]

	needsSave bool
	cfg       *config.Alwinn
	adb       dal.Alwinn
	gdb       dal.Gif
}

func NewAlwinn(adb dal.Alwinn, gdb dal.Gif, cfg *config.Alwinn) (*Alwinn, error) {
	a := &Alwinn{
		logs:           logging.NewChannel("bot-alwinn"),
		period:         NewSignal[period.Period](),
		pause:          NewSignal[bool](),
		inspirationDie: NewSignal[int64](),

		needsSave: true,
		cfg:       cfg,
		adb:       adb,
		gdb:       gdb,
	}
	a.ism = NewIsm(a.String(), &cfg.Bot, a.gdb, a.period, a.pause)
	a.ism.WithBucket(a.ism.PostIsm, 2, a.cfg.Subscribers, alwinnIsms...)
	a.ism.WithBucket(a.ism.PostGif, 1, a.cfg.Subscribers, alwinnGifs...)
	a.ism.WithBucket(a.PostInspirationDie, 1, a.cfg.Subscribers, alwinnInspirtationDieFmts...)
	a.WithDependencies(a.ism)
	return a, nil
}

func (a *Alwinn) Initialize(ctx context.Context) error {
	cfg, err := a.adb.Load(ctx, alwinnCfgFile)
	if err != nil {
		return err
	}

	if cfg != nil {
		cfg.Token = a.cfg.Token
		a.cfg = cfg
		a.needsSave = false
	}

	if a.cfg.Subscribers == nil {
		a.cfg.Subscribers = map[string]struct{}{}
	}

	a.period.Set(period.NewRandom(a.cfg.Period.Min, a.cfg.Period.Max))
	a.pause.Set(a.cfg.Pause)
	a.inspirationDie.Set(a.cfg.InspirationDie)

	_, err = a.ism.sess.ApplicationCommandCreate(a.ism.sess.State.User.ID, "", &discordgo.ApplicationCommand{
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
				Name:        "subscribe",
				Description: "Alwinn-ator should speak in the channel the command is called from.",
				Options: []*discordgo.ApplicationCommandOption{
					{
						Type:        discordgo.ApplicationCommandOptionBoolean,
						Description: "Whether to keep speaking in this channel.",
						Name:        "subscribe",
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

	a.ism.sess.WithInteractionCreateHandler("pause", a.handlePause)
	a.ism.sess.WithInteractionCreateHandler("subscribe", a.handleSubscribe)
	a.ism.sess.WithInteractionCreateHandler("set-ism-period", a.handleSetIsmPeriod)
	a.ism.sess.WithInteractionCreateHandler("set-inspiration-die", a.handleSetInspirationDie)

	return nil
}

func (a *Alwinn) Shutdown(ctx context.Context) error {
	a.period.Close()
	a.pause.Close()
	a.inspirationDie.Close()

	return nil
}

func (a *Alwinn) String() string {
	return "alwinn"
}

func (a *Alwinn) PostInspirationDie(ctx context.Context, msgFmt string, subscribers map[string]struct{}) error {
	for cid := range subscribers {
		if _, err := a.ism.sess.ChannelMessageSend(cid, fmt.Sprintf(msgFmt, a.inspirationDie.Get())); err != nil {
			a.logs.Print("could not post gif:", err)
		}
	}
	return nil
}

func (a *Alwinn) handlePause(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	p := findOption(subcmd.Options, "pause")
	pause := p.BoolValue()

	a.pause.Set(pause)
	a.cfg.Pause = pause
	resp.Write([]byte(fmt.Sprintf("pause set to: %v", pause)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.adb.Save(ctx, alwinnCfgFile, a.cfg)
}

func (a *Alwinn) handleSubscribe(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	s := findOption(subcmd.Options, "subscribe")
	sub := s.BoolValue()

	if sub {
		a.cfg.Subscribers[msg.ChannelID] = struct{}{}
		resp.Write([]byte("subscription added"))
	} else {
		delete(a.cfg.Subscribers, msg.ChannelID)
		resp.Write([]byte("subscription removed"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.adb.Save(ctx, alwinnCfgFile, a.cfg)
}

func (a *Alwinn) handleSetIsmPeriod(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
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

	a.period.Set(period.NewRandom(minDur, maxDur))
	a.cfg.Period.Min, a.cfg.Period.Max = minDur, maxDur
	resp.Write([]byte(fmt.Sprintf("speak periods set to: %v - %v", minDur, maxDur)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.adb.Save(ctx, alwinnCfgFile, a.cfg)
}

func (a *Alwinn) handleSetInspirationDie(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	inspirationDie := findOption(subcmd.Options, "inspiration-die").IntValue()
	if _, ok := dies[inspirationDie]; !ok {
		return fmt.Errorf("not a vaild die: %d", inspirationDie)
	}

	a.inspirationDie.Set(inspirationDie)
	a.cfg.InspirationDie = inspirationDie
	resp.Write([]byte(fmt.Sprintf("inspiration die set to: %v", inspirationDie)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return a.adb.Save(ctx, alwinnCfgFile, a.cfg)
}
