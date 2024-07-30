package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"httpr2/app_admin"
	"httpr2/app_demo"
	"httpr2/app_filebrowser"
	"httpr2/apps"
	"httpr2/middleware"
	"httpr2/mw_logging"
	"httpr2/mw_session"
	"httpr2/mw_template"
	"httpr2/sys_auth"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

const Perm string = "http.response.statuscode"

/* Configure the Default App the "/"" URL has to be redirected */
var DefaultAppPath = "/files/"

/* Define default number of logs to store */
var DefaultLogNumber int = 500

func DefaultRoute(w http.ResponseWriter, r *http.Request) {
	//fmt.Println(r.URL.Path)
	if r.URL.Path == "/" {
		http.Redirect(w, r, DefaultAppPath, http.StatusMovedPermanently)
	}
	fileInfo, err := os.Stat("./" + r.URL.Path)
	if os.IsNotExist(err) {
		apps.Default404(w, r)
		return
	}
	if !fileInfo.IsDir() {
		http.ServeFile(w, r, "./"+r.URL.Path)
		return
	} else {
		apps.Default404(w, r)
		return
	}

}

func prepareExit() {
	fmt.Println("Running exit tasks...")
	// Hier können Sie Ihre Aufräumarbeiten ausführen
	SaveFile("serversessions.json", &mw_session.SessionStore)
	SaveFile("serverlog.json", mw_logging.GetLastXitems(DefaultLogNumber))
	fmt.Println("Exit completed.")
}

func main() {
	/*  */
	//app_adinfo.MakeDefault()
	/*  */
	server_port := flag.String("port", "8080", "Port to be used for http server")
	server_mode := flag.String("mode", "http", "Security mode the server should run (http/https)")
	server_tls_key := flag.String("key", "private.key", "Private key file")
	server_tls_cert := flag.String("cert", "public.crt", "Public cert file")
	flag.Parse()

	// Signal-Kanal einrichten
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Goroutine, die auf Signale wartet
	go func() {
		<-stop
		fmt.Println("Received stop signal")
		prepareExit()
		os.Exit(0)
	}()

	// Hauptprogramm
	fmt.Println("Program is running. Press Ctrl+C to exit.")

	/* Load Infrastructure */
	/* Load Server-Log */
	LoadFile("serverlog.json", &mw_logging.LogBook)
	/* Load Session-Data */
	LoadFile("serversessions.json", &mw_session.SessionStore)

	/* Router-Definition */
	mainRouter := http.NewServeMux()
	defaultRouter := http.NewServeMux()
	apiRouter := http.NewServeMux()
	portalRouter := http.NewServeMux()
	//adminRouter := http.NewServeMux()
	userRouter := http.NewServeMux()

	/* Main-Router-Middleware */
	mainMiddlewareStack := middleware.CreateStack(
		mw_session.SessionMiddleware,
	)

	/* Admin-Router-Middleware */
	defaultMiddlewareStack := middleware.CreateStack(mw_logging.Logging)
	apiMiddlewareStack := middleware.CreateStack(mw_logging.Logging, sys_auth.BearerAuthMiddleware("auth_bearer.json"))
	portalMiddlewareStack := middleware.CreateStack(mw_logging.Logging, mw_template.WriteTemplate("", "", "html-templates", http.StatusBadGateway))
	adminMiddlewareStack := middleware.CreateStack(sys_auth.BasicAuthMiddleware("auth_basic_admin.json"))
	userMiddlewareStack := middleware.CreateStack(mw_logging.Logging)
	appDemoMiddlewareStack := middleware.CreateStack(mw_logging.Logging)
	appFileMiddlewareStack := middleware.CreateStack(mw_logging.Logging)

	/* Main-Router-Routen */
	mainRouter.Handle("/api/", apiMiddlewareStack(http.StripPrefix("/api", apiRouter)))
	mainRouter.Handle("/portal/", portalMiddlewareStack(http.StripPrefix("/portal", portalRouter)))
	mainRouter.Handle("/admin/", adminMiddlewareStack(http.StripPrefix("/admin", app_admin.Main())))
	mainRouter.Handle("/user/", userMiddlewareStack(http.StripPrefix("/user", userRouter)))
	mainRouter.Handle("/demo/", appDemoMiddlewareStack(http.StripPrefix("/demo", app_demo.Main())))
	mainRouter.Handle("/files/", appFileMiddlewareStack(http.StripPrefix("/files", app_filebrowser.Main())))

	/* Default-Route (Should stay here forever) */
	mainRouter.Handle("/", defaultMiddlewareStack(http.StripPrefix("", defaultRouter)))
	defaultRouter.HandleFunc("/", DefaultRoute)

	/* Sub-Router-Routen */
	//adminRouter.HandleFunc("/sessions", app_admin.adminSessionHandler)
	apiRouter.HandleFunc("/", HttpSaveFile("serverlog.json", &mw_logging.LogBook))
	/*  */
	server := http.Server{
		Addr:    ":" + *server_port,
		Handler: mainMiddlewareStack(mainRouter),
	}

	/*  */
	fmt.Println("Server listening on port :" + *server_port)
	if *server_mode == "http" {
		fmt.Println("Protocol is http (insecure)")
		StopServer(server.ListenAndServe())
	}
	if *server_mode == "https" {
		fmt.Println("Protocol is https (secure)")
		StopServer(server.ListenAndServeTLS(*server_tls_cert, *server_tls_key))
	}

}

func StopServer(e error) {
	fmt.Println("Stopping server...")
	prepareExit()
	fmt.Println("Server stopped!")
}

func HttpSaveFile(xfile string, a any) apps.App {
	return func(next_w http.ResponseWriter, next_r *http.Request) {
		file, _ := os.Create(xfile)
		encoder := json.NewEncoder(file)
		err := encoder.Encode(a)
		if err != nil {
			fmt.Println("Error encoding JSON:", err)
		}
		file.Close()
	}
}

func SaveFile(xfile string, a any) {
	file, _ := os.Create(xfile)
	encoder := json.NewEncoder(file)
	err := encoder.Encode(a)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
	}
	file.Close()
	fmt.Println("Saved content to: ", xfile)
}

func HttpLoadFile(xfile string, a any) apps.App {
	return func(next_w http.ResponseWriter, next_r *http.Request) {
		if _, err := os.Stat(xfile); os.IsNotExist(err) {
			fmt.Println("Speicherdatei nicht gefunden, neue Map wird erstellt")
		} else {
			file, _ := os.Open(xfile)
			decoder := json.NewDecoder(file)
			err := decoder.Decode(a)
			if err != nil {
				fmt.Println("Fehler in JSON-Datei:", err)
			}
			file.Close()
		}
	}
}

func LoadFile(xfile string, a any) {
	if _, err := os.Stat(xfile); os.IsNotExist(err) {
		fmt.Println("Speicherdatei nicht gefunden, neue Map wird erstellt")
	} else {
		file, _ := os.Open(xfile)
		decoder := json.NewDecoder(file)
		err := decoder.Decode(a)
		if err != nil {
			fmt.Println("Fehler in JSON-Datei:", err)
		}
		file.Close()
	}
	fmt.Println("Loaded content from: ", xfile)
}
