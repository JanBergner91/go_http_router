package app_demo

import (
	"httpr2/apps"
	"net/http"
)

func Main() *http.ServeMux {
	appRouter := http.NewServeMux()
	appRouter.HandleFunc("/", WriteBasicPage)
	return appRouter
}

var WriteBasicPage0 = apps.CreateAppResponse(http.StatusConflict, "OK")

func WriteBasicPage(w http.ResponseWriter, r *http.Request) {
	//apps.CreateAppResponse()
	w.WriteHeader(200)
	w.Write([]byte("Hello World!"))
}
