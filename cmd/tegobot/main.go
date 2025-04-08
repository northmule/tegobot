package main

import (
	"context"
	"log"
	"os/signal"
	"syscall"
	"tegobot/internal/config"
	"tegobot/internal/handler"
	"tegobot/internal/logger"
	"tegobot/internal/service"
	"tegobot/internal/storage"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	if err := run(ctx); err != nil {
		log.Fatal(err)
	}
}

func run(ctx context.Context) error {
	var err error

	cfg, err := config.NewConfig()
	if err != nil {
		return err
	}
	l, err := logger.NewLogger(cfg.Value().Log.FlePath, cfg.Value().Log.Level)
	if err != nil {
		return err
	}

	// Инициализация бота
	bot, err := tgbotapi.NewBotAPI(cfg.Value().Telegram.Token)
	if err != nil {
		log.Fatal(err)
	}

	bot.Debug = false
	l.Infof("Авторизован как %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	// Канал с данными
	updates := bot.GetUpdatesChan(u)

	// Зависимости
	akismetService := service.NewAkismet(cfg.Value().Spam.Akismet.ApiKey, cfg.Value().Spam.Akismet.SiteURL, l)
	storagePendingUsers := storage.NewPendingUser()
	newUserHandler := handler.NewNewUser(bot, cfg, storagePendingUsers, l)
	newUserHandler.AddVerifyService(service.NewLolsBot())

	messageHandler := handler.NewMessage(bot, akismetService, cfg, l)
	serviceMessageHandler := handler.NewServiceMessage(bot)

	for update := range updates {
		// Логируем все события в JSON формате
		l.WriteToJson("Входящие данные", update)

		// Обрабатываем добавление бота в группу
		if update.Message != nil && update.Message.NewChatMembers != nil {
			for _, member := range update.Message.NewChatMembers {
				if member.ID == bot.Self.ID {
					l.Infof(
						"[EVENT] Бот добавлен в группу: %s (ID: %d)",
						update.Message.Chat.Title,
						update.Message.Chat.ID,
					)
				}
			}
		}

		if update.Message != nil {
			if update.Message.From != nil {
				// Удаляем сообщения от неверифицированных пользователей
				if _, ok := storagePendingUsers.Get(update.Message.From.ID); ok {
					bot.Send(tgbotapi.NewDeleteMessage(
						update.Message.Chat.ID,
						update.Message.MessageID,
					))
					continue
				}
			}
			// Удаление всех сервисных сообщений
			serviceMessageHandler.Handle(update)

			// Обработка новых участников
			if update.Message.NewChatMembers != nil {
				for _, newUser := range update.Message.NewChatMembers {
					err = newUserHandler.HandleUser(update.Message.Chat.ID, &newUser)
					if err != nil {
						l.Error(err.Error())
					}
				}
			}

			// Проверка на спам через Akismet
			if update.Message.Text != "" {
				err = messageHandler.Handle(update.Message)
				if err != nil {
					l.Error(err.Error())
				}
			}

		}

		if update.CallbackQuery != nil {
			// Обработка нажатия кнопок
			err = newUserHandler.HandleUserAnswer(update.CallbackQuery)
			if err != nil {
				l.Error(err.Error())
			}
		}

	}

	return nil
}
