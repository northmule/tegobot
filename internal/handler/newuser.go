package handler

import (
	"fmt"
	"strconv"
	"tegobot/internal/config"
	"tegobot/internal/logger"
	"tegobot/internal/storage"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type VerificationByUserId interface {
	Verify(userID int64) (bool, error)
	GetServiceName() string
}

// NewUser Новый пользователь в телеграм
type NewUser struct {
	bot            *tgbotapi.BotAPI
	logger         *logger.Logger
	config         *config.Config
	pendingUsers   *storage.PendingUser
	verifyServices []VerificationByUserId
}

func NewNewUser(bot *tgbotapi.BotAPI, config *config.Config, pendingUsers *storage.PendingUser, logger *logger.Logger) *NewUser {
	instance := new(NewUser)
	instance.bot = bot
	instance.config = config
	instance.logger = logger
	instance.pendingUsers = pendingUsers
	return instance
}

// HandleUser обработка события "вступления в группу"
func (u *NewUser) HandleUser(chatID int64, user *tgbotapi.User) error {
	// Ограничиваем права пользователя
	err := u.restrictionRights(chatID, user.ID)
	u.logger.Info("Ограничение прав для пользователя", user.ID)
	if err != nil {
		return err
	}

	// Проверка по списку сервисов
	for _, v := range u.verifyServices {
		u.logger.Info(fmt.Sprintf("Проверка через сервис: %s, пользователя  %d", v.GetServiceName(), user.ID))
		var errorVerify error
		ok, errorVerify := v.Verify(user.ID)
		if errorVerify != nil {
			u.logger.Errorf(errorVerify.Error())
			continue
		}
		u.logger.Info(fmt.Sprintf("Результат проверки: %v", ok))
		if !ok {
			return nil
		}
	}

	// Создаем сообщение с кнопками
	msg := tgbotapi.NewMessage(chatID, fmt.Sprintf("%s, вы человек? (Ответьте в течение %s секунд)", user.FirstName, strconv.FormatInt(u.config.Value().Spam.Common.ResponseWaitingTime, 10)))
	msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData("Да", "yes"),
			tgbotapi.NewInlineKeyboardButtonData("Нет", "no"),
		),
	)

	sentMsg, _ := u.bot.Send(msg)
	u.logger.Info("Пользователю задан вопрос", user.ID)
	duration := time.Duration(u.config.Value().Spam.Common.ResponseWaitingTime) * time.Second
	// Таймер для блокировки
	timer := time.AfterFunc(duration, func() {
		// Блокируем пользователя если время вышло
		u.bot.Send(tgbotapi.NewDeleteMessage(chatID, sentMsg.MessageID))
		u.banUser(chatID, user.ID)
		u.pendingUsers.Delete(user.ID)
	})
	u.pendingUsers.Set(user.ID, &storage.UserTimer{
		Timer:  timer,
		ChatID: chatID,
		UserID: user.ID,
	})
	u.logger.Info("Установлен таймер для пользователя", user.ID)
	return nil
}

// HandleUserAnswer обработка ответа пользователя
func (u *NewUser) HandleUserAnswer(query *tgbotapi.CallbackQuery) error {
	var err error

	// Останавливаем таймер если он существует
	if userTimer, ok := u.pendingUsers.Get(query.From.ID); ok {
		userTimer.Timer.Stop()
		u.pendingUsers.Delete(query.From.ID)
		u.logger.Info("Пользователь дал ответ", query.Data)
		// Обрабатываем ответ
		if query.Data == "yes" {
			u.logger.Info("Пользователю выданы разрешения", query.From.ID)
			// Восстанавливаем права
			err = u.addingRights(userTimer.ChatID, userTimer.UserID)
		}
	} else {
		u.logger.Info("Пользователь не найден в списке ожидающих решения", query.From.ID)
	}

	// Удаляем сообщение с кнопками
	_, err = u.bot.Send(tgbotapi.NewDeleteMessage(query.Message.Chat.ID, query.Message.MessageID))

	return err
}

func (u *NewUser) AddVerifyService(service VerificationByUserId) {
	u.verifyServices = append(u.verifyServices, service)
}

// Ограничиваем права пользователя
func (u *NewUser) restrictionRights(chatID int64, userID int64) error {
	_, err := u.bot.Request(tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		Permissions: &tgbotapi.ChatPermissions{
			CanSendMessages:       false,
			CanSendMediaMessages:  false,
			CanSendPolls:          false,
			CanSendOtherMessages:  false,
			CanAddWebPagePreviews: false,
			CanChangeInfo:         false,
			CanInviteUsers:        false,
			CanPinMessages:        false,
		},
	})

	return err
}

// Восстанавливаем права
func (u *NewUser) addingRights(chatID int64, userID int64) error {

	_, err := u.bot.Request(tgbotapi.RestrictChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
		Permissions: &tgbotapi.ChatPermissions{
			CanSendMessages:       true,
			CanSendMediaMessages:  true,
			CanSendPolls:          true,
			CanSendOtherMessages:  true,
			CanAddWebPagePreviews: true,
			CanChangeInfo:         false,
			CanInviteUsers:        false,
			CanPinMessages:        false,
		},
	})

	return err
}

// Блокирование пользователя и добавление в черный список
func (u *NewUser) banUser(chatID int64, userID int64) error {

	_, err := u.bot.Send(tgbotapi.BanChatMemberConfig{
		ChatMemberConfig: tgbotapi.ChatMemberConfig{
			ChatID: chatID,
			UserID: userID,
		},
	})

	return err
}
