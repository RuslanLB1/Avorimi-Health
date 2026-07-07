package main

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

// ItemType различает врачей и процедуры в едином каталоге.
type ItemType string

const (
	TypeDoctor    ItemType = "doctor"
	TypeProcedure ItemType = "procedure"
)

type Item struct {
	ID          int
	Type        ItemType
	Name        string
	Category    string
	Price       int
	Duration    string
	Description string
	Emoji       string
	Rating      float64
}

type Slot struct {
	ID     int
	ItemID int
	When   time.Time
	Booked bool
}

type BookingStatus string

const (
	StatusPending BookingStatus = "pending"
	StatusPaid    BookingStatus = "paid"
	StatusFree    BookingStatus = "free_by_subscription"
)

type Booking struct {
	ID        int
	SlotID    int
	ItemID    int
	Name      string
	Phone     string
	Price     int
	Status    BookingStatus
	CreatedAt time.Time
}

type SubscriptionPlan struct {
	ID          int
	Name        string
	Price       string
	PriceValue  int
	Visits      int
	PeriodDays  int
	Description string
	Highlight   bool
}

type UserSubscription struct {
	Phone       string
	PlanID      int
	VisitsLeft  int
	PurchasedAt time.Time
	ExpiresAt   time.Time
}

// Store — простое потокобезопасное in-memory хранилище для демо-прототипа.
type Store struct {
	mu            sync.Mutex
	items         map[int]*Item
	slots         map[int]*Slot
	bookings      map[int]*Booking
	plans         map[int]*SubscriptionPlan
	subscriptions map[string]*UserSubscription // key: phone

	nextBookingID int
	nextSlotID    int
}

func NewStore() *Store {
	s := &Store{
		items:         map[int]*Item{},
		slots:         map[int]*Slot{},
		bookings:      map[int]*Booking{},
		plans:         map[int]*SubscriptionPlan{},
		subscriptions: map[string]*UserSubscription{},
		nextBookingID: 1,
		nextSlotID:    1,
	}
	s.seed()
	return s
}

func (s *Store) seed() {
	items := []*Item{
		{ID: 1, Type: TypeDoctor, Name: "Айгерим Сатпаева", Category: "Терапевт", Price: 8000, Duration: "30 мин", Description: "Первичный приём, консультация, назначение обследований.", Emoji: "🩺", Rating: 4.9},
		{ID: 2, Type: TypeDoctor, Name: "Марат Ким", Category: "Кардиолог", Price: 12000, Duration: "40 мин", Description: "Консультация кардиолога, расшифровка ЭКГ.", Emoji: "❤️", Rating: 4.8},
		{ID: 3, Type: TypeDoctor, Name: "Динара Абенова", Category: "Дерматолог", Price: 10000, Duration: "30 мин", Description: "Диагностика кожи, консультация по высыпаниям и родинкам.", Emoji: "🧴", Rating: 4.7},
		{ID: 4, Type: TypeDoctor, Name: "Ержан Беков", Category: "Невролог", Price: 11000, Duration: "40 мин", Description: "Приём невролога, консультация при головных болях и головокружении.", Emoji: "🧠", Rating: 4.9},
		{ID: 5, Type: TypeProcedure, Name: "УЗИ брюшной полости", Category: "Диагностика", Price: 9000, Duration: "20 мин", Description: "Комплексное ультразвуковое обследование органов брюшной полости.", Emoji: "🩻", Rating: 4.8},
		{ID: 6, Type: TypeProcedure, Name: "Общий анализ крови", Category: "Анализы", Price: 3500, Duration: "10 мин", Description: "Забор крови и полный клинический анализ.", Emoji: "🧪", Rating: 4.9},
		{ID: 7, Type: TypeProcedure, Name: "Массаж спины", Category: "Физиотерапия", Price: 7000, Duration: "45 мин", Description: "Лечебный массаж спины и шейно-воротниковой зоны.", Emoji: "💆", Rating: 4.7},
		{ID: 8, Type: TypeProcedure, Name: "ЭКГ с расшифровкой", Category: "Диагностика", Price: 4500, Duration: "15 мин", Description: "Электрокардиограмма с заключением врача.", Emoji: "📈", Rating: 4.8},
	}
	for _, it := range items {
		s.items[it.ID] = it
	}

	// Генерируем слоты на ближайшие 5 дней для каждого специалиста/процедуры.
	hours := []int{9, 10, 11, 13, 14, 15, 16, 17}
	now := time.Now()
	for _, it := range items {
		for d := 0; d < 5; d++ {
			day := now.AddDate(0, 0, d)
			for _, h := range hours {
				when := time.Date(day.Year(), day.Month(), day.Day(), h, 0, 0, 0, day.Location())
				if when.Before(now) {
					continue
				}
				slot := &Slot{ID: s.nextSlotID, ItemID: it.ID, When: when}
				s.slots[slot.ID] = slot
				s.nextSlotID++
			}
		}
	}

	plans := []*SubscriptionPlan{
		{ID: 1, Name: "Старт", Price: "14 900 ₸/мес", PriceValue: 14900, Visits: 3, PeriodDays: 30, Description: "3 бесплатных визита в месяц: приёмы врачей или процедуры на выбор."},
		{ID: 2, Name: "Комфорт", Price: "24 900 ₸/мес", PriceValue: 24900, Visits: 6, PeriodDays: 30, Description: "6 бесплатных визитов в месяц и приоритетная запись.", Highlight: true},
		{ID: 3, Name: "Премиум", Price: "39 900 ₸/мес", PriceValue: 39900, Visits: 12, PeriodDays: 30, Description: "12 визитов в месяц, включая узких специалистов, без доплат."},
	}
	for _, p := range plans {
		s.plans[p.ID] = p
	}
}

