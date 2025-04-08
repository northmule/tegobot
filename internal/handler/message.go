package handler

import (
	"tegobot/internal/config"
	"tegobot/internal/logger"
	"tegobot/internal/service"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Message struct {
	bot     *tgbotapi.BotAPI
	akismet *service.Akismet
	config  *config.Config
	logger  *logger.Logger
}

func NewMessage(bot *tgbotapi.BotAPI, akismet *service.Akismet, config *config.Config, logger *logger.Logger) *Message {
	instance := new(Message)
	instance.bot = bot
	instance.akismet = akismet
	instance.config = config
	instance.logger = logger
	return instance
}

func (m *Message) Handle(message *tgbotapi.Message) error {
	var err error
	commentData := service.CommentData{
		Blog:           m.config.Value().Spam.Akismet.SiteURL,
		UserIp:         service.GeneratePublicIPv4(),
		UserAgent:      "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:83.0) Gecko/20100101 Firefox/83.0",
		CommentType:    "comment",
		CommentAuthor:  message.From.FirstName + " " + message.From.LastName,
		CommentContent: message.Text,
	}

	spam, err := m.akismet.MessageIsSpam(commentData)

	if err != nil {
		return err
	}

	if !spam {
		return nil
	}

	_, _ = m.bot.Send(tgbotapi.DeleteMessageConfig{
		ChatID:    message.Chat.ID,
		MessageID: message.MessageID,
	})
	m.logger.Infof("[ANTISPAM] Удалено спам-сообщение от %s (ID: %d)",
		message.From.UserName,
		message.From.ID,
	)

	return nil
}
