package mw_logging

import (
	"fmt"
	"httpr2/mw_session"
	"net/http"
	"time"
)

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sessionID := r.Context().Value(mw_session.SessionKey).(string)
		WriteLog("Time:" + start.String() + "|Session:" + sessionID)
		next.ServeHTTP(w, r)
	})
}

func WriteLog(logitem string) {
	fmt.Println(logitem)
}
