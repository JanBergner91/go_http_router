package main

import (
	"flag"
	"fmt"
	"httprouter/middleware"
	middlewareauthbasic "httprouter/middleware_auth_basic"
	middlewarelogging "httprouter/middleware_logging"
	"log"
	"net/http"
)

func m0(w http.ResponseWriter, r *http.Request) {
	t := r.PathValue("id")
	fmt.Println(t)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(t))
}

func main() {

	server_port := flag.String("port", "8080", "Port to be used for http server")
	server_mode := flag.String("mode", "http", "Security mode the server should run (http/https)")
	server_tls_key := flag.String("key", "private.key", "Private key file")
	server_tls_cert := flag.String("cert", "public.crt", "Public cert file")
	flag.Parse()

	router := http.NewServeMux()
	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("/")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
		ctx := r.Context().Value(middlewareauthbasic.AuthUserID).(string)

		fmt.Println("H:" + ctx)
	})
	router.HandleFunc("GET /{id}", m0)

	middlewares := middleware.CreateStack(
		middlewarelogging.Logging,
		middlewareauthbasic.IsAuthenticated,
	)

	server := http.Server{
		Addr:    ":" + *server_port,
		Handler: middlewares(router),
	}
	fmt.Println("Server listening on port :" + *server_port)
	if *server_mode == "http" {
		server.ListenAndServe()
	}
	if *server_mode == "https" {
		log.Fatal(server.ListenAndServeTLS(*server_tls_cert, *server_tls_key))
	}

}
