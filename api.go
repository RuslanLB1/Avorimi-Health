package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
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

// apiOrWebUser принимает и мобильное приложение (Bearer-токен), и сайт
// (cookie-сессия) — чат поддержки встроен и туда, и туда на одних эндпоинтах.
func apiOrWebUser(r *http.Request) (*User, bool) {
	if user, ok := apiUser(r); ok {
		return user, true
	}
	return currentUser(r)
}

func requireSupportAuth(next func(w http.ResponseWriter, r *http.Request, user *User)) http.HandlerFunc {
	return withCORS(func(w http.ResponseWriter, r *http.Request) {
		user, ok := apiOrWebUser(r)
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
	AcceptTerms     bool   `json:"acceptTerms"`
	AcceptPdn       bool   `json:"acceptPdn"`
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
	if !req.AcceptTerms {
		writeErr(w, http.StatusBadRequest, "err.mustAcceptTerms")
		return
	}
	if !req.AcceptPdn {
		writeErr(w, http.StatusBadRequest, "err.mustAcceptPdn")
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
	user, err := store.CreateUser(req.IIN, req.FullName, phone, hash, currentPolicyVersion)
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

// --- Анализы / результаты ---

type apiLabResultView struct {
	Booking *Booking `json:"booking"`
	Item    *Item    `json:"item"`
	Slot    *Slot    `json:"slot"`
	Ready   bool     `json:"ready"`
}

func apiResultsHandler(w http.ResponseWriter, r *http.Request, user *User) {
	results := labResultsForUser(user.ID)
	views := make([]apiLabResultView, 0, len(results))
	for _, lr := range results {
		views = append(views, apiLabResultView{Booking: lr.Booking, Item: lr.Item, Slot: lr.Slot, Ready: lr.Ready})
	}
	writeJSON(w, http.StatusOK, views)
}

// --- Поддержка (чат-бот) ---

type supportAttachmentView struct {
	URL  string `json:"url"`
	Name string `json:"name"`
	Kind string `json:"kind"`
}

type supportMessageView struct {
	ID          int                     `json:"id"`
	Role        string                  `json:"role"`
	Body        string                  `json:"body"`
	Options     []string                `json:"options,omitempty"`
	Attachments []supportAttachmentView `json:"attachments,omitempty"`
	ReplyToID   int                     `json:"replyToId,omitempty"`
	ReplyToBody string                  `json:"replyToBody,omitempty"`
	CreatedAt   time.Time               `json:"createdAt"`
}

func toSupportMessageView(m *SupportMessage) supportMessageView {
	atts := make([]supportAttachmentView, 0, len(m.Attachments))
	for _, a := range m.Attachments {
		url := ""
		switch {
		case a.FileID != "":
			url = "/api/support/files/" + a.FileID
		case a.UploadID != "":
			url = "/api/support/uploads/" + a.UploadID
		}
		atts = append(atts, supportAttachmentView{URL: url, Name: a.Name, Kind: a.Kind})
	}
	return supportMessageView{
		ID: m.ID, Role: string(m.Role), Body: m.Body, Options: m.Options, Attachments: atts,
		ReplyToID: m.ReplyToID, ReplyToBody: m.ReplyToBody, CreatedAt: m.CreatedAt,
	}
}

// quotePreview обрезает текст цитируемого сообщения для превью (в чате и в
// пересылке в Telegram) — полный текст цитировать незачем.
func quotePreview(text string) string {
	text = strings.TrimSpace(text)
	const limit = 160
	r := []rune(text)
	if len(r) <= limit {
		return text
	}
	return string(r[:limit]) + "…"
}

func apiListSupportMessagesHandler(w http.ResponseWriter, r *http.Request, user *User) {
	msgs := store.SupportMessagesForUser(user.ID)
	views := make([]supportMessageView, 0, len(msgs))
	for _, m := range msgs {
		views = append(views, toSupportMessageView(m))
	}
	writeJSON(w, http.StatusOK, views)
}

// apiResetSupportChatHandler — «Начать новый чат»: сбрасывает шаг гид-бота и
// присылает приветствие с меню заново, не удаляя историю переписки (старые
// сообщения остаются доступны, в том числе для ответа через ReplyToID).
func apiResetSupportChatHandler(w http.ResponseWriter, r *http.Request, user *User) {
	store.SetSupportStage(user.ID, "")
	greeting := mainMenuGreeting(user)
	msg := store.AddSupportMessage(&SupportMessage{UserID: user.ID, Role: SupportRoleAssistant, Body: greeting.Body, Options: greeting.Options})
	writeJSON(w, http.StatusCreated, toSupportMessageView(msg))
}

type sendSupportMessageRequest struct {
	Body               string `json:"body"`
	ReplyToID          int    `json:"replyToId,omitempty"`
	AttachmentUploadID string `json:"attachmentUploadId,omitempty"`
	AttachmentKind     string `json:"attachmentKind,omitempty"`
	AttachmentName     string `json:"attachmentName,omitempty"`
}

func apiSendSupportMessageHandler(w http.ResponseWriter, r *http.Request, user *User) {
	var req sendSupportMessageRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (strings.TrimSpace(req.Body) == "" && req.AttachmentUploadID == "") {
		writeErr(w, http.StatusBadRequest, "err.supportBodyRequired")
		return
	}

	// Явный ответ на конкретное сообщение (в т.ч. на старое, из вчерашнего
	// разговора) — находим его текст, чтобы процитировать в Telegram: так
	// администратор сразу понимает контекст, даже если это та же проблема,
	// что была вчера и вроде бы решилась.
	var replyToBody string
	if req.ReplyToID > 0 {
		for _, m := range store.SupportMessagesForUser(user.ID) {
			if m.ID == req.ReplyToID {
				replyToBody = quotePreview(m.Body)
				break
			}
		}
	}

	var attachments []SupportAttachment
	if req.AttachmentUploadID != "" {
		attachments = append(attachments, SupportAttachment{UploadID: req.AttachmentUploadID, Name: req.AttachmentName, Kind: req.AttachmentKind})
	}

	userMsg := store.AddSupportMessage(&SupportMessage{
		UserID: user.ID, Role: SupportRoleUser, Body: req.Body, Attachments: attachments,
		ReplyToID: req.ReplyToID, ReplyToBody: replyToBody,
	})

	var replyText string
	var replyOptions []string

	if req.AttachmentUploadID != "" {
		// Вложение всегда уходит напрямую оператору (фото/видео/голосовое/файл
		// не укладываются в кнопочное меню) — текущий шаг гид-бота не трогаем.
		caption := req.Body
		if replyToBody != "" {
			caption = fmt.Sprintf("↩️ В ответ на «%s»\n\n%s", replyToBody, req.Body)
		}
		upload, ok := store.GetUpload(req.AttachmentUploadID)
		if !ok {
			replyText = "Не получилось передать файл оператору: вложение не найдено."
		} else if err := forwardSupportAttachmentToTelegram(user, req.AttachmentKind, upload.Name, upload.Data, caption); err != nil {
			log.Printf("[support] не удалось передать вложение оператору: %v", err)
			replyText = "Не получилось передать файл оператору: " + err.Error()
		} else {
			replyText = "Файл передан оператору — он ответит здесь же, в чате."
		}
	} else if replyToBody != "" {
		// Ответ на конкретное сообщение всегда уходит напрямую оператору с
		// цитатой — это не часть кнопочного меню, текущий шаг гид-бота не трогаем.
		quoted := fmt.Sprintf("↩️ В ответ на «%s»\n\n%s", replyToBody, req.Body)
		go forwardSupportQuestionToTelegram(user, quoted)
		replyText = "Передал ваш вопрос оператору с учётом контекста — он ответит здесь же, в чате."
	} else {
		stage := store.SupportStage(user.ID)
		outcome := runSupportFlow(user, stage, req.Body)
		if outcome.handled() {
			replyText = outcome.Reply.Body
			replyOptions = outcome.Reply.Options
			store.SetSupportStage(user.ID, outcome.Reply.Stage)
		} else {
			// Свободный текст вне гид-бота — regex-классификатор (или Anthropic, если подключён).
			history := store.SupportMessagesForUser(user.ID)
			aiHistory := make([]AIMessage, 0, len(history))
			for _, m := range history {
				aiHistory = append(aiHistory, AIMessage{Role: m.Role, Body: m.Body})
			}
			text, err := ai.Reply(aiHistory, req.Body)
			if err != nil {
				text = "Не получилось получить ответ ассистента — попробуйте ещё раз чуть позже."
			}
			replyText = text
			if feedbackRe.MatchString(req.Body) {
				go notifyTelegramFeedback(user, req.Body)
			}
			if registrationTroubleRe.MatchString(req.Body) {
				go notifyTelegramRelayable(user, fmt.Sprintf("🆘 Пользователь застрял на регистрации/входе\nОт: %s (%s)\n\n%s\n\n— Ответьте на это сообщение (Reply), ответ уйдёт пользователю в чат.", user.FullName, user.Phone, req.Body))
			}
			// Пока не подключён платный Anthropic API — дублируем вопрос админу в
			// Telegram; Reply на это сообщение придёт пользователю как ответ поддержки.
			go forwardSupportQuestionToTelegram(user, req.Body)
		}
	}

	reply := store.AddSupportMessage(&SupportMessage{UserID: user.ID, Role: SupportRoleAssistant, Body: replyText, Options: replyOptions})
	writeJSON(w, http.StatusCreated, map[string]any{"userMessage": toSupportMessageView(userMsg), "reply": toSupportMessageView(reply)})
}

type guestSupportRequest struct {
	GuestID            string `json:"guestId,omitempty"` // пусто при первом сообщении — сервер создаст новую сессию
	Name               string `json:"name,omitempty"`
	Contact            string `json:"contact,omitempty"`
	Message            string `json:"message"`
	ReplyToID          int    `json:"replyToId,omitempty"`
	AttachmentUploadID string `json:"attachmentUploadId,omitempty"`
	AttachmentKind     string `json:"attachmentKind,omitempty"`
	AttachmentName     string `json:"attachmentName,omitempty"`
}

// apiGuestSupportHandler — обращения в поддержку от тех, кто не может
// зарегистрироваться/войти. Реального аккаунта нет, поэтому переписка живёт
// в отдельном гостевом хранилище под сгенерированным guestID (фронт хранит
// его в localStorage и опрашивает apiGuestMessagesHandler на новые ответы —
// админ отвечает Reply'ем в Telegram, и это уходит прямо в чат на сайте.
func apiGuestSupportHandler(w http.ResponseWriter, r *http.Request) {
	var req guestSupportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || (strings.TrimSpace(req.Message) == "" && req.AttachmentUploadID == "") {
		writeErr(w, http.StatusBadRequest, "err.supportBodyRequired")
		return
	}
	guestID := strings.TrimSpace(req.GuestID)
	if guestID == "" {
		guestID = newToken()[:24]
	}

	name := strings.TrimSpace(req.Name)
	contact := strings.TrimSpace(req.Contact)
	if name != "" || contact != "" {
		store.SetGuestContact(guestID, name, contact)
	}
	gc, hasContact := store.GetGuestContact(guestID)
	if !hasContact {
		writeErr(w, http.StatusBadRequest, "err.supportContactRequired")
		return
	}

	var replyToBody string
	if req.ReplyToID > 0 {
		for _, m := range store.GuestMessages(guestID) {
			if m.ID == req.ReplyToID {
				replyToBody = quotePreview(m.Body)
				break
			}
		}
	}

	var attachments []SupportAttachment
	if req.AttachmentUploadID != "" {
		attachments = append(attachments, SupportAttachment{UploadID: req.AttachmentUploadID, Name: req.AttachmentName, Kind: req.AttachmentKind})
	}

	userMsg := store.AddGuestMessage(guestID, &SupportMessage{
		Role: SupportRoleUser, Body: req.Message, Attachments: attachments,
		ReplyToID: req.ReplyToID, ReplyToBody: replyToBody,
	})

	if req.AttachmentUploadID != "" {
		caption := req.Message
		if replyToBody != "" {
			caption = fmt.Sprintf("↩️ В ответ на «%s»\n\n%s", replyToBody, req.Message)
		}
		upload, ok := store.GetUpload(req.AttachmentUploadID)
		if !ok {
			store.AddGuestMessage(guestID, &SupportMessage{Role: SupportRoleAssistant, Body: "Не получилось передать файл оператору: вложение не найдено."})
		} else if err := forwardGuestAttachmentToTelegram(guestID, gc.Name, gc.Contact, req.AttachmentKind, upload.Name, upload.Data, caption); err != nil {
			log.Printf("[support] не удалось передать гостевое вложение оператору: %v", err)
			store.AddGuestMessage(guestID, &SupportMessage{Role: SupportRoleAssistant, Body: "Не получилось передать файл оператору: " + err.Error()})
		}
	} else {
		question := req.Message
		if replyToBody != "" {
			question = fmt.Sprintf("↩️ В ответ на «%s»\n\n%s", replyToBody, req.Message)
		}
		go forwardGuestQuestionToTelegram(guestID, gc.Name, gc.Contact, question)
	}

	writeJSON(w, http.StatusCreated, map[string]any{"guestId": guestID, "message": toSupportMessageView(userMsg)})
}

func apiGuestMessagesHandler(w http.ResponseWriter, r *http.Request) {
	guestID := strings.TrimSpace(r.URL.Query().Get("guestId"))
	if guestID == "" {
		writeErr(w, http.StatusBadRequest, "err.supportGuestIDRequired")
		return
	}
	msgs := store.GuestMessages(guestID)
	views := make([]supportMessageView, 0, len(msgs))
	for _, m := range msgs {
		views = append(views, toSupportMessageView(m))
	}
	writeJSON(w, http.StatusOK, views)
}
