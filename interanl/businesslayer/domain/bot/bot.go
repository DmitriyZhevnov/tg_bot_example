package bot

import (
	"context"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"testbot/interanl/businesslayer"
	"testbot/interanl/businesslayer/dto"
)

type Processor struct {
	apiBot *tgbotapi.BotAPI
	logger zerolog.Logger

	usersProcessor businesslayer.Users
}

func New(token string, logger zerolog.Logger, usersProcessor businesslayer.Users) (*Processor, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Processor{
		apiBot:         bot,
		logger:         logger,
		usersProcessor: usersProcessor,
	}, nil
}

func (p *Processor) SendMessage(chatID int64, message string) error {
	msg := tgbotapi.NewMessage(chatID, message)

	if _, err := p.apiBot.Send(msg); err != nil {
		return err
	}

	return nil
}

func (p *Processor) Listen(ctx context.Context) error {
	u := tgbotapi.NewUpdate(0)
	//u.Timeout = math.MaxInt

	updates := p.apiBot.GetUpdatesChan(u)

	for {
		select {
		case <-ctx.Done():
			p.logger.Debug().Msg("context is dode")

			return nil
		// закрытие контекста
		case update := <-updates:
			switch update.Message.Command() {
			case "start":
				// сохранить в базе
				if err := p.usersProcessor.CreateIfNotExist(
					ctx,
					dto.User{
						ID:     uuid.New(),
						Name:   update.Message.Chat.UserName,
						ChatID: update.Message.Chat.ID,
					},
				); err != nil {
					p.logger.Err(err).Send()

					continue
				}
				// приветствие
				if err := p.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Привет, %s!", update.Message.Chat.UserName)); err != nil {
					p.logger.Err(err).Send()
				}
			case "work":
				// проверить, что в базе есть такой пользователь
				_, err := p.usersProcessor.LoadByChatID(ctx, update.Message.Chat.ID)
				if err != nil {
					p.logger.Err(err).Send()
					// 	если нет, то порекомендовать запустить start
					if err := p.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Дорогой, %s! Воспользуйся командой /start", update.Message.Chat.UserName)); err != nil {
						p.logger.Err(err).Send()
					}

					continue
				}
				// выполняем работу
				if err := p.SendMessage(update.Message.Chat.ID, p.usersProcessor.Work(update.Message.Chat.UserName)); err != nil {
					p.logger.Err(err).Send()
				}
			default:
				// послать сообщение, что не понимаем че он хочет
				if err := p.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Дорогой, %s! Я не понимаю.", update.Message.Chat.UserName)); err != nil {
					p.logger.Err(err).Send()
				}
			}

			//if update.Message != nil { // If we got a message
			//	p.logger.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)
			//
			//	msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
			//	msg.ReplyToMessageID = update.Message.MessageID
			//
			//	_, err := p.apiBot.Send(msg)
			//	if err != nil {
			//		p.logger.Err(err).Send()
			//
			//		return err
			//	}
			//}
		}
	}
}