func (s *Store) AllItems() []*Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Item, 0, len(s.items))
	for _, it := range s.items {
		out = append(out, it)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) GetItem(id int) (*Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	it, ok := s.items[id]
	return it, ok
}

func (s *Store) SlotsForItem(itemID int) []*Slot {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Slot, 0)
	for _, sl := range s.slots {
		if sl.ItemID == itemID && !sl.Booked && sl.When.After(time.Now()) {
			out = append(out, sl)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].When.Before(out[j].When) })
	return out
}

func (s *Store) GetSlot(id int) (*Slot, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sl, ok := s.slots[id]
	return sl, ok
}

func (s *Store) AllPlans() []*SubscriptionPlan {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*SubscriptionPlan, 0, len(s.plans))
	for _, p := range s.plans {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) GetPlan(id int) (*SubscriptionPlan, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	p, ok := s.plans[id]
	return p, ok
}

// ActiveSubscription возвращает подписку по номеру телефона, если она ещё активна.
func (s *Store) ActiveSubscription(phone string) (*UserSubscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subscriptions[phone]
	if !ok || sub.VisitsLeft <= 0 || sub.ExpiresAt.Before(time.Now()) {
		return nil, false
	}
	return sub, true
}

func (s *Store) CreateSubscription(phone string, planID int) (*UserSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.plans[planID]
	if !ok {
		return nil, fmt.Errorf("план не найден")
	}
	sub := &UserSubscription{
		Phone:       phone,
		PlanID:      planID,
		VisitsLeft:  plan.Visits,
		PurchasedAt: time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, plan.PeriodDays),
	}
	s.subscriptions[phone] = sub
	return sub, nil
}

// CreateBooking резервирует слот и создаёт запись. useSubscription списывает визит с подписки.
func (s *Store) CreateBooking(slotID int, name, phone string, useSubscription bool) (*Booking, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	slot, ok := s.slots[slotID]
	if !ok || slot.Booked {
		return nil, fmt.Errorf("слот недоступен")
	}
	item, ok := s.items[slot.ItemID]
	if !ok {
		return nil, fmt.Errorf("услуга не найдена")
	}

	status := StatusPending
	price := item.Price

	if useSubscription {
		sub, ok := s.subscriptions[phone]
		if !ok || sub.VisitsLeft <= 0 || sub.ExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("нет активной подписки с доступными визитами")
		}
		sub.VisitsLeft--
		status = StatusFree
		price = 0
	}

	slot.Booked = true
	booking := &Booking{
		ID:        s.nextBookingID,
		SlotID:    slotID,
		ItemID:    item.ID,
		Name:      name,
		Phone:     phone,
		Price:     price,
		Status:    status,
		CreatedAt: time.Now(),
	}
	s.bookings[booking.ID] = booking
	s.nextBookingID++
	return booking, nil
}

func (s *Store) GetBooking(id int) (*Booking, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[id]
	return b, ok
}

func (s *Store) MarkPaid(bookingID int) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	b, ok := s.bookings[bookingID]
	if !ok {
		return fmt.Errorf("запись не найдена")
	}
	if b.Status == StatusPending {
		b.Status = StatusPaid
	}
	return nil
}

// BookingsByPhone для страницы "Мои записи".
func (s *Store) BookingsByPhone(phone string) []*Booking {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Booking, 0)
	for _, b := range s.bookings {
		if b.Phone == phone {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}
