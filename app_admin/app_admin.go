package app_admin

import (
	"httpr2/mw_logging"
	"httpr2/mw_session"
	"httpr2/mw_template"
	"net/http"
)

func Main() *http.ServeMux {
	appRouter := http.NewServeMux()
	appRouter.HandleFunc("/sessions", adminSessionHandler)
	appRouter.HandleFunc("/logs", adminLogHandler)
	return appRouter
}

func adminSessionHandler(w http.ResponseWriter, r *http.Request) {
	sessionID := r.Context().Value(mw_session.SessionKey).(string)
	item := mw_session.SessionItem{Key: "exampleKey", Value: "exampleValue"}
	mw_session.AddOrUpdateSessionItem(sessionID, item)
	mw_template.ProcessTemplate(w, "adminSessions.html", "./html-templates", 200, mw_session.SessionStore)
}

func adminLogHandler(w http.ResponseWriter, _ *http.Request) {
	mw_template.ProcessTemplate(w, "adminLogs.html", "./html-templates", 200, mw_logging.LogBook)
}
