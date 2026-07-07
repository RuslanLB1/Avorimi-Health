package main

import (
	"net/http"
	"strconv"
)

// BookingView объединяет запись, услугу и слот для удобного отображения в шаблонах.
type BookingView struct {
	Booking *Booking
	Item    *Item
	Slot    *Slot
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	all := store.AllItems()
	var doctors, procedures []*Item
	for _, it := range all {
		if it.Type == TypeDoctor {
			doctors = append(doctors, it)
		} else {
			procedures = append(procedures, it)
		}
	}
	if len(doctors) > 4 {
		doctors = doctors[:4]
	}
	if len(procedures) > 4 {
		procedures = procedures[:4]
	}

	data := map[string]any{
		"Doctors":    doctors,
		"Procedures": procedures,
		"Plans":      store.AllPlans(),
	}
	render(w, "home.html", data)
}

func catalogHandler(w http.ResponseWriter, r *http.Request) {
	activeType := r.URL.Query().Get("type")
	activeCategory := r.URL.Query().Get("category")

	all := store.AllItems()
	categorySet := map[string]bool{}
	filtered := make([]*Item, 0, len(all))
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

	for _, it := range all {
		if activeType != "" && string(it.Type) != activeType {
			continue
		}
		if activeCategory != "" && it.Category != activeCategory {
			continue
		}
		filtered = append(filtered, it)
	}

	data := map[string]any{
		"Items":          filtered,
		"Categories":     categories,
		"ActiveType":     activeType,
		"ActiveCategory": activeCategory,
	}
	render(w, "catalog.html", data)
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
	slots := store.SlotsForItem(id)

	data := map[string]any{
		"Item":  item,
		"Slots": slots,
	}
	render(w, "item.html", data)
}

func bookingFormHandler(w http.ResponseWriter, r *http.Request) {
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

	data := map[string]any{
		"Item":  item,
		"Slot":  slot,
		"Error": r.URL.Query().Get("error"),
	}
	render(w, "book.html", data)
}

func createBookingHandler(w http.ResponseWriter, r *http.Request) {
	slotID, err := strconv.Atoi(r.PathValue("slotID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректные данные формы", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	useSubscription := r.FormValue("use_subscription") == "on"

	if name == "" || phone == "" {
		http.Redirect(w, r, "/book/"+r.PathValue("slotID")+"?error=Заполните+имя+и+телефон", http.StatusSeeOther)
		return
	}

	booking, err := store.CreateBooking(slotID, name, phone, useSubscription)
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

func paymentPageHandler(w http.ResponseWriter, r *http.Request) {
	bookingID, err := strconv.Atoi(r.PathValue("bookingID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	booking, ok := store.GetBooking(bookingID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	if booking.Status != StatusPending {
		http.Redirect(w, r, "/success/"+strconv.Itoa(bookingID), http.StatusSeeOther)
		return
	}
	item, _ := store.GetItem(booking.ItemID)

	data := map[string]any{
		"Booking": booking,
		"Item":    item,
	}
	render(w, "pay.html", data)
}

func processPaymentHandler(w http.ResponseWriter, r *http.Request) {
	bookingID, err := strconv.Atoi(r.PathValue("bookingID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := store.MarkPaid(bookingID); err != nil {
		http.NotFound(w, r)
		return
	}
	http.Redirect(w, r, "/success/"+strconv.Itoa(bookingID), http.StatusSeeOther)
}

func successHandler(w http.ResponseWriter, r *http.Request) {
	bookingID, err := strconv.Atoi(r.PathValue("bookingID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	booking, ok := store.GetBooking(bookingID)
	if !ok {
		http.NotFound(w, r)
		return
	}
	item, _ := store.GetItem(booking.ItemID)
	slot, _ := store.GetSlot(booking.SlotID)

	data := map[string]any{
		"Booking": booking,
		"Item":    item,
		"Slot":    slot,
	}
	render(w, "success.html", data)
}

func subscriptionsHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Plans": store.AllPlans(),
	}
	render(w, "subscriptions.html", data)
}

func subscribeFormHandler(w http.ResponseWriter, r *http.Request) {
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
	data := map[string]any{
		"Plan":  plan,
		"Error": r.URL.Query().Get("error"),
	}
	render(w, "subscribe.html", data)
}

func subscribeToPaymentHandler(w http.ResponseWriter, r *http.Request) {
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
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректные данные формы", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	phone := r.FormValue("phone")
	if name == "" || phone == "" {
		http.Redirect(w, r, "/subscribe/"+r.PathValue("planID")+"?error=Заполните+имя+и+телефон", http.StatusSeeOther)
		return
	}

	data := map[string]any{
		"Plan":  plan,
		"Name":  name,
		"Phone": phone,
	}
	render(w, "subscribe_pay.html", data)
}

func confirmSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	planID, err := strconv.Atoi(r.PathValue("planID"))
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "некорректные данные формы", http.StatusBadRequest)
		return
	}
	phone := r.FormValue("phone")

	sub, err := store.CreateSubscription(phone, planID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	plan, _ := store.GetPlan(planID)

	data := map[string]any{
		"Plan":         plan,
		"Subscription": sub,
	}
	render(w, "subscribe_success.html", data)
}

func accountHandler(w http.ResponseWriter, r *http.Request) {
	phone := r.URL.Query().Get("phone")

	data := map[string]any{
		"Phone": phone,
	}

	if phone != "" {
		bookings := store.BookingsByPhone(phone)
		views := make([]BookingView, 0, len(bookings))
		for _, b := range bookings {
			item, _ := store.GetItem(b.ItemID)
			slot, _ := store.GetSlot(b.SlotID)
			views = append(views, BookingView{Booking: b, Item: item, Slot: slot})
		}
		data["Bookings"] = views

		if sub, ok := store.ActiveSubscription(phone); ok {
			plan, _ := store.GetPlan(sub.PlanID)
			data["Subscription"] = sub
			data["SubPlan"] = plan
		}
	}

	render(w, "account.html", data)
}
