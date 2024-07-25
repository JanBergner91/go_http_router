package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"httpr2/apps"
	"httpr2/middleware"
	"httpr2/mw_auth_basic"
	"httpr2/mw_auth_bearer"
	"httpr2/mw_logging"
	"httpr2/mw_session"
	"httpr2/mw_template"
	"log"
	"net/http"
)

const Perm string = "http.response.statuscode"

/* SessionStore */

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
	jsonstore, err := json.Marshal(SessionStore)
	if err != nil {
		panic("LOL")
	}
	fmt.Println(string(jsonstore))
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

func main() {
	/*  */
	server_port := flag.String("port", "8080", "Port to be used for http server")
	server_mode := flag.String("mode", "http", "Security mode the server should run (http/https)")
	server_tls_key := flag.String("key", "private.key", "Private key file")
	server_tls_cert := flag.String("cert", "public.crt", "Public cert file")
	flag.Parse()

	/* Router-Definition */
	mainRouter := http.NewServeMux()
	apiRouter := http.NewServeMux()
	portalRouter := http.NewServeMux()
	adminRouter := http.NewServeMux()
	userRouter := http.NewServeMux()

	/* Main-Router-Middleware */
	mainMiddlewareStack := middleware.CreateStack(
		mw_session.SessionMiddleware,
		mw_logging.Logging,
	)

	/* Admin-Router-Middleware */
	apiMiddlewareStack := middleware.CreateStack(mw_auth_bearer.BearerAuthMiddleware("auth_bearer.json"))
	portalMiddlewareStack := middleware.CreateStack(mw_template.WriteTemplate("", "", "html-templates", http.StatusBadGateway))
	adminMiddlewareStack := middleware.CreateStack(mw_auth_basic.BasicAuthMiddleware("auth_basic_admin.json"))
	userMiddlewareStack := middleware.CreateStack()

	/* Main-Router-Routen */
	mainRouter.HandleFunc("/", apps.Default405)
	mainRouter.Handle("/api/", apiMiddlewareStack(http.StripPrefix("/api", apiRouter)))
	mainRouter.Handle("/portal/", portalMiddlewareStack(http.StripPrefix("/portal", portalRouter)))
	mainRouter.Handle("/admin/", adminMiddlewareStack(http.StripPrefix("/admin", adminRouter)))
	mainRouter.Handle("/user/", userMiddlewareStack(http.StripPrefix("/user", userRouter)))
	/* Sub-Router-Routen */
	adminRouter.HandleFunc("/dashboard", adminDashboardHandler)
	/*  */
	server := http.Server{
		Addr:    ":" + *server_port,
		Handler: mainMiddlewareStack(mainRouter),
	}

	/*  */
	fmt.Println("Server listening on port :" + *server_port)
	if *server_mode == "http" {
		fmt.Println("Protocol is http (insecure)")
		log.Fatal(server.ListenAndServe())
	}
	if *server_mode == "https" {
		fmt.Println("Protocol is https (secure)")
		log.Fatal(server.ListenAndServeTLS(*server_tls_cert, *server_tls_key))
	}
}

func adminDashboardHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(mw_session.SessionKey).(string)
	item := SessionItem{Key: "exampleKey", Value: "exampleValue"}
	AddOrUpdateSessionItem(sessionID, item)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Admin dashboard:" + sessionID))
}

func m0(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("KO"))
}
