package tg_bot_example

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"strconv"
)

type Question struct {
	Text    string
	Options []Option
}

type Option struct {
	Text           string
	NextQuestionID int
}

var questions = map[int]Question{
	1: {
		Text: "Первый вопрос: Выберите вариант",
		Options: []Option{
			{Text: "Вариант 1", NextQuestionID: 2},
			{Text: "Вариант 2", NextQuestionID: 3},
		},
	},
	2: {
		Text: "Второй вопрос для Варианта 1: Что дальше?",
		Options: []Option{
			{Text: "Вариант 1-1", NextQuestionID: 4},
			{Text: "Вариант 1-2", NextQuestionID: 5},
		},
	},
	3: {
		Text: "Второй вопрос для Варианта 2: Что дальше?",
		Options: []Option{
			{Text: "Вариант 2-1", NextQuestionID: 4},
			{Text: "Вариант 2-2", NextQuestionID: 5},
		},
	},
	4: {
		Text: "Финальный вопрос 4: Последний выбор",
		Options: []Option{
			{Text: "Завершить 4-1", NextQuestionID: -1},
			{Text: "Завершить 4-2", NextQuestionID: -1},
		},
	},
	5: {
		Text: "Финальный вопрос 5: Последний выбор",
		Options: []Option{
			{Text: "Завершить 5-1", NextQuestionID: -1},
			{Text: "Завершить 5-2", NextQuestionID: -1},
		},
	},
}

var userLastMessageID = make(map[int64]int)

func main() {
	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates, err := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil && update.Message.Text == "/start" {
			handleStartCommand(bot, update.Message)
		} else if update.CallbackQuery != nil {
			handleCallbackQuery(bot, update.CallbackQuery)
		}
	}
}

func handleStartCommand(bot *tgbotapi.BotAPI, msg *tgbotapi.Message) {
	sendQuestion(bot, msg.Chat.ID, 1)
}

func sendQuestion(bot *tgbotapi.BotAPI, chatID int64, questionID int) {
	question, exists := questions[questionID]
	if !exists {
		return
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

	sentMsg, err := bot.Send(msg)
	if err != nil {
		log.Println("Ошибка отправки сообщения:", err)
		return
	}

	if lastMsgID, ok := userLastMessageID[chatID]; ok {
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, lastMsgID)
		bot.Send(deleteMsg)
	}
	userLastMessageID[chatID] = sentMsg.MessageID
}

func handleCallbackQuery(bot *tgbotapi.BotAPI, callbackQuery *tgbotapi.CallbackQuery) {
	chatID := callbackQuery.Message.Chat.ID
	nextQuestionID, err := strconv.Atoi(callbackQuery.Data)
	if err != nil {
		log.Println("Ошибка преобразования callback data:", err)
		return
	}

	if lastMsgID, ok := userLastMessageID[chatID]; ok {
		deleteMsg := tgbotapi.NewDeleteMessage(chatID, lastMsgID)
		bot.Send(deleteMsg)
		delete(userLastMessageID, chatID)
	}

	if nextQuestionID != -1 {
		sendQuestion(bot, chatID, nextQuestionID)
	} else {
		msg := tgbotapi.NewMessage(chatID, "Спасибо за ответы! Диалог завершен.")
		bot.Send(msg)
	}

	bot.AnswerCallbackQuery(tgbotapi.NewCallback(callbackQuery.ID, ""))
}
