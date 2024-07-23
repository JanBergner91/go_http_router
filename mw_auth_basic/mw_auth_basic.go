package mw_auth_basic

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

// Credential repräsentiert eine Benutzername-Passwort-Kombination
type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// BasicAuthMiddleware ist eine Middleware, die Basic-Authentifizierung erfordert.
func BasicAuthMiddleware(credFilePath string) func(http.Handler) http.Handler {
	// Lade die Anmeldeinformationen aus der JSON-Datei
	credentials, err := loadCredentials(credFilePath)
	if err != nil {
		fmt.Println("Error loading credentials:", err)
		return nil
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !validateBasicAuth(authHeader, credentials) {
				// Wenn die Authentifizierung fehlschlägt, gib 401 Unauthorized zurück
				w.Header().Set("WWW-Authenticate", `Basic realm="Please provide credentials"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			// Wenn die Authentifizierung erfolgreich ist, fahre mit dem nächsten Handler fort
			next.ServeHTTP(w, r)
		})
	}
}

// loadCredentials lädt die Anmeldeinformationen aus einer JSON-Datei
func loadCredentials(filePath string) ([]Credential, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var credentials []Credential
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&credentials); err != nil {
		return nil, err
	}

	return credentials, nil
}

// validateBasicAuth überprüft, ob der Authorization-Header gültige Anmeldeinformationen enthält
func validateBasicAuth(authHeader string, credentials []Credential) bool {
	if authHeader == "" {
		return false
	}

	// Erwarte, dass der Authorization-Header im Format "Basic <credentials>" ist
	if !strings.HasPrefix(authHeader, "Basic ") {
		return false
	}

	// Extrahiere die Basis64-kodierten Anmeldeinformationen
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return false
	}

	// Trenne Benutzername und Passwort
	parts := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(parts) != 2 {
		return false
	}

	username := parts[0]
	password := parts[1]

	// Überprüfe die Anmeldeinformationen gegen die geladenen Credentials
	for _, cred := range credentials {
		if username == cred.Username && password == cred.Password {
			return true
		}
	}

	return false
}
