package mw_session

import (
	"context"
	"net/http"

	"github.com/google/uuid"
)

type contextKey string

const SessionKey contextKey = "session_id"

// SessionMiddleware überprüft, ob ein Session-Cookie vorhanden ist,
// und erstellt bei Bedarf eine neue SessionID.
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			// Erstelle eine neue SessionID
			sessionID := uuid.New().String()
			// Speichere die SessionID in einem Cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: sessionID,
				Path:  "/",
				// Optional: Setze weitere Cookie-Attribute wie Secure, HttpOnly, etc.
			})
			// Setze die SessionID im Kontext des Requests
			r = r.WithContext(context.WithValue(r.Context(), SessionKey, sessionID))
		} else {
			// Setze die vorhandene SessionID im Kontext des Requests
			r = r.WithContext(context.WithValue(r.Context(), SessionKey, cookie.Value))
		}
		next.ServeHTTP(w, r)
	})
}
