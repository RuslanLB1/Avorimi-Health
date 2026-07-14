package main

import (
	"fmt"
	"math"
	"math/rand"
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
	Rating      float64
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
	ID             int
	IIN            string
	FullName       string
	Phone          string
	PasswordHash   string
	CreatedAt      time.Time
	ConsentAt      time.Time // когда пользователь принял соглашения
	ConsentVersion string    // версия документов на момент согласия (см. consentVersion в handlers.go)
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

// specialtyDef описывает одно направление (обследование/приём), которое разные
// клиники могут предлагать через разных врачей с разной ценой и рейтингом.
type specialtyDef struct {
	Category string
	Type     ItemType
	Emoji    string
	Base     int
	Duration string
}

var specialtyPool = []specialtyDef{
	{"Терапевт", TypeDoctor, "🩺", 8000, "30 мин"},
	{"Кардиолог", TypeDoctor, "❤️", 12000, "40 мин"},
	{"Дерматолог", TypeDoctor, "🧴", 10000, "30 мин"},
	{"Невролог", TypeDoctor, "🧠", 11000, "40 мин"},
	{"Стоматолог", TypeDoctor, "🦷", 9000, "40 мин"},
	{"Уролог", TypeDoctor, "🩺", 10500, "30 мин"},
	{"Гинеколог", TypeDoctor, "🩺", 11000, "30 мин"},
	{"Педиатр", TypeDoctor, "🧸", 8500, "30 мин"},
	{"Офтальмолог", TypeDoctor, "👁️", 9500, "30 мин"},
	{"Отоларинголог (ЛОР)", TypeDoctor, "👂", 9000, "30 мин"},
	{"Эндокринолог", TypeDoctor, "🩺", 11500, "30 мин"},
	{"Гастроэнтеролог", TypeDoctor, "🩺", 11000, "30 мин"},
	{"Аллерголог", TypeDoctor, "🤧", 9500, "30 мин"},
	{"Психотерапевт", TypeDoctor, "🧘", 13000, "50 мин"},
	{"Хирург", TypeDoctor, "🩹", 12500, "30 мин"},
	{"Ортопед", TypeDoctor, "🦴", 11000, "30 мин"},
	{"Флеболог", TypeDoctor, "🦵", 10500, "30 мин"},
	{"Ревматолог", TypeDoctor, "🩺", 10500, "30 мин"},
	{"Маммолог", TypeDoctor, "🩺", 11000, "30 мин"},
	{"Диетолог", TypeDoctor, "🥗", 9000, "30 мин"},
	{"УЗИ брюшной полости", TypeProcedure, "🩻", 9000, "20 мин"},
	{"УЗИ малого таза", TypeProcedure, "🩻", 9500, "20 мин"},
	{"ЭКГ с расшифровкой", TypeProcedure, "📈", 4500, "15 мин"},
	{"Общий анализ крови", TypeProcedure, "🧪", 3500, "10 мин"},
	{"Биохимический анализ крови", TypeProcedure, "🧪", 5500, "10 мин"},
	{"Массаж спины", TypeProcedure, "💆", 7000, "45 мин"},
	{"Рентген", TypeProcedure, "🩻", 6000, "15 мин"},
	{"Флюорография", TypeProcedure, "🩻", 3000, "10 мин"},
	{"Вакцинация", TypeProcedure, "💉", 4000, "15 мин"},
}

var clinicSeedNames = []string{
	"Экомед на Абая", "Da Vinci Clinic", "Sova Clinic", "Family Health",
	"Асыл Ана", "GMS Clinic Almaty", "МедЮнион", "Invitro Almaty",
	"Клиника Vita", "Merey Med", "Almaty Health Center", "Салют Мед",
	"Emir Med", "Shipager Clinic", "Best Health Clinic", "Медцентр Дару",
	"Bereke Clinic", "Клиника Радуга", "Fortis Med", "Клиника Айболит",
	"Nova Clinic", "Медцентр Жету", "Клиника Максимед", "City Clinic Almaty",
	"Медицинский центр Достар",
}

var clinicSeedAddresses = []string{
	"ул. Абая, 10", "мкр. Самал-2, 89", "ул. Розыбакиева, 247", "пр. Райымбека, 348",
	"ул. Жандосова, 98", "ул. Наурызбай батыра, 44", "ул. Толе би, 187", "пр. Достык, 105",
	"ул. Гоголя, 86", "ул. Сатпаева, 30а", "мкр. Коктем-3, 12", "ул. Тимирязева, 42",
	"ул. Байтурсынова, 74", "пр. Аль-Фараби, 15", "ул. Жарокова, 240", "мкр. Аксай-3, 8",
	"ул. Момышулы, 15", "ул. Сейфуллина, 520", "ул. Кабанбай батыра, 111", "мкр. Орбита-2, 5",
	"ул. Утеген батыра, 76", "ул. Мынбаева, 43", "пр. Абылай хана, 63", "ул. Панфилова, 98",
	"мкр. Мамыр-4, 21",
}

var clinicEmojis = []string{"🏥", "🏨", "🏩"}

var doctorFirstNames = []string{
	"Айгерим", "Марат", "Динара", "Ержан", "Сауле", "Нурлан", "Гульнара", "Асхат",
	"Айдана", "Данияр", "Жанна", "Бекзат", "Индира", "Тимур", "Алия", "Санжар",
	"Мадина", "Ерлан", "Камила", "Арман", "Дана", "Олжас", "Аружан", "Бауыржан",
}

var doctorLastNames = []string{
	"Сатпаева", "Ким", "Абенова", "Беков", "Ищанова", "Жаксыбеков", "Оспанова", "Тулегенов",
	"Смагулова", "Касымов", "Нурланова", "Оразов", "Байжанова", "Кенжебаев", "Абдразакова",
	"Сейтказиева", "Тлеубергенов", "Ахметова", "Байбосынов", "Дюсенова", "Калиев", "Жумабекова",
	"Сарсенов", "Утешева",
}

func (s *Store) seed() {
	rng := rand.New(rand.NewSource(42))

	for i, name := range clinicSeedNames {
		id := i + 1
		s.clinics[id] = &Clinic{
			ID:          id,
			Name:        name,
			Address:     "г. Алматы, " + clinicSeedAddresses[i],
			Lat:         43.15 + rng.Float64()*0.15,
			Lng:         76.83 + rng.Float64()*0.15,
			Emoji:       clinicEmojis[i%len(clinicEmojis)],
			Rating:      math.Round((4.2+rng.Float64()*0.7)*10) / 10,
			Description: "Многопрофильная клиника-партнёр Avorimi Health.",
		}
	}

	itemID := 1
	for clinicID := 1; clinicID <= len(clinicSeedNames); clinicID++ {
		perm := rng.Perm(len(specialtyPool))
		numSpecialties := 5 + rng.Intn(4) // 5..8 направлений на клинику
		for si := 0; si < numSpecialties; si++ {
			spec := specialtyPool[perm[si]]
			numDoctors := 3 + rng.Intn(2) // 3..4 врача на направление
			for d := 0; d < numDoctors; d++ {
				name := doctorFirstNames[rng.Intn(len(doctorFirstNames))] + " " + doctorLastNames[rng.Intn(len(doctorLastNames))]
				price := spec.Base + (rng.Intn(7)-2)*500
				if price < 1000 {
					price = spec.Base
				}
				desc := spec.Category + " — приём и консультация специалиста."
				if spec.Type == TypeProcedure {
					desc = spec.Category + " — быстро, без очередей, с выдачей результата."
				}
				s.items[itemID] = &Item{
					ID:          itemID,
					ClinicID:    clinicID,
					Type:        spec.Type,
					Name:        name,
					Category:    spec.Category,
					Price:       price,
					Duration:    spec.Duration,
					Description: desc,
					Emoji:       spec.Emoji,
					Rating:      math.Round((4.3+rng.Float64()*0.7)*10) / 10,
				}
				itemID++
			}
		}
	}

	// Генерируем слоты на ближайшие 5 дней для каждого врача/процедуры.
	hours := []int{9, 10, 11, 13, 14, 15, 16, 17}
	now := time.Now()
	for _, it := range s.items {
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

// CategoryGroup — одно обследование/направление в клинике со сводкой по врачам.
type CategoryGroup struct {
	Category  string
	Emoji     string
	Type      ItemType
	Count     int
	MinPrice  int
	MaxRating float64
}

// CategoriesByClinic группирует врачей/процедуры клиники по направлению —
// именно такой список видит пациент, зайдя в карточку клиники.
func (s *Store) CategoriesByClinic(clinicID int) []CategoryGroup {
	s.mu.Lock()
	defer s.mu.Unlock()
	groups := map[string]*CategoryGroup{}
	for _, it := range s.items {
		if it.ClinicID != clinicID {
			continue
		}
		g, ok := groups[it.Category]
		if !ok {
			g = &CategoryGroup{Category: it.Category, Emoji: it.Emoji, Type: it.Type, MinPrice: it.Price}
			groups[it.Category] = g
		}
		g.Count++
		if it.Price < g.MinPrice {
			g.MinPrice = it.Price
		}
		if it.Rating > g.MaxRating {
			g.MaxRating = it.Rating
		}
	}
	out := make([]CategoryGroup, 0, len(groups))
	for _, g := range groups {
		out = append(out, *g)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Category < out[j].Category })
	return out
}

// ItemsByClinicCategory — конкретные врачи (3-4 варианта) по направлению в клинике,
// отсортированные по рейтингу.
func (s *Store) ItemsByClinicCategory(clinicID int, category string) []*Item {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Item, 0)
	for _, it := range s.items {
		if it.ClinicID == clinicID && it.Category == category {
			out = append(out, it)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Rating > out[j].Rating })
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

func (s *Store) CreateUser(iin, fullName, phone, passwordHash, consentVersion string) (*User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.usersByPhone[phone]; exists {
		return nil, fmt.Errorf("пользователь с таким номером телефона уже зарегистрирован")
	}
	u := &User{
		ID:             s.nextUserID,
		IIN:            iin,
		FullName:       fullName,
		Phone:          phone,
		PasswordHash:   passwordHash,
		CreatedAt:      time.Now(),
		ConsentAt:      time.Now(),
		ConsentVersion: consentVersion,
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

// BookingsCount — общее число записей (для статистики на главной).
func (s *Store) BookingsCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.bookings)
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
