package main

import (
	"fmt"
	"math"
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

// Clinic — медицинский центр-партнёр. Avorimi Health выступает мостом между
// клиникой и пациентом: у каждой клиники свой набор врачей/процедур (Item.ClinicID).
type Clinic struct {
	ID          int
	Name        string
	Address     string
	Lat         float64
	Lng         float64
	Emoji       string
	Description string
}

type Item struct {
	ID          int
	ClinicID    int
	Type        ItemType
	Name        string
	Category    string
	Price       int
	Duration    string
	Description string
	Emoji       string
	Rating      float64
}

// haversineKm считает расстояние по прямой между двумя точками в километрах —
// используется, чтобы сортировать клиники по близости к геолокации пользователя.
func haversineKm(lat1, lng1, lat2, lng2 float64) float64 {
	const earthRadiusKm = 6371.0
	dLat := (lat2 - lat1) * math.Pi / 180
	dLng := (lng2 - lng1) * math.Pi / 180
	a := math.Sin(dLat/2)*math.Sin(dLat/2) +
		math.Cos(lat1*math.Pi/180)*math.Cos(lat2*math.Pi/180)*math.Sin(dLng/2)*math.Sin(dLng/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	return earthRadiusKm * c
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
	UserID    int
	SlotID    int
	ItemID    int
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
	UserID      int
	PlanID      int
	VisitsLeft  int
	PurchasedAt time.Time
	ExpiresAt   time.Time
}

// User — зарегистрированный аккаунт платформы.
type User struct {
	ID           int
	IIN          string
	FullName     string
	Phone        string
	PasswordHash string
	CreatedAt    time.Time
}

// Store — простое потокобезопасное in-memory хранилище для демо-прототипа.
type Store struct {
	mu            sync.Mutex
	clinics       map[int]*Clinic
	items         map[int]*Item
	slots         map[int]*Slot
	bookings      map[int]*Booking
	plans         map[int]*SubscriptionPlan
	users         map[int]*User
	usersByPhone  map[string]int
	subscriptions map[int]*UserSubscription // key: userID (latest/active)

	nextBookingID int
	nextSlotID    int
	nextUserID    int
}

func NewStore() *Store {
	s := &Store{
		clinics:       map[int]*Clinic{},
		items:         map[int]*Item{},
		slots:         map[int]*Slot{},
		bookings:      map[int]*Booking{},
		plans:         map[int]*SubscriptionPlan{},
		users:         map[int]*User{},
		usersByPhone:  map[string]int{},
		subscriptions: map[int]*UserSubscription{},
		nextBookingID: 1,
		nextSlotID:    1,
		nextUserID:    1,
	}
	s.seed()
	return s
}

func (s *Store) seed() {
	clinics := []*Clinic{
		{ID: 1, Name: "Экомед на Абая", Address: "г. Алматы, ул. Абая, 10", Lat: 43.2380, Lng: 76.9010, Emoji: "🏥", Description: "Многопрофильная клиника в центре города."},
		{ID: 2, Name: "Da Vinci Clinic", Address: "г. Алматы, мкр. Самал-2, 89", Lat: 43.2200, Lng: 76.9550, Emoji: "🏨", Description: "Частная клиника с узкими специалистами и диагностикой."},
		{ID: 3, Name: "Sova Clinic", Address: "г. Алматы, ул. Розыбакиева, 247", Lat: 43.2050, Lng: 76.8900, Emoji: "🏩", Description: "Семейная клиника с удобной записью день в день."},
		{ID: 4, Name: "Family Health", Address: "г. Алматы, пр. Райымбека, 348", Lat: 43.2630, Lng: 76.9450, Emoji: "🏥", Description: "Педиатрия, терапия и лабораторная диагностика."},
		{ID: 5, Name: "Асыл Ана", Address: "г. Алматы, ул. Жандосова, 98", Lat: 43.1950, Lng: 76.8650, Emoji: "🏨", Description: "Женское здоровье и общая терапия."},
		{ID: 6, Name: "GMS Clinic Almaty", Address: "г. Алматы, ул. Наурызбай батыра, 44", Lat: 43.2500, Lng: 76.9450, Emoji: "🏩", Description: "Современная многопрофильная клиника премиум-класса."},
	}
	for _, c := range clinics {
		s.clinics[c.ID] = c
	}

	items := []*Item{
		{ID: 1, ClinicID: 1, Type: TypeDoctor, Name: "Айгерим Сатпаева", Category: "Терапевт", Price: 8000, Duration: "30 мин", Description: "Первичный приём, консультация, назначение обследований.", Emoji: "🩺", Rating: 4.9},
		{ID: 2, ClinicID: 1, Type: TypeDoctor, Name: "Нурлан Жаксыбеков", Category: "Стоматолог", Price: 9000, Duration: "40 мин", Description: "Осмотр, консультация, лечение и профилактика кариеса.", Emoji: "🦷", Rating: 4.8},
		{ID: 3, ClinicID: 1, Type: TypeProcedure, Name: "УЗИ брюшной полости", Category: "Диагностика", Price: 9000, Duration: "20 мин", Description: "Комплексное ультразвуковое обследование органов брюшной полости.", Emoji: "🩻", Rating: 4.8},

		{ID: 4, ClinicID: 2, Type: TypeDoctor, Name: "Марат Ким", Category: "Кардиолог", Price: 12000, Duration: "40 мин", Description: "Консультация кардиолога, расшифровка ЭКГ.", Emoji: "❤️", Rating: 4.8},
		{ID: 5, ClinicID: 2, Type: TypeDoctor, Name: "Асхат Тулегенов", Category: "Уролог", Price: 10500, Duration: "30 мин", Description: "Консультация уролога, УЗИ по показаниям.", Emoji: "🩺", Rating: 4.7},
		{ID: 6, ClinicID: 2, Type: TypeProcedure, Name: "ЭКГ с расшифровкой", Category: "Диагностика", Price: 4500, Duration: "15 мин", Description: "Электрокардиограмма с заключением врача.", Emoji: "📈", Rating: 4.8},

		{ID: 7, ClinicID: 3, Type: TypeDoctor, Name: "Динара Абенова", Category: "Дерматолог", Price: 10000, Duration: "30 мин", Description: "Диагностика кожи, консультация по высыпаниям и родинкам.", Emoji: "🧴", Rating: 4.7},
		{ID: 8, ClinicID: 3, Type: TypeDoctor, Name: "Сауле Ищанова", Category: "Гинеколог", Price: 11000, Duration: "30 мин", Description: "Плановый осмотр и консультация гинеколога.", Emoji: "🩺", Rating: 4.9},
		{ID: 9, ClinicID: 3, Type: TypeProcedure, Name: "Массаж спины", Category: "Физиотерапия", Price: 7000, Duration: "45 мин", Description: "Лечебный массаж спины и шейно-воротниковой зоны.", Emoji: "💆", Rating: 4.7},

		{ID: 10, ClinicID: 4, Type: TypeDoctor, Name: "Ержан Беков", Category: "Невролог", Price: 11000, Duration: "40 мин", Description: "Приём невролога, консультация при головных болях и головокружении.", Emoji: "🧠", Rating: 4.9},
		{ID: 11, ClinicID: 4, Type: TypeDoctor, Name: "Гульнара Оспанова", Category: "Педиатр", Price: 8500, Duration: "30 мин", Description: "Осмотр и консультация детского врача.", Emoji: "🧸", Rating: 4.9},
		{ID: 12, ClinicID: 4, Type: TypeProcedure, Name: "Общий анализ крови", Category: "Анализы", Price: 3500, Duration: "10 мин", Description: "Забор крови и полный клинический анализ.", Emoji: "🧪", Rating: 4.9},

		{ID: 13, ClinicID: 5, Type: TypeDoctor, Name: "Бекзат Оразов", Category: "Терапевт", Price: 7500, Duration: "30 мин", Description: "Первичная консультация терапевта.", Emoji: "🩺", Rating: 4.6},
		{ID: 14, ClinicID: 5, Type: TypeDoctor, Name: "Айдана Смагулова", Category: "Стоматолог", Price: 8500, Duration: "40 мин", Description: "Лечение и профилактика заболеваний зубов и дёсен.", Emoji: "🦷", Rating: 4.7},

		{ID: 15, ClinicID: 6, Type: TypeDoctor, Name: "Данияр Касымов", Category: "Уролог", Price: 13000, Duration: "30 мин", Description: "Консультация уролога в клинике премиум-класса.", Emoji: "🩺", Rating: 4.9},
		{ID: 16, ClinicID: 6, Type: TypeProcedure, Name: "УЗИ малого таза", Category: "Диагностика", Price: 9500, Duration: "20 мин", Description: "Ультразвуковое исследование органов малого таза.", Emoji: "🩻", Rating: 4.8},
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

// --- Клиники ---

func (s *Store) AllClinics() []*Clinic {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Clinic, 0, len(s.clinics))
	for _, c := range s.clinics {
		out = append(out, c)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (s *Store) GetClinic(id int) (*Clinic, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.clinics[id]
	return c, ok
}

func (s *Store) ItemsByClinic(clinicID int) []*Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Item, 0)
	for _, it := range s.items {
		if it.ClinicID == clinicID {
			out = append(out, it)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
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

// HasSlotToday сообщает, есть ли у услуги свободное время сегодня (для фильтра "сегодня").
func (s *Store) HasSlotToday(itemID int) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now()
	for _, sl := range s.slots {
		if sl.ItemID == itemID && !sl.Booked && sl.When.After(now) && sameDay(sl.When, now) {
			return true
		}
	}
	return false
}

func sameDay(a, b time.Time) bool {
	ay, am, ad := a.Date()
	by, bm, bd := b.Date()
	return ay == by && am == bm && ad == bd
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

// --- Пользователи ---

func (s *Store) CreateUser(iin, fullName, phone, passwordHash string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.usersByPhone[phone]; exists {
		return nil, fmt.Errorf("пользователь с таким номером телефона уже зарегистрирован")
	}
	u := &User{
		ID:           s.nextUserID,
		IIN:          iin,
		FullName:     fullName,
		Phone:        phone,
		PasswordHash: passwordHash,
		CreatedAt:    time.Now(),
	}
	s.users[u.ID] = u
	s.usersByPhone[phone] = u.ID
	s.nextUserID++
	return u, nil
}

func (s *Store) GetUser(id int) (*User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	u, ok := s.users[id]
	return u, ok
}

func (s *Store) GetUserByPhone(phone string) (*User, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id, ok := s.usersByPhone[phone]
	if !ok {
		return nil, false
	}
	return s.users[id], true
}

// --- Подписки ---

// ActiveSubscription возвращает подписку пользователя, если она ещё активна.
func (s *Store) ActiveSubscription(userID int) (*UserSubscription, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sub, ok := s.subscriptions[userID]
	if !ok || sub.VisitsLeft <= 0 || sub.ExpiresAt.Before(time.Now()) {
		return nil, false
	}
	return sub, true
}

func (s *Store) CreateSubscription(userID, planID int) (*UserSubscription, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	plan, ok := s.plans[planID]
	if !ok {
		return nil, fmt.Errorf("план не найден")
	}
	sub := &UserSubscription{
		UserID:      userID,
		PlanID:      planID,
		VisitsLeft:  plan.Visits,
		PurchasedAt: time.Now(),
		ExpiresAt:   time.Now().AddDate(0, 0, plan.PeriodDays),
	}
	s.subscriptions[userID] = sub
	return sub, nil
}

// --- Записи ---

// CreateBooking резервирует слот и создаёт запись. useSubscription списывает визит с подписки.
func (s *Store) CreateBooking(slotID, userID int, useSubscription bool) (*Booking, error) {
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
		sub, ok := s.subscriptions[userID]
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
		UserID:    userID,
		SlotID:    slotID,
		ItemID:    item.ID,
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

// BookingsByUser для личного кабинета.
func (s *Store) BookingsByUser(userID int) []*Booking {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Booking, 0)
	for _, b := range s.bookings {
		if b.UserID == userID {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.After(out[j].CreatedAt) })
	return out
}
