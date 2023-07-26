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
	turgCfgFile = "turg-config"
)

var (
	turgTempHPFmts = []string{
		"don't forget your temp hp! (%v)",
	}

	turgGifs = []string{
		"bless",
		"blessed",
		"you are blessed",
	}

	turgIsms = []string{
		"did you remember you’re blessed?",
		"*you're blessed!*",
		"bless your attack rolls ya goofs",
		"4 level bless, starry form dragon",
	}
)

type Turg struct {
	service.UnimplementedService

	ism    *Ism
	logs   *logging.Channel
	period *Signal[period.Period]
	pause  *Signal[bool]
	tmpHP  *Signal[int64]

	needsSave bool
	cfg       *config.Turg
	tdb       dal.Turg
	gdb       dal.Gif
}

func NewTurg(tdb dal.Turg, gdb dal.Gif, cfg *config.Turg) (*Turg, error) {
	t := &Turg{
		logs:   logging.NewChannel("turg"),
		period: NewSignal[period.Period](),
		pause:  NewSignal[bool](),
		tmpHP:  NewSignal[int64](),

		needsSave: true,
		cfg:       cfg,
		tdb:       tdb,
		gdb:       gdb,
	}
	t.ism = NewIsm(t.String(), &cfg.Bot, t.gdb, t.period, t.pause)
	t.ism.WithBucket(t.ism.PostIsm, 2, t.cfg.Subscribers, turgIsms...)
	t.ism.WithBucket(t.ism.PostGif, 1, t.cfg.Subscribers, turgGifs...)
	t.ism.WithBucket(t.PostTempHP, 1, t.cfg.Subscribers, turgTempHPFmts...)
	t.WithDependencies(t.ism)
	return t, nil
}

func (t *Turg) Initialize(ctx context.Context) error {
	cfg, err := t.tdb.Load(ctx, turgCfgFile)
	if err != nil {
		return err
	}

	if cfg != nil {
		cfg.Token = t.cfg.Token
		t.cfg = cfg
		t.needsSave = false
	}

	if t.cfg.Subscribers == nil {
		t.cfg.Subscribers = map[string]struct{}{}
	}

	t.period.Set(period.NewRandom(t.cfg.Period.Min, t.cfg.Period.Max))
	t.pause.Set(t.cfg.Pause)
	t.tmpHP.Set(t.cfg.TempHP)

	_, err = t.ism.sess.ApplicationCommandCreate(t.ism.sess.State.User.ID, "", &discordgo.ApplicationCommand{
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
				Name:        "subscribe",
				Description: "Turg-o-tron should speak in the channel the command is called from.",
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

	t.ism.sess.WithInteractionCreateHandler("pause", t.handlePause)
	t.ism.sess.WithInteractionCreateHandler("subscribe", t.handleSubscribe)
	t.ism.sess.WithInteractionCreateHandler("set-ism-period", t.handleSetIsmPeriod)
	t.ism.sess.WithInteractionCreateHandler("set-tmp-hp", t.handleSetTmpHP)

	return nil
}

func (t *Turg) Shutdown(ctx context.Context) error {
	t.period.Close()
	t.pause.Close()
	t.tmpHP.Close()

	return nil
}

func (t *Turg) String() string {
	return "turg"
}

func (t *Turg) PostTempHP(ctx context.Context, msgFmt string, subscribers map[string]struct{}) error {
	for cid := range subscribers {
		if _, err := t.ism.sess.ChannelMessageSend(cid, fmt.Sprintf(msgFmt, t.tmpHP.Get())); err != nil {
			t.logs.Print("could not post gif:", err)
		}
	}
	return nil
}

func (t *Turg) handlePause(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	p := findOption(subcmd.Options, "pause")
	pause := p.BoolValue()

	t.pause.Set(pause)
	t.cfg.Pause = pause
	resp.Write([]byte(fmt.Sprintf("pause set to: %v", pause)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.tdb.Save(ctx, turgCfgFile, t.cfg)
}

func (t *Turg) handleSubscribe(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	s := findOption(subcmd.Options, "subscribe")
	sub := s.BoolValue()

	if sub {
		t.cfg.Subscribers[msg.ChannelID] = struct{}{}
		resp.Write([]byte("subscription added"))
	} else {
		delete(t.cfg.Subscribers, msg.ChannelID)
		resp.Write([]byte("subscription removed"))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.tdb.Save(ctx, turgCfgFile, t.cfg)
}

func (t *Turg) handleSetIsmPeriod(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
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

	t.period.Set(period.NewRandom(minDur, maxDur))
	t.cfg.Period.Min, t.cfg.Period.Max = minDur, maxDur
	resp.Write([]byte(fmt.Sprintf("speak periods set to: %v - %v", minDur, maxDur)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.tdb.Save(ctx, turgCfgFile, t.cfg)
}

func (t *Turg) handleSetTmpHP(subcmd *discordgo.ApplicationCommandInteractionDataOption, msg *discordgo.InteractionCreate, resp io.Writer) error {
	tmpHP := findOption(subcmd.Options, "tmp-hp").IntValue()

	t.tmpHP.Set(tmpHP)
	t.cfg.TempHP = tmpHP
	resp.Write([]byte(fmt.Sprintf("tmp hp set to: %v", tmpHP)))

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return t.tdb.Save(ctx, turgCfgFile, t.cfg)
}
