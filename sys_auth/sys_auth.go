package sys_auth

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"httpr2/mw_session"
	"net/http"
	"os"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

type Credential struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// BearerToken repräsentiert ein gültiges Token
type BearerToken string

// MultiAuthMiddleware ist eine Middleware, die Basic Auth, Bearer Auth und Active Directory Auth unterstützt.
func MultiAuthMiddleware(basicCredFilePath, bearerTokenFilePath, adServer, adBaseDN, adUserDN, adPassword string) func(http.Handler) http.Handler {
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
			z0, _ := validateAuth(authHeader, basicCredentials, bearerTokens, adServer, adBaseDN, adUserDN, adPassword)
			if !z0 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Please provide credentials"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// ADAuthMiddleware ist eine Middleware, die Active Directory-Authentifizierung durchführt.
func ADAuthMiddleware(adServer, adBaseDN, adUserDN, adPassword string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
				w.Header().Set("WWW-Authenticate", `Basic realm="Please provide credentials"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
			decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
			if err != nil {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			parts := strings.SplitN(string(decodedCredentials), ":", 2)
			if len(parts) != 2 {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			username := parts[0]
			password := parts[1]
			if !validateADCredentials(adServer, adBaseDN, adUserDN, adPassword, username, password) {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BasicAuthMiddleware ist eine Middleware, die Basic-Authentifizierung erfordert.
func BasicAuthMiddleware(credFilePath string) func(http.Handler) http.Handler {
	credentials, err := loadCredentials(credFilePath)
	if err != nil {
		fmt.Println("Error loading credentials:", err)
		return nil
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			z0, z1 := validateBasicAuth(authHeader, credentials)
			if !z0 {
				w.Header().Set("WWW-Authenticate", `Basic realm="Please provide credentials"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			} else {
				sessionID := r.Context().Value(mw_session.SessionKey).(string)
				authMethod := mw_session.SessionItem{Key: "authmethod", Value: "Basic"}
				mw_session.AddOrUpdateSessionItem(sessionID, authMethod)
				authUser := mw_session.SessionItem{Key: "authuser", Value: z1}
				mw_session.AddOrUpdateSessionItem(sessionID, authUser)
			}
			next.ServeHTTP(w, r)
		})
	}
}

// BearerAuthMiddleware ist eine Middleware, die Bearer-Authentifizierung erfordert.
func BearerAuthMiddleware(tokenFilePath string) func(http.Handler) http.Handler {
	tokens, err := loadBearerTokens(tokenFilePath)
	if err != nil {
		fmt.Println("Error loading tokens:", err)
		return nil
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if !validateBearerToken(authHeader, tokens) {
				w.Header().Set("WWW-Authenticate", `Bearer realm="Please provide a valid token"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
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
func validateAuth(authHeader string, basicCredentials []Credential, bearerTokens []BearerToken, adServer, adBaseDN, adUserDN, adPassword string) (bool, string) {
	if authHeader == "" {
		return false, ""
	}
	if strings.HasPrefix(authHeader, "Basic ") {
		return validateBasicAuth(authHeader, basicCredentials)
	} else if strings.HasPrefix(authHeader, "Bearer ") {
		return validateBearerToken(authHeader, bearerTokens), ""
	} else if strings.HasPrefix(authHeader, "AD ") {
		return validateADAuth(authHeader, adServer, adBaseDN, adUserDN, adPassword), ""
	}
	return false, ""
}

// validateBasicAuth überprüft, ob der Authorization-Header gültige Basic-Authentifizierungsdaten enthält
func validateBasicAuth(authHeader string, credentials []Credential) (bool, string) {
	encodedCredentials := strings.TrimPrefix(authHeader, "Basic ")
	decodedCredentials, err := base64.StdEncoding.DecodeString(encodedCredentials)
	if err != nil {
		return false, ""
	}
	parts := strings.SplitN(string(decodedCredentials), ":", 2)
	if len(parts) != 2 {
		return false, ""
	}
	username := parts[0]
	password := parts[1]
	for _, cred := range credentials {
		if username == cred.Username && password == cred.Password {
			return true, cred.Username
		}
	}
	return false, ""
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
