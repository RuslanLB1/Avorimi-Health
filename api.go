package main

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

// api.go — JSON-версия основных сценариев платформы для мобильного приложения
// (React Native / Expo). Переиспользует ту же бизнес-логику (store.go, auth.go,
// payments.go), просто отдаёт JSON вместо HTML-шаблонов.

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, key string) {
	writeJSON(w, status, map[string]string{"error": key})
}

func withCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next(w, r)
	}
}

// apiUser достаёт текущего пользователя из заголовка Authorization: Bearer <token>.
func apiUser(r *http.Request) (*User, bool) {
	auth := r.Header.Get("Authorization")
	token, ok := strings.CutPrefix(auth, "Bearer ")
	if !ok || token == "" {
		return nil, false
	}
	userID, ok := sessions.UserID(token)
	if !ok {
		return nil, false
	}
	return store.GetUser(userID)
}

func requireAPIAuth(next func(w http.ResponseWriter, r *http.Request, user *User)) http.HandlerFunc {
	return withCORS(func(w http.ResponseWriter, r *http.Request) {
		user, ok := apiUser(r)
		if !ok {
			writeErr(w, http.StatusUnauthorized, "err.unauthorized")
			return
		}
		next(w, r, user)
	})
}

// --- Пользователь для JSON-ответов (без PasswordHash) ---

type apiUserView struct {
	ID       int    `json:"id"`
	FullName string `json:"fullName"`
	Phone    string `json:"phone"`
}

func toAPIUser(u *User) apiUserView {
	return apiUserView{ID: u.ID, FullName: u.FullName, Phone: u.Phone}
}

// --- Регистрация / вход ---

type registerRequest struct {
	FullName        string `json:"fullName"`
	IIN             string `json:"iin"`
	PhoneLocal      string `json:"phoneLocal"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirmPassword"`
}

func apiRegisterHandler(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "err.badRequest")
		return
	}
	req.FullName = strings.TrimSpace(req.FullName)
	if req.FullName == "" {
		writeErr(w, http.StatusBadRequest, "err.fillFullName")
		return
	}
	phone, err := buildPhone(req.PhoneLocal)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	if !iinRe.MatchString(req.IIN) {
		writeErr(w, http.StatusBadRequest, "err.iinInvalid")
		return
	}
	if len(req.Password) < 6 {
		writeErr(w, http.StatusBadRequest, "err.passwordTooShort")
		return
	}
	if req.Password != req.ConfirmPassword {
		writeErr(w, http.StatusBadRequest, "err.passwordMismatch")
		return
	}
	hash, err := hashPassword(req.Password)
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "err.generic")
		return
	}
	user, err := store.CreateUser(req.IIN, req.FullName, phone, hash)
	if err != nil {
		writeErr(w, http.StatusConflict, "err.phoneTaken")
		return
	}
	token := sessions.Create(user.ID)
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": toAPIUser(user)})
}

type loginRequest struct {
	PhoneLocal string `json:"phoneLocal"`
	Password   string `json:"password"`
}

func apiLoginHandler(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "err.badRequest")
		return
	}
	phone, err := buildPhone(req.PhoneLocal)
	if err != nil {
		writeErr(w, http.StatusBadRequest, err.Error())
		return
	}
	user, ok := store.GetUserByPhone(phone)
	if !ok || !checkPassword(user.PasswordHash, req.Password) {
		writeErr(w, http.StatusUnauthorized, "err.invalidLogin")
		return
	}
	token := sessions.Create(user.ID)
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": toAPIUser(user)})
}

func apiMeHandler(w http.ResponseWriter, r *http.Request, user *User) {
	resp := map[string]any{"user": toAPIUser(user)}
	if sub, ok := store.ActiveSubscription(user.ID); ok {
		plan, _ := store.GetPlan(sub.PlanID)
		resp["subscription"] = map[string]any{
			"planName":   plan.Name,
			"visitsLeft": sub.VisitsLeft,
			"expiresAt":  sub.ExpiresAt,
		}
	}
	writeJSON(w, http.StatusOK, resp)
}

// --- Клиники ---

