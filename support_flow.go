package main

import (
	"fmt"
	"regexp"
	"strings"
)

// support_flow.go — гид-бот поддержки с кнопками быстрого ответа. Хранит
// текущий шаг диалога в Store.supportStage (по userID) и решает, что
// ответить дальше. Если сообщение не относится ни к одной кнопке меню и
// пользователь не в спецстадии — возвращает нулевой supportFlowOutcome,
// и вызывающий код (api.go) откатывается на regex-классификатор (ai.Reply)
// для свободного текста.

const (
	stageAwaitingReview          = "awaiting_review"
	stageAwaitingIdea            = "awaiting_idea"
	stageAwaitingBookingFeedback = "awaiting_booking_feedback"
	stageAwaitingRegFeedback     = "awaiting_reg_feedback"
	stageAwaitingPaymentFeedback = "awaiting_payment_feedback"
	stageAwaitingOperator        = "awaiting_operator"
	// stageClosed — чат автоматически закрыт по неактивности (см.
	// support_sweep.go). Следующее сообщение пользователя снова показывает
	// приветствие с меню, как в самом начале, а не пытается его обработать
	// в контексте старого диалога.
	stageClosed = "closed"
)

// mainMenuOptions — метки кнопок главного меню. Ровно эти же строки
// использует static/support.js для синтетического приветствия при пустой
// истории чата — при изменении текста меняйте в обоих местах.
var mainMenuOptions = []string{
	"📝 Оставить отзыв",
	"💡 Предложение идеи",
	"📅 Как записаться на приём",
	"🔑 Помощь с регистрацией",
	"💳 Оплата и подписка",
	"🙋 Другой вопрос",
}

var yesNoOptions = []string{"Да, помогло", "Нет, не помогло"}
var ratingOptions = []string{"1", "2", "3", "4", "5"}

const menuHint = "Чем ещё могу помочь? Выберите вариант ниже или напишите свободно."

var ratingInMessageRe = regexp.MustCompile(`(?:^|\D)([1-5])(?:\D|$)`)

// firstNameOrFull — в User нет отдельного поля имени, берём первое слово ФИО.
func firstNameOrFull(user *User) string {
	parts := strings.Fields(user.FullName)
	if len(parts) > 0 {
		return parts[0]
	}
	return user.FullName
}

// mainMenuGreeting — стартовое сообщение бота с кнопками меню. Используется
// и для «Начать новый чат», и для автоматического приветствия после того,
// как предыдущий диалог закрылся по неактивности.
func mainMenuGreeting(user *User) supportBotResult {
	return supportBotResult{
		Body:    fmt.Sprintf("Здравствуйте, %s! Я поддержка Avorimi Health. Чем могу помочь?", firstNameOrFull(user)),
		Options: mainMenuOptions,
	}
}

