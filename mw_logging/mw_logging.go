package mw_logging

import (
	"fmt"
	"httpr2/mw_session"
	"net/http"
	"time"
)

type LogItem struct {
	URL      string
	Time     string
	Session  string
	RemoteIP string
}

var LogBook []LogItem

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sessionID := r.Context().Value(mw_session.SessionKey).(string)
		li := LogItem{URL: r.URL.Path, Time: start.Format(time.DateTime), Session: sessionID, RemoteIP: r.RemoteAddr}
		WriteLog(li)
		next.ServeHTTP(w, r)
	})
}

func WriteLog(logitem LogItem) {
	LogBook = append(LogBook, logitem)
	fmt.Println(logitem)
}
