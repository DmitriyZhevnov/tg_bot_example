package bot

import (
	"context"
	"errors"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"log"
	"strconv"
	"testbot/interanl/businesslayer"
	"testbot/interanl/businesslayer/dto"
	"time"
)

type Processor struct {
	apiBot *tgbotapi.BotAPI
	logger zerolog.Logger

	usersProcessor    businesslayer.Users
	executorProcessor businesslayer.Executor

	usersChannels map[int64]chan tgbotapi.Update
}

func New(
	token string,
	logger zerolog.Logger,
	usersProcessor businesslayer.Users,
	executorProcessor businesslayer.Executor,
) (*Processor, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, err
	}

	return &Processor{
		apiBot:            bot,
		logger:            logger,
		usersProcessor:    usersProcessor,
		executorProcessor: executorProcessor,
		usersChannels:     make(map[int64]chan tgbotapi.Update),
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
			if update.CallbackQuery != nil {
				go p.handleCallbackQuery(update.CallbackQuery)
				continue
			}
			userChannel, isChannelFound := p.usersChannels[update.Message.Chat.ID]
			if isChannelFound {

				go func() {
					userChannel <- update
				}()
				continue
			}

			switch update.Message.Command() {
			case "start":
				go p.handleStart(ctx, update)
			case "work":
				if !p.isUserAuthorized(ctx, update.Message.Chat.ID, update.Message.Chat.UserName) {
					p.suggestToRunStartCommand(update.Message.Chat.ID, update.Message.Chat.UserName)

					continue
				}

				if !isChannelFound {
					userChannel = make(chan tgbotapi.Update)
					p.usersChannels[update.Message.Chat.ID] = userChannel
				}
				go p.handleWork(ctx, update, userChannel)
			case "lalal":
			default:
				// послать сообщение, что не понимаем че он хочет
				if err := p.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Дорогой, %s! Я не понимаю.", update.Message.Chat.UserName)); err != nil {
					p.logger.Err(err).Send()
				}
			}
		}
	}
}

func (p *Processor) handleStart(ctx context.Context, update tgbotapi.Update) {
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
		return
	}
	// приветствие
	if err := p.SendMessage(update.Message.Chat.ID, fmt.Sprintf("Привет, %s!", update.Message.Chat.UserName)); err != nil {
		p.logger.Err(err).Send()
		return
	}

	return
}

func (p *Processor) suggestToRunStartCommand(chatID int64, userName string) {
	if err := p.SendMessage(chatID, fmt.Sprintf("Дорогой, %s! Воспользуйся командой /start", userName)); err != nil {
		p.logger.Err(err).Send()
	}
}

func (p *Processor) isUserAuthorized(ctx context.Context, chatID int64, userName string) bool {
	// проверить, что в базе есть такой пользователь
	if _, err := p.usersProcessor.LoadByChatID(ctx, chatID); err != nil {
		p.logger.Err(err).Send()
		return false
	}

	return true
}

func (p *Processor) handleWork(ctx context.Context, update tgbotapi.Update, userChannel chan tgbotapi.Update) {
	defer func() {
		close(userChannel)
		delete(p.usersChannels, update.Message.Chat.ID)
	}()

	// выполняем работу
	if err := p.SendMessage(update.Message.Chat.ID, p.usersProcessor.Work(update.Message.Chat.UserName)); err != nil {
		p.logger.Err(err).Send()
		return
	}

	// задал вопрос.
	if err := p.sendQuestion(update.Message.Chat.ID, 1); err != nil {
		p.logger.Err(err).Send()

		return
	}

	responseTimer := time.NewTimer(1 * time.Minute)
	defer responseTimer.Stop()

	// получили ответ.
	select {
	case <-ctx.Done():
		p.logger.Debug().Msg("ctx is done")
	case <-responseTimer.C:
		p.logger.Debug().Msg("no response from user")
		return
	case response := <-userChannel:
		// в зависимости от ответа попросили ввести 2 значения.
		a, err := strconv.Atoi(response.Message.Text)
		if err != nil {
			msg := tgbotapi.NewMessage(response.Message.Chat.ID, "a должно быть числом")
			p.apiBot.Send(msg)

			p.logger.Debug().Msg("user entered incorrect a value")
			return
		}

		msg := tgbotapi.NewMessage(response.Message.Chat.ID, "Введи b:")
		p.apiBot.Send(msg)

		responseTimer.Reset(1 * time.Minute)

		select {
		case <-ctx.Done():
			p.logger.Debug().Msg("ctx is done")
		case <-responseTimer.C:
			p.logger.Debug().Msg("no response from user")
			return
		case response = <-userChannel:
			b, err := strconv.Atoi(response.Message.Text)
			if err != nil {
				msg := tgbotapi.NewMessage(response.Message.Chat.ID, "b должно быть числом")
				p.apiBot.Send(msg)

				p.logger.Debug().Msg("user entered incorrect b value")
				return
			}

			if err := p.executorProcessor.SaveValues(a, b); err != nil {
				/// some logic
				msg := tgbotapi.NewMessage(response.Message.Chat.ID, "Сохранено, спасибо")
				p.apiBot.Send(msg)
			}
		}
	}

	return
}

func (p *Processor) sendQuestion(chatID int64, questionID int) error {
	question, exists := questions[questionID]
	if !exists {
		return errors.New("question is not found")
	}

	var keyboard [][]tgbotapi.InlineKeyboardButton
	for _, option := range question.Options {
		callbackData := strconv.Itoa(option.NextQuestionID)
		row := []tgbotapi.InlineKeyboardButton{
			tgbotapi.NewInlineKeyboardButtonData(option.Text, callbackData),
		}
		keyboard = append(keyboard, row)
	}

	msg := tgbotapi.NewMessage(chatID, question.Text)
	msg.ReplyMarkup = tgbotapi.InlineKeyboardMarkup{
		InlineKeyboard: keyboard,
	}

	_, err := p.apiBot.Send(msg)
	if err != nil {
		p.logger.Err(err).Send()
		return err
	}

	return nil
}

func (p *Processor) handleCallbackQuery(callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	nextQuestionID, err := strconv.Atoi(callbackQuery.Data)
	if err != nil {
		log.Println("Ошибка преобразования callback data:", err)
		return
	}

	switch nextQuestionID {
	case finishOfFirstQuestion:
		msg := tgbotapi.NewMessage(chatID, "Введите а:")
		p.apiBot.Send(msg)
	case defaultFinish:
		msg := tgbotapi.NewMessage(chatID, "Спасибо за ответы! Диалог завершен.")
		p.apiBot.Send(msg)
	default:
		p.sendQuestion(chatID, nextQuestionID)
	}
}
