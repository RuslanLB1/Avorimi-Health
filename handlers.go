package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"math/rand"
	"net/http"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// BookingView объединяет запись, услугу и слот для удобного отображения в шаблонах.
type BookingView struct {
	Booking *Booking
	Item    *Item
	Slot    *Slot
}

// PopularDoctor — карточка специалиста для блока "Популярные специалисты" на главной.
type PopularDoctor struct {
	Item       *Item
	ClinicName string
	NextSlot   *Slot
}

// clinicPin — минимальные данные клиники для отрисовки метки на карте (JSON для JS).
type clinicPin struct {
	ID      int     `json:"id"`
	Name    string  `json:"name"`
	Address string  `json:"address"`
	Lat     float64 `json:"lat"`
	Lng     float64 `json:"lng"`
	Rating  float64 `json:"rating"`
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	clinics := store.AllClinics()
	items := store.AllItems()

	var doctorItems []*Item
	for _, it := range items {
		if it.Type == TypeDoctor {
			doctorItems = append(doctorItems, it)
		}
	}
	totalDoctors := len(doctorItems)

	sort.Slice(doctorItems, func(i, j int) bool { return doctorItems[i].Rating > doctorItems[j].Rating })
	if len(doctorItems) > 6 {
		doctorItems = doctorItems[:6]
	}
	popular := make([]PopularDoctor, 0, len(doctorItems))
	for _, it := range doctorItems {
		clinicName := ""
		if c, ok := store.GetClinic(it.ClinicID); ok {
			clinicName = c.Name
		}
		var next *Slot
		if slots := store.SlotsForItem(it.ID); len(slots) > 0 {
			next = slots[0]
		}
		popular = append(popular, PopularDoctor{Item: it, ClinicName: clinicName, NextSlot: next})
	}

	var ratingSum float64
	for _, c := range clinics {
		ratingSum += c.Rating
	}
	avgRating := 0.0
	if len(clinics) > 0 {
		avgRating = ratingSum / float64(len(clinics))
	}

	pins := make([]clinicPin, 0, len(clinics))
	for _, c := range clinics {
		pins = append(pins, clinicPin{ID: c.ID, Name: c.Name, Address: c.Address, Lat: c.Lat, Lng: c.Lng, Rating: c.Rating})
	}
	pinsJSON, _ := json.Marshal(pins)

	render(w, r, "home.html", map[string]any{
		"Plans":          store.AllPlans(),
		"ClinicsCount":   len(clinics),
		"DoctorsCount":   totalDoctors,
		"BookingsCount":  store.BookingsCount(),
		"AvgRating":      avgRating,
		"PopularDoctors": popular,
		"ClinicsJSON":    template.JS(pinsJSON),
	})
}

// ClinicView — клиника вместе с расстоянием до пользователя (если геолокация известна).
type ClinicView struct {
	*Clinic
	DistanceKm float64
	ItemCount  int
}

func clinicsHandler(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	hasLocation := latStr != "" && lngStr != ""
	var lat, lng float64
	if hasLocation {
		var errLat, errLng error
		lat, errLat = strconv.ParseFloat(latStr, 64)
		lng, errLng = strconv.ParseFloat(lngStr, 64)
		if errLat != nil || errLng != nil {
			hasLocation = false
		}
	}

	clinics := store.AllClinics()
	views := make([]ClinicView, 0, len(clinics))
	for _, c := range clinics {
		v := ClinicView{Clinic: c, ItemCount: len(store.ItemsByClinic(c.ID))}
		if hasLocation {
			v.DistanceKm = haversineKm(lat, lng, c.Lat, c.Lng)
		}
		views = append(views, v)
	}
	if hasLocation {
		sort.Slice(views, func(i, j int) bool { return views[i].DistanceKm < views[j].DistanceKm })
	} else {
		rand.Shuffle(len(views), func(i, j int) { views[i], views[j] = views[j], views[i] })
	}

	render(w, r, "clinics.html", map[string]any{
		"Clinics":     views,
		"HasLocation": hasLocation,
	})
}

func clinicDetailHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	clinic, ok := store.GetClinic(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	groups := store.CategoriesByClinic(id)

	render(w, r, "clinic.html", map[string]any{
		"Clinic": clinic,
		"Groups": groups,
	})
}

func clinicCategoryHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	clinic, ok := store.GetClinic(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	category := r.PathValue("category")
	items := store.ItemsByClinicCategory(id, category)
	if len(items) == 0 {
		http.NotFound(w, r)
		return
	}

	render(w, r, "clinic_category.html", map[string]any{
		"Clinic":   clinic,
		"Category": category,
		"Items":    items,
	})
}

