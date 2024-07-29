package mw_session

import (
	"context"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type contextKey string

const SessionKey contextKey = "session_id"

/* Default Session Expiration */
var SessionDuration time.Duration = 24 * 60 * time.Minute

// SessionMiddleware überprüft, ob ein Session-Cookie vorhanden ist,
// und erstellt bei Bedarf eine neue SessionID.
func SessionMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("session_id")
		if err != nil || cookie.Value == "" {
			// Erstelle eine neue SessionID
			sessionID := uuid.New().String()
			// Erstelle Time to Expire
			///now := time.Now()
			// Ablaufzeit in 60 Minuten
			///expiration := now.Local().Add(SessionDuration)
			// Speichere die SessionID in einem Cookie
			http.SetCookie(w, &http.Cookie{
				Name:  "session_id",
				Value: sessionID,
				Path:  "/",
				///Expires: expiration,
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

type SessionItem struct {
	Key   string
	Value string
}

var SessionStore = map[string][]SessionItem{}

func AddOrUpdateSessionItem(sessionID string, item SessionItem) {
	if items, exists := SessionStore[sessionID]; exists {
		for i, existingItem := range items {
			if existingItem.Key == item.Key {
				SessionStore[sessionID][i].Value = item.Value
				return
			}
		}
		SessionStore[sessionID] = append(items, item)
	} else {
		SessionStore[sessionID] = []SessionItem{item}
	}
	/*jsonstore, err := json.Marshal(SessionStore)
	if err != nil {
		panic("LOL")
	}
	fmt.Println(string(jsonstore))*/
}

func RemoveSessionItem(sessionID string, keyToRemove string) {
	if items, exists := SessionStore[sessionID]; exists {
		updatedItems := []SessionItem{}
		for _, item := range items {
			if item.Key != keyToRemove {
				updatedItems = append(updatedItems, item)
			}
		}
		if len(updatedItems) == 0 {
			delete(SessionStore, sessionID)
		} else {
			SessionStore[sessionID] = updatedItems
		}
	}
}

func GetSessionItem(sessionID string, keyToInspect string) string {
	if items, exists := SessionStore[sessionID]; exists {
		for i, existingItem := range items {
			if existingItem.Key == keyToInspect {
				return SessionStore[sessionID][i].Value
			}
		}
		return ""
	} else {
		return ""
	}
}
