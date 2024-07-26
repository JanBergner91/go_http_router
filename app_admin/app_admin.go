package app_admin

import "net/http"

func Main() *http.ServeMux {
	appRouter := http.NewServeMux()
	//appRouter.HandleFunc("/", fileHandler)
	return appRouter
}
