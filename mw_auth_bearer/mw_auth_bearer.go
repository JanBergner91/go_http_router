package mw_auth_bearer

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// BearerAuthMiddleware ist eine Middleware, die Bearer-Authentifizierung erfordert.
func BearerAuthMiddleware(tokenFilePath string) func(http.Handler) http.Handler {
	// Lade die gültigen Tokens aus der JSON-Datei
	tokens, err := loadTokens(tokenFilePath)
	if err != nil {
		fmt.Println("Error loading tokens:", err)
		return nil
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !validateBearerToken(authHeader, tokens) {
				// Wenn die Authentifizierung fehlschlägt, gib 401 Unauthorized zurück
				w.Header().Set("WWW-Authenticate", `Bearer realm="Please provide a valid token"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Wenn die Authentifizierung erfolgreich ist, fahre mit dem nächsten Handler fort
			next.ServeHTTP(w, r)
		})
	}
}

// loadTokens lädt die gültigen Bearer-Tokens aus einer JSON-Datei
func loadTokens(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tokens []string
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// validateBearerToken überprüft, ob der Authorization-Header ein gültiges Bearer-Token enthält
func validateBearerToken(authHeader string, validTokens []string) bool {
	if authHeader == "" {
		return false
	}

	// Erwarte, dass der Authorization-Header im Format "Bearer <token>" ist
	if !strings.HasPrefix(authHeader, "Bearer ") {
		return false
	}

	// Extrahiere das Token
	token := strings.TrimPrefix(authHeader, "Bearer ")
	// Überprüfe das Token gegen die Liste der gültigen Tokens
	for _, validToken := range validTokens {
		if token == validToken {
			return true
		}
	}

	return false
}
