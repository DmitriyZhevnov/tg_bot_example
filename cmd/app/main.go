package main

import (
	"context"
	"github.com/rs/zerolog"
	"os"
	"os/signal"
	"syscall"
	"testbot/interanl/businesslayer/domain/bot"
	"testbot/interanl/businesslayer/domain/executor"
	"testbot/interanl/businesslayer/domain/users"
	"testbot/interanl/datalayer/collections/cache"
)

// tgBot - @BotFather
// @GetMyChatID_BestBot

// THIS VALUES SHOULD BE IN CONFIG/ENV FILE
const (
	apiKey      = "8008625848:AAFQ-xFfdNo7KS3cBM0JBqbNvS-bDEmDkzI"
	adminChatID = 401631302
)

func main() {
	ctx := context.Background()
	logger := zerolog.New(os.Stderr).With().Timestamp().Logger()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	usersCollection := cache.NewUsersCollection()
	usersProcessor := users.NewProcessor(logger, usersCollection)
	executorProcessor := executor.NewProcessor()
	tgBot, err := bot.New(apiKey, logger, usersProcessor, executorProcessor)
	if err != nil {
		logger.Err(err).Send()

		return
	}

	// Канал для приема сигналов от операционной системы
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		// Ждем сигнал остановки
		<-stop
		if err := tgBot.SendMessage(adminChatID, "Бот остановлен!"); err != nil {
			logger.Err(err).Send()

			return
		}
	}()

	if err := tgBot.SendMessage(adminChatID, "bot started"); err != nil {
		logger.Err(err).Send()

		return
	}

	defer func() {
		if err := tgBot.SendMessage(adminChatID, "bot finished"); err != nil {
			logger.Err(err).Send()

			return
		}
	}()

	if err := tgBot.Listen(ctx); err != nil {
		logger.Err(err).Send()

		return
	}
}
