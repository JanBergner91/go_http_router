package mw_auth_activedirectory

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-ldap/ldap/v3"
)

// ADAuthMiddleware ist eine Middleware, die Active Directory-Authentifizierung durchführt.
func ADAuthMiddleware(adServer, adBaseDN, adUserDN, adPassword string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" || !strings.HasPrefix(authHeader, "Basic ") {
				// Wenn der Authorization-Header fehlt oder nicht im Basic-Format vorliegt, gib 401 zurück
				w.Header().Set("WWW-Authenticate", `Basic realm="Please provide credentials"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Extrahiere die Anmeldeinformationen aus dem Header
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
