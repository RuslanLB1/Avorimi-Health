package main

import (
	"fmt"
	"os"
)

// PaymentProvider — единая точка подключения реального эквайринга.
// Сейчас используется MockProvider (демо-оплата без списаний). Чтобы запустить
// приём настоящих денег, нужно реализовать Charge для нужного банка/агрегатора
// и подставить его в main.go вместо MockProvider — остальной код (handlers.go)
// менять не придётся.
type PaymentProvider interface {
	// Charge выполняет списание amount тенге и возвращает ID транзакции.
	Charge(amount int, description string) (transactionID string, err error)
	Name() string
}

// MockProvider имитирует оплату — используется, пока нет боевых ключей банка.
type MockProvider struct{}

func (MockProvider) Name() string { return "mock" }

func (MockProvider) Charge(amount int, description string) (string, error) {
	return "mock_" + newToken()[:12], nil
}

// KaspiProvider — заготовка под Kaspi Pay / Kaspi Business API.
// Чтобы включить реальные платежи через Kaspi:
//  1. Оформить приём платежей в Kaspi Business и получить merchantID + API-ключ.
//  2. Задать переменные окружения KASPI_MERCHANT_ID и KASPI_API_KEY на Render.
//  3. Реализовать Charge() по документации Kaspi Pay API (создание счёта/платежа,
//     редирект пользователя на оплату, обработка вебхука подтверждения).
//  4. В main.go заменить payments = MockProvider{} на payments = NewKaspiProvider().
//
// По такой же схеме можно добавить HalykProvider (Epay от Halyk Bank) или любой
// другой банк/агрегатора — интерфейс PaymentProvider у всех один и тот же.
type KaspiProvider struct {
	MerchantID string
	APIKey     string
}

func NewKaspiProvider() *KaspiProvider {
	return &KaspiProvider{
		MerchantID: os.Getenv("KASPI_MERCHANT_ID"),
		APIKey:     os.Getenv("KASPI_API_KEY"),
	}
}

func (p *KaspiProvider) Name() string { return "kaspi" }

func (p *KaspiProvider) Charge(amount int, description string) (string, error) {
	if p.MerchantID == "" || p.APIKey == "" {
		return "", fmt.Errorf("Kaspi не подключён: заполните KASPI_MERCHANT_ID и KASPI_API_KEY")
	}
	return "", fmt.Errorf("интеграция с Kaspi Pay API ещё не реализована — см. комментарий в payments.go")
}
