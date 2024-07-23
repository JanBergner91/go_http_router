package mw_auth_multi

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// Credential repräsentiert eine Benutzername-Passwort-Kombination
type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// BearerToken repräsentiert ein gültiges Token
type BearerToken string

// MultiAuthMiddleware ist eine Middleware, die Basic Auth, Bearer Auth und Active Directory Auth unterstützt.
func MultiAuthMiddleware(basicCredFilePath, bearerTokenFilePath, adServer, adBaseDN, adUserDN, adPassword string) func(http.Handler) http.Handler {
	// Lade die Anmeldeinformationen und Tokens
	basicCredentials, err := loadBasicCredentials(basicCredFilePath)
	if err != nil {
		fmt.Println("Error loading basic credentials:", err)
		return nil
	}

	bearerTokens, err := loadBearerTokens(bearerTokenFilePath)
	if err != nil {
		fmt.Println("Error loading bearer tokens:", err)
		return nil
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if !validateAuth(authHeader, basicCredentials, bearerTokens, adServer, adBaseDN, adUserDN, adPassword) {
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

// loadBasicCredentials lädt die Anmeldeinformationen für Basic Auth aus einer JSON-Datei
func loadBasicCredentials(filePath string) ([]Credential, error) {
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

// loadBearerTokens lädt die gültigen Bearer-Tokens aus einer JSON-Datei
func loadBearerTokens(filePath string) ([]BearerToken, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var tokens []BearerToken
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(&tokens); err != nil {
		return nil, err
	}

	return tokens, nil
}

// validateAuth überprüft, ob der Authorization-Header entweder Basic Auth, Bearer Token oder Active Directory enthält
func validateAuth(authHeader string, basicCredentials []Credential, bearerTokens []BearerToken, adServer, adBaseDN, adUserDN, adPassword string) bool {
	if authHeader == "" {
		return false
	}

	if strings.HasPrefix(authHeader, "Basic ") {
		// Prüfe Basic Auth
		return validateBasicAuth(authHeader, basicCredentials)
	} else if strings.HasPrefix(authHeader, "Bearer ") {
		// Prüfe Bearer Token
		return validateBearerToken(authHeader, bearerTokens)
	} else if strings.HasPrefix(authHeader, "AD ") {
		// Prüfe Active Directory Auth
		return validateADAuth(authHeader, adServer, adBaseDN, adUserDN, adPassword)
	}

	return false
}

// validateBasicAuth überprüft, ob der Authorization-Header gültige Basic-Authentifizierungsdaten enthält
func validateBasicAuth(authHeader string, credentials []Credential) bool {
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return false
	}

	parts := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(parts) != 2 {
		return false
	}

	username := parts[0]
	password := parts[1]

	for _, cred := range credentials {
		if username == cred.Username && password == cred.Password {
			return true
		}
	}

	return false
}

// validateBearerToken überprüft, ob der Authorization-Header ein gültiges Bearer-Token enthält
func validateBearerToken(authHeader string, validTokens []BearerToken) bool {
	token := strings.TrimPrefix(authHeader, "Bearer ")
	for _, validToken := range validTokens {
		if BearerToken(token) == validToken {
			return true
		}
	}
	return false
}

// validateADAuth überprüft die Benutzeranmeldeinformationen gegen Active Directory
func validateADAuth(authHeader, adServer, adBaseDN, adUserDN, adPassword string) bool {
	if !strings.HasPrefix(authHeader, "AD ") {
		return false
	}

	encodedCredentials := strings.TrimPrefix(authHeader, "AD ")
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return false
	}

	parts := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(parts) != 2 {
		return false
	}

	username := parts[0]
	password := parts[1]

	return validateADCredentials(adServer, adBaseDN, adUserDN, adPassword, username, password)
}

// validateADCredentials überprüft die Benutzeranmeldeinformationen gegen Active Directory
func validateADCredentials(adServer, adBaseDN, adUserDN, adPassword, username, password string) bool {
	l, err := ldap.DialURL(adServer)
	if err != nil {
		fmt.Println("Failed to connect to AD server:", err)
		return false
	}
	defer l.Close()

	err = l.Bind(adUserDN, adPassword)
	if err != nil {
		fmt.Println("Failed to bind to AD server:", err)
		return false
	}

	searchRequest := ldap.NewSearchRequest(
		adBaseDN,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(sAMAccountName=%s)", username),
		[]string{"dn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		fmt.Println("Failed to search user in AD:", err)
		return false
	}

	if len(sr.Entries) != 1 {
		fmt.Println("User not found or multiple users found")
		return false
	}

	userDN := sr.Entries[0].DN
	err = l.Bind(userDN, password)
	if err != nil {
		fmt.Println("Failed to authenticate user:", err)
		return false
	}

	return true
}
