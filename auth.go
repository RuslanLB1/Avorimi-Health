package main

import (
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const sessionCookieName = "avorimi_session"
const sessionTTL = 30 * 24 * time.Hour

type sessionEntry struct {
	UserID    int
	ExpiresAt time.Time
}

// SessionManager хранит активные сессии в памяти — при перезапуске сервера все выходят из аккаунта.
type SessionManager struct {
	mu       sync.Mutex
	sessions map[string]sessionEntry
}

func NewSessionManager() *SessionManager {
	return &SessionManager{sessions: map[string]sessionEntry{}}
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

func (sm *SessionManager) Create(userID int) string {
	token := newToken()
	sm.mu.Lock()
	sm.sessions[token] = sessionEntry{UserID: userID, ExpiresAt: time.Now().Add(sessionTTL)}
	sm.mu.Unlock()
	return token
}

func (sm *SessionManager) UserID(token string) (int, bool) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	e, ok := sm.sessions[token]
	if !ok || e.ExpiresAt.Before(time.Now()) {
		return 0, false
	}
	return e.UserID, true
}

func (sm *SessionManager) Destroy(token string) {
	sm.mu.Lock()
	delete(sm.sessions, token)
	sm.mu.Unlock()
}

func hashPassword(pw string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(pw), bcrypt.DefaultCost)
	return string(b), err
}

func checkPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func normalizePhone(phone string) string {
	return strings.TrimSpace(phone)
}

func setSessionCookie(w http.ResponseWriter, token string) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
}

// currentUser возвращает залогиненного пользователя, если сессия валидна.
func currentUser(r *http.Request) (*User, bool) {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil || cookie.Value == "" {
		return nil, false
	}
	userID, ok := sessions.UserID(cookie.Value)
	if !ok {
		return nil, false
	}
	return store.GetUser(userID)
}

// requireAuth оборачивает обработчик, требуя авторизации; иначе редиректит на /login.
func requireAuth(next func(w http.ResponseWriter, r *http.Request, user *User)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := currentUser(r)
		if !ok {
			http.Redirect(w, r, "/login?next="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next(w, r, user)
	}
}
