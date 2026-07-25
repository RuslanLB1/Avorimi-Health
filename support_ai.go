package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strings"
	"time"
)

// support_ai.go — провайдер AI-чата поддержки. По аналогии с payments.go:
// сейчас используется MockAIProvider (правила без реального ИИ), а когда
// появится ключ Anthropic API, включается AnthropicProvider — остальной код
// (support-хендлеры в api.go) менять не придётся.
//
// Что умеет чат поддержки:
//  1. Отвечает на вопросы «как сделать X» по работе платформы (записаться,
//     оплатить, найти результаты анализов).
//  2. Принимает отзывы и предложения по платформе — такие сообщения
//     дополнительно улетают в Telegram (см. telegram.go).
//  3. Эскалирует проблемы с регистрацией/входом администратору с возможностью
//     ответить прямо в чат пользователя (см. telegram_relay.go).

type AIMessage struct {
	Role SupportRole
	Body string
}

type AIProvider interface {
	Name() string
	Reply(history []AIMessage, userMessage string) (string, error)
}

const supportSystemPrompt = `Ты — ассистент поддержки казахстанской медицинской платформы Avorimi Health.
Платформа связывает пациентов с клиниками: пациент выбирает клинику, направление (врач или
обследование), удобное время и оплачивает приём (или списывает визит с подписки). Помогаешь
разобраться с платформой (как найти клинику, как записаться, как оплатить, как посмотреть
результаты анализов, как работает подписка), помогаешь с проблемами регистрации/входа и
принимаешь отзывы/предложения по сайту. Отвечай кратко, по-русски, дружелюбно и по делу.
Не придумывай функций, которых нет на платформе. Ты не даёшь медицинских консультаций —
по вопросам здоровья направляй к врачу в выбранной клинике.`

// --- Mock-провайдер: работает без внешних ключей уже сейчас ---

type MockAIProvider struct{}

func (MockAIProvider) Name() string { return "mock" }

var (
	feedbackRe = regexp.MustCompile(`(?i)(отзыв|предложени|фидбек|обратн\p{L}*\s*связ|не хватает|добавьте|плохо работает|баг|ошибк)`)
	findRe     = regexp.MustCompile(`(?i)(как\s+(найти|записаться|записат)|найти\s+клинику|где\s+клиника|выбрать\s+врача)`)
	paymentRe  = regexp.MustCompile(`(?i)(как\s+оплатить|оплата\s+не\s+прошла|списал\s+деньги|не\s+могу\s+оплатить)`)
	subRe      = regexp.MustCompile(`(?i)(подписк|абонемент|сколько визитов)`)
	resultsRe  = regexp.MustCompile(`(?i)(анализ|результат\p{L}*\s*(готов|анализ)|где.*результат)`)
	cancelRe   = regexp.MustCompile(`(?i)(отменить\s+запись|перенести\s+запись|не могу отменить|как\s+отменить)`)
	greetRe    = regexp.MustCompile(`(?i)^(привет|здравствуй|добрый день|добрый вечер|салам|hi|hello)\b`)

	// registrationTroubleRe ловит жалобы на регистрацию/вход —
	// такие сообщения дополнительно эскалируются в Telegram (см. api.go).
	registrationTroubleRe = regexp.MustCompile(`(?i)(не могу зарегистр|не получается зарегистр|не регистрирует|не работает регистрац|ошибка регистрац|не могу войти|не получается войти|иин не принима)`)

	// closingRe — короткий отрицательный/завершающий ответ ("нет", "спасибо", "ок").
	// Правила не понимают контекст, поэтому хотя бы не повторяем то же меню заново.
	closingRe = regexp.MustCompile(`(?i)^(нет|нету|ничего|неа|не\s*надо|спасибо|благодарю|ок|окей|хорошо|пока|всё|все)[.!\s]*$`)
)