type apiClinicView struct {
	*Clinic
	ItemCount  int     `json:"itemCount"`
	DistanceKm float64 `json:"distanceKm,omitempty"`
}

func apiClinicsHandler(w http.ResponseWriter, r *http.Request) {
	latStr := r.URL.Query().Get("lat")
	lngStr := r.URL.Query().Get("lng")
	hasLocation := latStr != "" && lngStr != ""
	var lat, lng float64
	if hasLocation {
		var errLat, errLng error
		lat, errLat = strconv.ParseFloat(latStr, 64)
		lng, errLng = strconv.ParseFloat(lngStr, 64)
		hasLocation = errLat == nil && errLng == nil
	}

	clinics := store.AllClinics()
	views := make([]apiClinicView, 0, len(clinics))
	for _, c := range clinics {
		v := apiClinicView{Clinic: c, ItemCount: len(store.ItemsByClinic(c.ID))}
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
	writeJSON(w, http.StatusOK, views)
}

func apiClinicDetailHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	clinic, ok := store.GetClinic(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	groups := store.CategoriesByClinic(id)
	writeJSON(w, http.StatusOK, map[string]any{"clinic": clinic, "categories": groups})
}

func apiClinicItemsHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	category := r.URL.Query().Get("category")
	items := store.ItemsByClinicCategory(id, category)
	writeJSON(w, http.StatusOK, items)
}

// --- Врачи / процедуры ---

func apiItemDetailHandler(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	item, ok := store.GetItem(id)
	if !ok {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	clinic, _ := store.GetClinic(item.ClinicID)
	slots := store.SlotsForItem(id)
	writeJSON(w, http.StatusOK, map[string]any{"item": item, "clinic": clinic, "slots": slots})
}

// --- Подписки ---

func apiPlansHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, store.AllPlans())
}

func apiSubscribeHandler(w http.ResponseWriter, r *http.Request, user *User) {
	planID, err := strconv.Atoi(r.PathValue("planID"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	sub, err := store.CreateSubscription(user.ID, planID)
	if err != nil {
		writeErr(w, http.StatusNotFound, "err.planNotFound")
		return
	}
	writeJSON(w, http.StatusCreated, sub)
}

// --- Бронирования ---

type bookingRequest struct {
	SlotID          int  `json:"slotId"`
	UseSubscription bool `json:"useSubscription"`
}

func apiCreateBookingHandler(w http.ResponseWriter, r *http.Request, user *User) {
	var req bookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "err.badRequest")
		return
	}
	booking, err := store.CreateBooking(req.SlotID, user.ID, req.UseSubscription)
	if err != nil {
		writeErr(w, http.StatusBadRequest, bookingErrorKey(err))
		return
	}
	writeJSON(w, http.StatusCreated, booking)
}

func apiPayBookingHandler(w http.ResponseWriter, r *http.Request, user *User) {
	bookingID, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	booking, ok := store.GetBooking(bookingID)
	if !ok || booking.UserID != user.ID {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	if _, err := payments.Charge(booking.Price, "Запись №"+strconv.Itoa(booking.ID)); err != nil {
		writeErr(w, http.StatusBadGateway, "err.paymentFailed")
		return
	}
	if err := store.MarkPaid(bookingID); err != nil {
		writeErr(w, http.StatusNotFound, "err.notFound")
		return
	}
	booking, _ = store.GetBooking(bookingID)
	writeJSON(w, http.StatusOK, booking)
}

type apiBookingView struct {
	*Booking
	Item   *Item   `json:"item"`
	Slot   *Slot   `json:"slot"`
	Clinic *Clinic `json:"clinic"`
}

func apiMyBookingsHandler(w http.ResponseWriter, r *http.Request, user *User) {
	bookings := store.BookingsByUser(user.ID)
	views := make([]apiBookingView, 0, len(bookings))
	for _, b := range bookings {
		item, _ := store.GetItem(b.ItemID)
		slot, _ := store.GetSlot(b.SlotID)
		var clinic *Clinic
		if item != nil {
			clinic, _ = store.GetClinic(item.ClinicID)
		}
		views = append(views, apiBookingView{Booking: b, Item: item, Slot: slot, Clinic: clinic})
	}
	writeJSON(w, http.StatusOK, views)
}
