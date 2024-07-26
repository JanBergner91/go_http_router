package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"httpr2/app_demo"
	"httpr2/app_filebrowser"
	"httpr2/apps"
	"httpr2/middleware"
	"httpr2/mw_logging"
	"httpr2/mw_session"
	"httpr2/mw_template"
	"httpr2/sys_auth"
	"log"
	"net/http"
	"os"
)

const Perm string = "http.response.statuscode"

/* Configure the Default App the "/"" URL has to be redirected */
var DefaultAppPath = "/files"

/* SessionStore */

type SessionItem struct {
	Key   string
	Value string
}

var SessionStore = map[string][]SessionItem{}

func DefaultRoute(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.URL.Path)
	if r.URL.Path == "/" {
		http.Redirect(w, r, DefaultAppPath, http.StatusFound)
	}
	fileInfo, err := os.Stat(r.URL.Path)
	if os.IsNotExist(err) {
		//CreateStatusPage(Default402)
		apps.CreateAppResponse(404, "Not Found")
		return
	}
	if !fileInfo.IsDir() {
		http.ServeFile(w, r, r.URL.Path)
	} else {
		apps.Default404(w, r)
	}

}

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
	apiMiddlewareStack := middleware.CreateStack(sys_auth.BearerAuthMiddleware("auth_bearer.json"))
	portalMiddlewareStack := middleware.CreateStack(mw_template.WriteTemplate("", "", "html-templates", http.StatusBadGateway))
	adminMiddlewareStack := middleware.CreateStack(sys_auth.BasicAuthMiddleware("auth_basic_admin.json"))
	userMiddlewareStack := middleware.CreateStack()
	appDemoMiddlewareStack := middleware.CreateStack()
	appFileMiddlewareStack := middleware.CreateStack()

	/* Main-Router-Routen */
	mainRouter.Handle("/api/", apiMiddlewareStack(http.StripPrefix("/api", apiRouter)))
	mainRouter.Handle("/portal/", portalMiddlewareStack(http.StripPrefix("/portal", portalRouter)))
	mainRouter.Handle("/admin/", adminMiddlewareStack(http.StripPrefix("/admin", adminRouter)))
	mainRouter.Handle("/user/", userMiddlewareStack(http.StripPrefix("/user", userRouter)))
	mainRouter.Handle("/demo/", appDemoMiddlewareStack(http.StripPrefix("/demo", app_demo.Main())))
	mainRouter.Handle("/files/", appFileMiddlewareStack(http.StripPrefix("/files", app_filebrowser.Main())))
	/* Default-Route (Should stay here forever) */
	mainRouter.HandleFunc("/", DefaultRoute)

	/* Sub-Router-Routen */
	adminRouter.HandleFunc("/sessions", adminSessionHandler)
	apiRouter.HandleFunc("/", apps.Default405)
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

func adminSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(mw_session.SessionKey).(string)
	item := mw_session.SessionItem{Key: "exampleKey", Value: "exampleValue"}
	mw_session.AddOrUpdateSessionItem(sessionID, item)
	mw_template.ProcessTemplate(w, "adminSessions.html", "./html-templates", 200, mw_session.SessionStore)
}
