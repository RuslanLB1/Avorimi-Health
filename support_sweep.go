package main

import "time"

// support_sweep.go — автоматическое закрытие чата поддержки после 5 минут
// молчания пользователя: бот присылает уведомление и переводит диалог в
// стадию stageClosed. Следующее сообщение пользователя (в любое время)
// снова начинается с приветствия и меню — как в первый раз.

const supportInactivityTimeout = 5 * time.Minute

const closedChatNotice = "Чат закрыт из-за отсутствия сообщений. Если появятся вопросы — просто напишите снова, я подскажу заново 🙂"

func startSupportInactivitySweeper(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for range ticker.C {
			sweepInactiveSupportChats()
		}
	}()
}

func sweepInactiveSupportChats() {
	for _, userID := range store.AllSupportUserIDs() {
		if store.SupportStage(userID) == stageClosed {
			continue
		}
		msgs := store.SupportMessagesForUser(userID)
		if len(msgs) == 0 {
			continue
		}
		last := msgs[len(msgs)-1]
		if time.Since(last.CreatedAt) < supportInactivityTimeout {
			continue
		}
		store.AddSupportMessage(&SupportMessage{UserID: userID, Role: SupportRoleAssistant, Body: closedChatNotice})
		store.SetSupportStage(userID, stageClosed)
	}
}
