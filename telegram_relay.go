package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"
)

// telegram_relay.go — человек-в-контуре поддержка через Telegram, пока не
// подключён платный Anthropic API. Каждый вопрос пользователя дублируется
// администратору в Telegram; если он отвечает через "Reply" на это
// сообщение, ответ автоматически попадает в чат пользователя как реплика
// ассистента. Работает поверх обычного long polling (getUpdates) — не
// требует публичного HTTPS-домена и вебхука.

type tgMessage struct {
	MessageID      int        `json:"message_id"`
	Text           string     `json:"text"`
	Chat           tgChat     `json:"chat"`
	ReplyToMessage *tgMessage `json:"reply_to_message"`
}

type tgChat struct {
	ID int64 `json:"id"`
}

type tgUpdate struct {
	UpdateID int       `json:"update_id"`
	Message  tgMessage `json:"message"`
}

type tgUpdatesResponse struct {
	OK     bool       `json:"ok"`
	Result []tgUpdate `json:"result"`
}

type tgSendResponse struct {
	OK     bool      `json:"ok"`
	Result tgMessage `json:"result"`
}

// forwardSupportQuestionToTelegram пересылает вопрос пользователя админу и
// запоминает ID отправленного в Telegram сообщения, чтобы связать будущий
// Reply-ответ с нужным пользователем поддержки.
func forwardSupportQuestionToTelegram(user *User, question string) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		return
	}
	text := fmt.Sprintf("❓ Вопрос в поддержку от %s (%s)\n\n%s\n\n— Ответьте на это сообщение (Reply), чтобы ответ ушёл пользователю в чат.", user.FullName, user.Phone, question)

	body, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text})
	resp, err := telegramClient.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
		"application/json", bytes.NewReader(body),
	)
	if err != nil {
		log.Printf("[telegram-relay] не удалось переслать вопрос: %v", err)
		return
	}
	defer resp.Body.Close()

	var parsed tgSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil || !parsed.OK {
		log.Printf("[telegram-relay] некорректный ответ Telegram при пересылке вопроса")
		return
	}
	store.RegisterRelay(parsed.Result.MessageID, user.ID)
}

// forwardGuestQuestionToTelegram — аналог forwardSupportQuestionToTelegram
// для обращений без аккаунта (например, застрял на регистрации и не может
// войти): Reply на это сообщение уходит гостю в чат на сайте (пока открыта
// вкладка/сессия), а не на телефон/почту.
func forwardGuestQuestionToTelegram(guestID, name, contact, question string) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		return
	}
	text := fmt.Sprintf("🆘 Вопрос без аккаунта от %s (%s)\n\n%s\n\n— Ответьте на это сообщение (Reply), ответ уйдёт прямо в чат на сайте.", name, contact, question)

	body, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text})
	resp, err := telegramClient.Post(
		fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token),
		"application/json", bytes.NewReader(body),
	)
	if err != nil {
		log.Printf("[telegram-relay] не удалось переслать гостевой вопрос: %v", err)
		return
	}
	defer resp.Body.Close()

	var parsed tgSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil || !parsed.OK {
		log.Printf("[telegram-relay] некорректный ответ Telegram при пересылке гостевого вопроса")
		return
	}
	store.RegisterGuestRelay(parsed.Result.MessageID, guestID)
}

// notifyTelegramRelayable — как notifyTelegram, но дополнительно регистрирует
// релей на отправленное сообщение: используется там, где после уведомления
// ожидается, что администратор сможет ответить пользователю Reply'ем
// (эскалации гид-бота, проблемы с регистрацией и т.п.). Обычный notifyTelegram
// для этого не годится — Reply на такое сообщение никуда не приведёт.
func notifyTelegramRelayable(user *User, text string) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		log.Printf("[telegram] не настроен — событие только в логах: %s", text)
		return
	}
	body, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text})
	resp, err := telegramClient.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token), "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[telegram-relay] не удалось отправить: %v", err)
		return
	}
	defer resp.Body.Close()
	var parsed tgSendResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil || !parsed.OK {
		log.Printf("[telegram-relay] некорректный ответ Telegram")
		return
	}
	store.RegisterRelay(parsed.Result.MessageID, user.ID)
}

// startTelegramRelayPoller запускает фоновый long-polling за ответами
// администратора. Ничего не делает, если TELEGRAM_BOT_TOKEN не задан.
func startTelegramRelayPoller() {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return
	}
	go func() {
		offset := latestTelegramUpdateOffset(token)
		for {
			updates, err := fetchTelegramUpdates(token, offset)
			if err != nil {
				time.Sleep(5 * time.Second)
				continue
			}
			for _, u := range updates {
				offset = u.UpdateID + 1
				handleTelegramReply(token, u.Message)
			}
		}
	}()
}

// latestTelegramUpdateOffset подтягивает текущий курсор обновлений, чтобы
// после рестарта сервера не переигрывать старые сообщения из Telegram.
func latestTelegramUpdateOffset(token string) int {
	updates, err := fetchTelegramUpdates(token, 0)
	if err != nil || len(updates) == 0 {
		return 0
	}
	last := updates[len(updates)-1]
	return last.UpdateID + 1
}

func fetchTelegramUpdates(token string, offset int) ([]tgUpdate, error) {
	u := fmt.Sprintf("https://api.telegram.org/bot%s/getUpdates?offset=%d&timeout=25", token, offset)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(u)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var parsed tgUpdatesResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}
	return parsed.Result, nil
}

func handleTelegramReply(token string, msg tgMessage) {
	if msg.ReplyToMessage == nil || msg.Text == "" {
		return
	}
	var confirmation string
	if userID, ok := store.ResolveRelay(msg.ReplyToMessage.MessageID); ok {
		store.AddSupportMessage(&SupportMessage{UserID: userID, Role: SupportRoleAssistant, Body: msg.Text})
		confirmation = "✅ Отправлено пользователю"
	} else if guestID, ok := store.ResolveGuestRelay(msg.ReplyToMessage.MessageID); ok {
		store.AddGuestMessage(guestID, &SupportMessage{Role: SupportRoleAssistant, Body: msg.Text})
		confirmation = "✅ Отправлено в чат на сайте"
	} else {
		return
	}

	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	body, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": confirmation})
	resp, err := telegramClient.Post(fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token), "application/json", bytes.NewReader(body))
	if err == nil {
		resp.Body.Close()
	}
}