// ItemView — врач/процедура вместе с названием клиники (для общего каталога).
type ItemView struct {
	*Item
	ClinicName string
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	activeType := r.URL.Query().Get("type")
	activeCategory := r.URL.Query().Get("category")
	sortBy := r.URL.Query().Get("sort")
	todayOnly := r.URL.Query().Get("today") == "1"

	all := store.AllItems()
	categorySet := map[string]bool{}
	for _, it := range all {
		if activeType != "" && string(it.Type) != activeType {
			continue
		}
		categorySet[it.Category] = true
	}
	categories := make([]string, 0, len(categorySet))
	for c := range categorySet {
		categories = append(categories, c)
	}
	sort.Strings(categories)

	filtered := make([]*Item, 0, len(all))
	for _, it := range all {
		if activeType != "" && string(it.Type) != activeType {
			continue
		}
		if activeCategory != "" && it.Category != activeCategory {
			continue
		}
		if todayOnly && !store.HasSlotToday(it.ID) {
			continue
		}
		filtered = append(filtered, it)
	}

	switch sortBy {
	case "price_asc":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price < filtered[j].Price })
	case "price_desc":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Price > filtered[j].Price })
	case "rating":
		sort.Slice(filtered, func(i, j int) bool { return filtered[i].Rating > filtered[j].Rating })
	}

	views := make([]ItemView, 0, len(filtered))
	for _, it := range filtered {
		clinicName := ""
		if c, ok := store.GetClinic(it.ClinicID); ok {
			clinicName = c.Name
		}
		views = append(views, ItemView{Item: it, ClinicName: clinicName})
	}

	render(w, r, "catalog.html", map[string]any{
		"Items":          views,
		"Categories":     categories,
		"ActiveType":     activeType,
		"ActiveCategory": activeCategory,
		"Sort":           sortBy,
		"Today":          todayOnly,
	})
}

func itemDetailHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	item, ok := store.GetItem(id)
	if !ok {
		http.NotFound(w, r)
		return
	}
	clinic, _ := store.GetClinic(item.ClinicID)
	slots := store.SlotsForItem(id)

	render(w, r, "item.html", map[string]any{
		"Item":   item,
		"Clinic": clinic,
		"Slots":  slots,
	})
}

// --- Регистрация / вход ---

var iinRe = regexp.MustCompile(`^\d{12}$`)
var nonDigitRe = regexp.MustCompile(`\D`)

// buildPhone собирает номер вида +7XXXXXXXXXX из введённых после "+7" цифр.
func buildPhone(local string) (string, error) {
	digits := nonDigitRe.ReplaceAllString(local, "")
	if len(digits) != 10 {
		return "", fmt.Errorf("введите+10+цифр+номера+телефона")
	}
	return "+7" + digits, nil
}

func registerFormHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "register.html", map[string]any{
		"Error": r.URL.Query().Get("error"),
	})
}

func registerSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректные данные формы", http.StatusBadRequest)
		return
	}
	iin := strings.TrimSpace(r.FormValue("iin"))
	fullName := strings.TrimSpace(r.FormValue("full_name"))
	password := r.FormValue("password")
	confirm := r.FormValue("confirm_password")

	fail := func(msg string) {
		http.Redirect(w, r, "/register?error="+msg, http.StatusSeeOther)
	}

	if fullName == "" {
		fail("Заполните+ФИО")
		return
	}
	phone, err := buildPhone(r.FormValue("phone_local"))
	if err != nil {
		fail(err.Error())
		return
	}
	if !iinRe.MatchString(iin) {
		fail("ИИН+должен+содержать+ровно+12+цифр")
		return
	}
	if len(password) < 6 {
		fail("Пароль+должен+быть+не+короче+6+символов")
		return
	}
	if password != confirm {
		fail("Пароли+не+совпадают")
		return
	}

	hash, err := hashPassword(password)
	if err != nil {
		http.Error(w, "не удалось создать аккаунт", http.StatusInternalServerError)
		return
	}
	user, err := store.CreateUser(iin, fullName, phone, hash)
	if err != nil {
		fail("Пользователь+с+таким+телефоном+уже+зарегистрирован")
		return
	}

	token := sessions.Create(user.ID)
	setSessionCookie(w, token)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func loginFormHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "login.html", map[string]any{
		"Error": r.URL.Query().Get("error"),
		"Next":  r.URL.Query().Get("next"),
	})
}

func loginSubmitHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректные данные формы", http.StatusBadRequest)
		return
	}
	password := r.FormValue("password")
	next := r.FormValue("next")

	phone, err := buildPhone(r.FormValue("phone_local"))
	if err != nil {
		target := "/login?error=" + err.Error()
		if next != "" {
			target += "&next=" + next
		}
		http.Redirect(w, r, target, http.StatusSeeOther)
		return
	}

	user, ok := store.GetUserByPhone(phone)
	if !ok || !checkPassword(user.PasswordHash, password) {
		target := "/login?error=Неверный+телефон+или+пароль"
		if next != "" {
			target += "&next=" + next
		}
		http.Redirect(w, r, target, http.StatusSeeOther)
		return
	}

	token := sessions.Create(user.ID)
	setSessionCookie(w, token)
	if next == "" {
		next = "/"
	}
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		sessions.Destroy(cookie.Value)
	}
	clearSessionCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// --- Запись на приём ---

