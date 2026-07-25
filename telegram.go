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

// telegram.go — уведомления в Telegram о важных событиях платформы: отзывы,
// идеи, проблемы с регистрацией, критичные ошибки.
//
// Чтобы включить:
//  1. Создайте бота через @BotFather, получите TELEGRAM_BOT_TOKEN.
//  2. Напишите боту любое сообщение, затем узнайте свой chat_id (например,
//     через @userinfobot) и задайте его в TELEGRAM_CHAT_ID.
//  3. Задайте обе переменные окружения на Render — уведомления заработают
//     без изменений кода.
var telegramClient = &http.Client{Timeout: 10 * time.Second}

// notifyTelegram отправляет произвольный текст в личный чат администратора.
// Если переменные окружения не заданы — событие просто уходит в лог, ничего не падает.
func notifyTelegram(text string) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	if token == "" || chatID == "" {
		log.Printf("[telegram] не настроен (TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID) — событие только в логах: %s", text)
		return
	}

	body, _ := json.Marshal(map[string]string{"chat_id": chatID, "text": text})
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	resp, err := telegramClient.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		log.Printf("[telegram] не удалось отправить уведомление: %v", err)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("[telegram] сервер вернул статус %d", resp.StatusCode)
	}
}

func notifyTelegramFeedback(user *User, message string) {
	notifyTelegram(fmt.Sprintf("💬 Новый отзыв на Avorimi Health\nОт: %s (%s)\n\n%s", user.FullName, user.Phone, message))
}

func notifyTelegramIdea(user *User, message string) {
	notifyTelegram(fmt.Sprintf("💡 Идея от %s (%s)\n\n%s", user.FullName, user.Phone, message))
}

func notifyTelegramNewBooking(user *User, item *Item, booking *Booking) {
	notifyTelegram(fmt.Sprintf("📅 Новая запись\nПациент: %s (%s)\nНа приём: %s\nСумма: %d ₸", user.FullName, user.Phone, item.Name, booking.Price))
}

func notifyTelegramCritical(context string, err error) {
	notifyTelegram(fmt.Sprintf("🚨 Критическая ошибка на Avorimi Health\n%s\n%v", context, err))
}

// recoverMiddleware ловит панику в обработчиках, шлёт алерт в Telegram и
// отвечает 500 вместо падения процесса целиком.
func recoverMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("[panic] %s %s: %v", r.Method, r.URL.Path, rec)
				notifyTelegramCritical(fmt.Sprintf("%s %s", r.Method, r.URL.Path), fmt.Errorf("%v", rec))
				http.Error(w, "внутренняя ошибка сервера", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
