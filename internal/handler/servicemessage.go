package handler

import tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

// ServiceMessage сервисные сообщения
type ServiceMessage struct {
	bot *tgbotapi.BotAPI
}

func NewServiceMessage(bot *tgbotapi.BotAPI) *ServiceMessage {
	instance := new(ServiceMessage)
	instance.bot = bot
	return instance
}

func (s *ServiceMessage) Handle(update tgbotapi.Update) {
	if s.isServiceMessage(update.Message) {
		s.bot.Send(tgbotapi.NewDeleteMessage(
			update.Message.Chat.ID,
			update.Message.MessageID,
		))
	}
}

// Проверка на сервисное сообщение
func (s *ServiceMessage) isServiceMessage(msg *tgbotapi.Message) bool {
	return msg.NewChatMembers != nil ||
		msg.LeftChatMember != nil ||
		msg.NewChatTitle != "" ||
		msg.NewChatPhoto != nil ||
		msg.DeleteChatPhoto ||
		msg.GroupChatCreated ||
		msg.SuperGroupChatCreated ||
		msg.ChannelChatCreated ||
		msg.MigrateToChatID != 0 ||
		msg.MigrateFromChatID != 0 ||
		msg.PinnedMessage != nil ||
		msg.ProximityAlertTriggered != nil
}