func extractRating(text string) string {
	m := ratingInMessageRe.FindStringSubmatch(text)
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

const bookingHelpText = `Чтобы записаться на приём:
1. Откройте «Клиники» — можно разрешить геолокацию, чтобы список отсортировался по близости.
2. Выберите клинику → нужное направление (например, «Кардиолог» или «УЗИ брюшной полости»).
3. Сравните врачей: цена, рейтинг и свободное время у каждого разное.
4. Выберите удобный слот и нажмите «Записаться» — если ещё не входили, попросим войти или зарегистрироваться на этом шаге.
5. Оплатите приём (или спишите визит с подписки, если она у вас активна) — запись подтвердится сразу.`

const registrationHelpText = `Если не получается зарегистрироваться или войти:
1. ИИН должен состоять ровно из 12 цифр, без пробелов и букв.
2. Номер телефона вводится после +7, 10 цифр.
3. Пароль — не короче 6 символов, и оба поля (пароль и повтор) должны совпадать.
4. Нужно поставить обе галочки: согласие с Пользовательским соглашением/Политикой конфиденциальности и согласие на обработку персональных данных — без них аккаунт не создаётся.
5. Если пишет, что телефон уже занят — скорее всего, аккаунт уже был зарегистрирован раньше, попробуйте войти вместо регистрации.`

const paymentHelpText = `Про оплату и подписку:
1. Оплата приёма проходит сразу после записи на слот — деньги списываются один раз за визит.
2. Если оформлена подписка — визиты, входящие в неё, можно списывать без отдельной оплаты (это будет предложено при записи).
3. Подписки оформляются в разделе «Подписки»: разное количество визитов и срок действия.
4. Оплаченная запись сразу видна в личном кабинете со статусом.`

type supportBotResult struct {
	Body    string
	Options []string
	Stage   string // "" — вернуться в меню
}

type supportFlowOutcome struct {
	Reply supportBotResult
}

func (o supportFlowOutcome) handled() bool { return o.Reply.Body != "" }

func handleHelpFeedback(user *User, choice, topic string) supportFlowOutcome {
	if choice == yesNoOptions[0] {
		return supportFlowOutcome{Reply: supportBotResult{
			Body: "Рад был помочь! " + menuHint, Options: mainMenuOptions,
		}}
	}
	go notifyTelegramRelayable(user, fmt.Sprintf("⚠️ Проблема с %s\nОт: %s (%s)\n\nГотовая инструкция не помогла — пожалуйста, свяжитесь с пользователем.\n\n— Ответьте на это сообщение (Reply), ответ уйдёт пользователю в чат.", topic, user.FullName, user.Phone))
	return supportFlowOutcome{Reply: supportBotResult{
		Body: "Понял. Соединяю с оператором — он подключится в этом чате в ближайшее время.",
	}}
}

// runSupportFlow — основная стейт-машина. stage — текущий шаг пользователя
// (store.SupportStage), msg — только что присланное сообщение.
func runSupportFlow(user *User, stage, msg string) supportFlowOutcome {
	choice := strings.TrimSpace(msg)

	if stage == stageClosed {
		// Прошлый диалог закрылся по неактивности — начинаем как в первый раз.
		return supportFlowOutcome{Reply: mainMenuGreeting(user)}
	}

	switch stage {
	case stageAwaitingReview:
		rating := extractRating(choice)
		telegramBody := choice
		if rating != "" {
			telegramBody = fmt.Sprintf("Оценка: %s/5\n\n%s", rating, choice)
		}
		go notifyTelegramRelayable(user, fmt.Sprintf("💬 Новый отзыв на Avorimi Health\nОт: %s (%s)\n\n%s\n\n— Ответьте на это сообщение (Reply), если хотите поблагодарить пользователя лично.", user.FullName, user.Phone, telegramBody))
		return supportFlowOutcome{Reply: supportBotResult{
			Body: "Спасибо большое за отзыв! Передал его команде Avorimi. 🙌\n\n" + menuHint, Options: mainMenuOptions,
		}}

	case stageAwaitingIdea:
		go notifyTelegramIdea(user, choice)
		return supportFlowOutcome{Reply: supportBotResult{
			Body: "Отличная идея, спасибо! Передал команде на рассмотрение.\n\n" + menuHint, Options: mainMenuOptions,
		}}

	case stageAwaitingBookingFeedback:
		return handleHelpFeedback(user, choice, "записью на приём")

	case stageAwaitingRegFeedback:
		return handleHelpFeedback(user, choice, "регистрацией")

	case stageAwaitingPaymentFeedback:
		return handleHelpFeedback(user, choice, "оплатой/подпиской")

	case stageAwaitingOperator:
		go forwardSupportQuestionToTelegram(user, choice)
		return supportFlowOutcome{Reply: supportBotResult{
			Body: "Передал ваш вопрос оператору — он ответит здесь же, в этом чате.",
		}}
	}

	// stage == "" (меню) — сперва проверяем клик по кнопке главного меню.
	switch choice {
	case mainMenuOptions[0]: // Оставить отзыв
		return supportFlowOutcome{Reply: supportBotResult{
			Body:    fmt.Sprintf("Отлично, %s! Расскажите в двух словах, как вам платформа, и оцените от 1 до 5 ⭐", firstNameOrFull(user)),
			Options: ratingOptions, Stage: stageAwaitingReview,
		}}
	case mainMenuOptions[1]: // Предложение идеи
		return supportFlowOutcome{Reply: supportBotResult{
			Body: "Слушаю! Какая у вас идея по улучшению платформы?", Stage: stageAwaitingIdea,
		}}
	case mainMenuOptions[2]: // Как записаться на приём
		return supportFlowOutcome{Reply: supportBotResult{
			Body: bookingHelpText + "\n\nВам помогло?", Options: yesNoOptions, Stage: stageAwaitingBookingFeedback,
		}}
	case mainMenuOptions[3]: // Помощь с регистрацией
		return supportFlowOutcome{Reply: supportBotResult{
			Body: registrationHelpText + "\n\nВам помогло?", Options: yesNoOptions, Stage: stageAwaitingRegFeedback,
		}}
	case mainMenuOptions[4]: // Оплата и подписка
		return supportFlowOutcome{Reply: supportBotResult{
			Body: paymentHelpText + "\n\nВам помогло?", Options: yesNoOptions, Stage: stageAwaitingPaymentFeedback,
		}}
	case mainMenuOptions[5]: // Другой вопрос / оператор
		return supportFlowOutcome{Reply: supportBotResult{
			Body: "Опишите ваш вопрос — оператор ответит вам здесь же, в чате.", Stage: stageAwaitingOperator,
		}}
	}

	return supportFlowOutcome{} // не кнопка меню — пусть обработает обычный классификатор
}
