package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	store    *Store
	sessions *SessionManager
	payments PaymentProvider
	tmpl     *template.Template
)

func main() {
	store = NewStore()
	sessions = NewSessionManager()
	payments = MockProvider{}

	funcMap := template.FuncMap{
		"money": func(v int) string {
			s := strconv.Itoa(v)
			out := ""
			for i, c := range s {
				if i != 0 && (len(s)-i)%3 == 0 {
					out += " "
				}
				out += string(c)
			}
			return out + " ₸"
		},
		"when": func(t time.Time) string {
			months := [...]string{"января", "февраля", "марта", "апреля", "мая", "июня", "июля", "августа", "сентября", "октября", "ноября", "декабря"}
			return t.Format("02 ") + months[t.Month()-1] + t.Format(", 15:04")
		},
		"initial": func(name string) string {
			for _, r := range name {
				return string(r)
			}
			return "?"
		},
	}

	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("ошибка разбора шаблонов: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /clinics", clinicsHandler)
	mux.HandleFunc("GET /clinic/{id}", clinicDetailHandler)
	mux.HandleFunc("GET /catalog", catalogHandler)
	mux.HandleFunc("GET /item/{id}", itemDetailHandler)

	mux.HandleFunc("GET /register", registerFormHandler)
	mux.HandleFunc("POST /register", registerSubmitHandler)
	mux.HandleFunc("GET /login", loginFormHandler)
	mux.HandleFunc("POST /login", loginSubmitHandler)
	mux.HandleFunc("POST /logout", logoutHandler)

	mux.HandleFunc("GET /book/{slotID}", requireAuth(bookingFormHandler))
	mux.HandleFunc("POST /book/{slotID}", requireAuth(createBookingHandler))

	mux.HandleFunc("GET /pay/{bookingID}", requireAuth(paymentPageHandler))
	mux.HandleFunc("POST /pay/{bookingID}", requireAuth(processPaymentHandler))

	mux.HandleFunc("GET /success/{bookingID}", requireAuth(successHandler))

	mux.HandleFunc("GET /subscriptions", subscriptionsHandler)
	mux.HandleFunc("GET /subscribe/{planID}", requireAuth(subscribeFormHandler))
	mux.HandleFunc("POST /subscribe/{planID}", requireAuth(confirmSubscriptionHandler))

	mux.HandleFunc("GET /account", requireAuth(accountHandler))

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	addr := ":" + port
	log.Printf("Avorimi Health запущен: http://localhost%s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

// render рендерит шаблон, подмешивая в data информацию о текущем пользователе
// (для профиля/подписки в шапке сайта на любой странице).
func render(w http.ResponseWriter, r *http.Request, name string, data map[string]any) {
	if data == nil {
		data = map[string]any{}
	}
	if user, ok := currentUser(r); ok {
		data["CurrentUser"] = user
		if sub, ok := store.ActiveSubscription(user.ID); ok {
			plan, _ := store.GetPlan(sub.PlanID)
			data["ActiveSub"] = sub
			data["ActiveSubPlan"] = plan
		}
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
