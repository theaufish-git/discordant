package bots

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/bwmarrin/discordgo"
	"github.com/rleszilm/genms/logging"
	"github.com/rleszilm/genms/service"
	"github.com/theaufish-git/discordant/cmd/discordant/config"
)

type InteractionCreateHandler func(*discordgo.ApplicationCommandInteractionDataOption, *discordgo.InteractionCreate, io.Writer) error

type Session struct {
	service.UnimplementedService
	*discordgo.Session

	logs *logging.Channel
	bot  *config.Bot

	interactionCreateHandlers map[string]InteractionCreateHandler
}

func NewSession(bot *config.Bot) *Session {
	return &Session{
		logs:                      logging.NewChannel("session"),
		bot:                       bot,
		interactionCreateHandlers: map[string]InteractionCreateHandler{},
	}
}

func (s *Session) Initialize(ctx context.Context) error {
	s.logs.Print("starting session...")
	defer s.logs.Print("session started...")

	sess, err := discordgo.New("Bot " + s.bot.Token)
	if err != nil {
		return err
	}
	s.Session = sess

	err = s.Session.Open()
	if err != nil {
		return err
	}

	s.AddHandler(s.handleInteractionCreate)
	return nil
}

func (s *Session) Shutdown(ctx context.Context) error {
	s.logs.Print("stopping session...")
	defer s.logs.Print("session stopped...")
	return s.Close()
}

func (s *Session) String() string {
	return "session"
}

func (s *Session) WithInteractionCreateHandler(cmd string, handler InteractionCreateHandler) {
	s.interactionCreateHandlers[cmd] = handler
}

func (s *Session) handleInteractionCreate(sess *discordgo.Session, msg *discordgo.InteractionCreate) {
	switch msg.Type {
	case discordgo.InteractionApplicationCommand:
		data := msg.ApplicationCommandData()
		for _, subcmd := range data.Options {
			resp := &bytes.Buffer{}
			handler, ok := s.interactionCreateHandlers[subcmd.Name]
			if !ok {
				sess.InteractionRespond(msg.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: "not implemented",
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				continue
			}

			if err := handler(subcmd, msg, resp); err != nil {
				sess.InteractionRespond(msg.Interaction, &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{
						Content: fmt.Sprintf("error: %+v", err),
						Flags:   discordgo.MessageFlagsEphemeral,
					},
				})
				continue
			}

			sess.InteractionRespond(msg.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Content: resp.String(),
					Flags:   discordgo.MessageFlagsEphemeral,
				},
			})
		}
	}
}
