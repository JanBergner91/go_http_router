package app_demo

import (
	"httpr2/apps"
	"net/http"
)

func Main() *http.ServeMux {
	appRouter := http.NewServeMux()
	appRouter.HandleFunc("/", WriteBasicPage)
	appRouter.HandleFunc("/d", apps.Default405)
	return appRouter
}

var WriteBasicPage0 = apps.CreateAppResponse(http.StatusConflict, "OK")

func WriteBasicPage(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
	w.Write([]byte("Hello World!"))
}