func (MockAIProvider) Reply(history []AIMessage, userMessage string) (string, error) {
	msg := strings.TrimSpace(userMessage)

	if closingRe.MatchString(msg) && len(history) > 0 && history[len(history)-1].Role == SupportRoleAssistant {
		return "Хорошо! Если появятся вопросы — я тут, просто напишите ещё раз.", nil
	}

	switch {
	case registrationTroubleRe.MatchString(msg):
		return "Понимаю, это неприятно. Проверьте: 1) ИИН — ровно 12 цифр, 2) телефон введён после +7, 3) обе галочки согласия отмечены, 4) пароль не короче 6 символов и совпадает в обоих полях. Если не поможет — напишите здесь, что именно происходит (какой шаг, какая ошибка), я передам это команде Avorimi, разберёмся вручную.", nil

	case feedbackRe.MatchString(msg):
		return "Спасибо! Передал ваше сообщение команде Avorimi — уже улетело в наш внутренний канал. Если хотите, опишите подробнее, что стоит добавить или исправить.", nil

	case findRe.MatchString(msg):
		return "Откройте «Клиники» — можно разрешить геолокацию, чтобы список отсортировался по близости. Дальше выберите клинику → направление (например, «Кардиолог») → врача или процедуру → удобное время.", nil

	case cancelRe.MatchString(msg):
		return "Отменить или перенести запись пока можно только через оператора — опишите здесь номер записи и что нужно изменить, я передам команде Avorimi.", nil

	case resultsRe.MatchString(msg):
		return "Результаты анализов появляются в разделе «Мои анализы» — обычно готовы в течение некоторого времени после визита, статус меняется с «Ожидается» на «Готово».", nil

	case subRe.MatchString(msg):
		return "Подписка даёт определённое количество визитов на выбранный срок — оформить можно в разделе «Подписки». При записи на приём можно списать визит с активной подписки вместо разовой оплаты.", nil

	case paymentRe.MatchString(msg):
		return "Оплата проходит сразу после выбора времени приёма — по карте или со списанием визита с подписки, если она у вас активна. Если оплата не проходит — попробуйте ещё раз чуть позже или напишите здесь, я передам команде.", nil

	case greetRe.MatchString(msg):
		return "Здравствуйте! Я помогу разобраться с платформой — найти клинику, записаться на приём, разобраться с оплатой или подпиской, либо передам отзыв команде Avorimi. Что вас интересует?", nil

	default:
		return "Могу подсказать, как найти клинику и записаться на приём, как работает оплата и подписка, где смотреть результаты анализов, или передать отзыв о платформе команде Avorimi. Уточните, пожалуйста, что вас интересует?", nil
	}
}

// --- Anthropic-провайдер: включается, когда появится ANTHROPIC_API_KEY ---

type AnthropicProvider struct {
	APIKey string
	Model  string
	client *http.Client
}

func NewAnthropicProvider() *AnthropicProvider {
	model := os.Getenv("ANTHROPIC_MODEL")
	if model == "" {
		model = "claude-haiku-4-5-20251001" // быстрая и дешёвая модель для чата поддержки
	}
	return &AnthropicProvider{
		APIKey: os.Getenv("ANTHROPIC_API_KEY"),
		Model:  model,
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (p *AnthropicProvider) Name() string { return "anthropic:" + p.Model }

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error"`
}

func (p *AnthropicProvider) Reply(history []AIMessage, userMessage string) (string, error) {
	if p.APIKey == "" {
		return "", fmt.Errorf("Anthropic API не подключён: задайте ANTHROPIC_API_KEY")
	}

	messages := make([]anthropicMessage, 0, len(history)+1)
	for _, m := range history {
		role := "user"
		if m.Role == SupportRoleAssistant {
			role = "assistant"
		}
		messages = append(messages, anthropicMessage{Role: role, Content: m.Body})
	}
	messages = append(messages, anthropicMessage{Role: "user", Content: userMessage})

	reqBody, err := json.Marshal(anthropicRequest{
		Model:     p.Model,
		MaxTokens: 1024,
		System:    supportSystemPrompt,
		Messages:  messages,
	})
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", "https://api.anthropic.com/v1/messages", bytes.NewReader(reqBody))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.APIKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	res, err := p.client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	var parsed anthropicResponse
	if err := json.NewDecoder(res.Body).Decode(&parsed); err != nil {
		return "", err
	}
	if parsed.Error != nil {
		return "", fmt.Errorf("anthropic: %s", parsed.Error.Message)
	}
	if len(parsed.Content) == 0 {
		return "", fmt.Errorf("anthropic: пустой ответ")
	}
	return parsed.Content[0].Text, nil
}