func bookingFormHandler(w http.ResponseWriter, r *http.Request, user *User) {
	slotID, err := strconv.Atoi(r.PathValue("slotID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	slot, ok := store.GetSlot(slotID)
	if !ok || slot.Booked {
		http.Error(w, "Это время уже занято, выберите другое.", http.StatusGone)
		return
	}
	item, ok := store.GetItem(slot.ItemID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	_, hasSub := store.ActiveSubscription(user.ID)

	render(w, r, "book.html", map[string]any{
		"Item":  item,
		"Slot":  slot,
		"HasSub": hasSub,
		"Error": r.URL.Query().Get("error"),
	})
}

func createBookingHandler(w http.ResponseWriter, r *http.Request, user *User) {
	slotID, err := strconv.Atoi(r.PathValue("slotID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректные данные формы", http.StatusBadRequest)
		return
	}
	useSubscription := r.FormValue("use_subscription") == "on"

	booking, err := store.CreateBooking(slotID, user.ID, useSubscription)
	if err != nil {
		http.Redirect(w, r, "/book/"+r.PathValue("slotID")+"?error="+err.Error(), http.StatusSeeOther)
		return
	}

	if booking.Status == StatusFree {
		http.Redirect(w, r, "/success/"+strconv.Itoa(booking.ID), http.StatusSeeOther)
		return
	}
	http.Redirect(w, r, "/pay/"+strconv.Itoa(booking.ID), http.StatusSeeOther)
}

func paymentPageHandler(w http.ResponseWriter, r *http.Request, user *User) {
	bookingID, err := strconv.Atoi(r.PathValue("bookingID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	booking, ok := store.GetBooking(bookingID)
	if !ok || booking.UserID != user.ID {
		http.NotFound(w, r)
		return
	}
	if booking.Status != StatusPending {
		http.Redirect(w, r, "/success/"+strconv.Itoa(bookingID), http.StatusSeeOther)
		return
	}
	item, _ := store.GetItem(booking.ItemID)

	render(w, r, "pay.html", map[string]any{
		"Booking": booking,
		"Item":    item,
	})
}

func processPaymentHandler(w http.ResponseWriter, r *http.Request, user *User) {
	bookingID, err := strconv.Atoi(r.PathValue("bookingID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	booking, ok := store.GetBooking(bookingID)
	if !ok || booking.UserID != user.ID {
		http.NotFound(w, r)
		return
	}
	if _, err := payments.Charge(booking.Price, "Запись №"+strconv.Itoa(booking.ID)); err != nil {
		http.Error(w, "Оплата не прошла: "+err.Error(), http.StatusBadGateway)
		return
	}
	if err := store.MarkPaid(bookingID); err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/success/"+strconv.Itoa(bookingID), http.StatusSeeOther)
}

func successHandler(w http.ResponseWriter, r *http.Request, user *User) {
	bookingID, err := strconv.Atoi(r.PathValue("bookingID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	booking, ok := store.GetBooking(bookingID)
	if !ok || booking.UserID != user.ID {
		http.NotFound(w, r)
		return
	}
	item, _ := store.GetItem(booking.ItemID)
	slot, _ := store.GetSlot(booking.SlotID)

	render(w, r, "success.html", map[string]any{
		"Booking": booking,
		"Item":    item,
		"Slot":    slot,
	})
}

// --- Подписки ---

func subscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	render(w, r, "subscriptions.html", map[string]any{
		"Plans": store.AllPlans(),
	})
}

func subscribeFormHandler(w http.ResponseWriter, r *http.Request, user *User) {
	planID, err := strconv.Atoi(r.PathValue("planID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	plan, ok := store.GetPlan(planID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	render(w, r, "subscribe.html", map[string]any{
		"Plan": plan,
	})
}

func confirmSubscriptionHandler(w http.ResponseWriter, r *http.Request, user *User) {
	planID, err := strconv.Atoi(r.PathValue("planID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	plan, ok := store.GetPlan(planID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if _, err := payments.Charge(plan.PriceValue, "Подписка "+plan.Name); err != nil {
		http.Error(w, "Оплата не прошла: "+err.Error(), http.StatusBadGateway)
		return
	}
	sub, err := store.CreateSubscription(user.ID, planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	render(w, r, "subscribe_success.html", map[string]any{
		"Plan":         plan,
		"Subscription": sub,
	})
}

// --- Личный кабинет ---

func accountHandler(w http.ResponseWriter, r *http.Request, user *User) {
	bookings := store.BookingsByUser(user.ID)
	views := make([]BookingView, 0, len(bookings))
	for _, b := range bookings {
		item, _ := store.GetItem(b.ItemID)
		slot, _ := store.GetSlot(b.SlotID)
		views = append(views, BookingView{Booking: b, Item: item, Slot: slot})
	}

	render(w, r, "account.html", map[string]any{
		"Bookings": views,
	})
}
