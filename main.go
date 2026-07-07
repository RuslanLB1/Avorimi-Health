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
	store *Store
	tmpl  *template.Template
)

func main() {
	store = NewStore()

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
	}

	var err error
	tmpl, err = template.New("").Funcs(funcMap).ParseGlob("templates/*.html")
	if err != nil {
		log.Fatalf("ошибка разбора шаблонов: %v", err)
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	mux.HandleFunc("GET /{$}", homeHandler)
	mux.HandleFunc("GET /catalog", catalogHandler)
	mux.HandleFunc("GET /item/{id}", itemDetailHandler)

	mux.HandleFunc("GET /book/{slotID}", bookingFormHandler)
	mux.HandleFunc("POST /book/{slotID}", createBookingHandler)

	mux.HandleFunc("GET /pay/{bookingID}", paymentPageHandler)
	mux.HandleFunc("POST /pay/{bookingID}", processPaymentHandler)

	mux.HandleFunc("GET /success/{bookingID}", successHandler)

	mux.HandleFunc("GET /subscriptions", subscriptionsHandler)
	mux.HandleFunc("GET /subscribe/{planID}", subscribeFormHandler)
	mux.HandleFunc("POST /subscribe/{planID}", subscribeToPaymentHandler)
	mux.HandleFunc("POST /subscribe/{planID}/pay", confirmSubscriptionHandler)

	mux.HandleFunc("GET /account", accountHandler)

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

func render(w http.ResponseWriter, name string, data any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, name, data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
