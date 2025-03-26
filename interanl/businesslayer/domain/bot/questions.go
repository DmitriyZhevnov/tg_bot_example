package bot

const (
	// TODO: rename
	defaultFinish         = -1
	finishOfFirstQuestion = -100
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
			{Text: "Вариант 1, спрошу 2 числа", NextQuestionID: finishOfFirstQuestion},
			{Text: "Вариант 2", NextQuestionID: 3},
		},
	},
	2: {
		Text: "Введите число a:",
		//Options: []Option{
		//	{Text: "Вариант 1-1", NextQuestionID: 4},
		//	{Text: "Вариант 1-2", NextQuestionID: 5},
		//},
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
			{Text: "Завершить 4-1", NextQuestionID: 5},
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
