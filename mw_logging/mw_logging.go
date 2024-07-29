package mw_logging

import (
	"fmt"
	"httpr2/mw_session"
	"net/http"
	"strings"
	"time"
)

type LogItem struct {
	Module     string
	URL        string
	Time       string
	Session    string
	RemoteIP   string
	RemotePort string
	Method     string
	Status     string
}

var LogBook []LogItem

func GetLastXitems(amount int) []LogItem {
	var temp []LogItem
	startIndex := len(LogBook) - amount
	// Sicherstellen, dass startIndex nicht negativ wird
	if startIndex < 0 {
		startIndex = 0
	}
	for i := startIndex; i < len(LogBook); i++ {
		temp = append(temp, LogBook[i])
	}
	return temp
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		AppendToLogMainFunction(w, r, "")
		next.ServeHTTP(w, r)
	})
}

func AppendToLogMainFunction(w http.ResponseWriter, r *http.Request, status string) {
	start := time.Now()
	sessionID := r.Context().Value(mw_session.SessionKey).(string)
	meth := r.Method
	IP_ONLY := strings.Split(r.RemoteAddr, ":")[0]
	PORT_ONLY := strings.Split(r.RemoteAddr, ":")[1]
	li := LogItem{Module: "http", URL: r.URL.Path, Time: start.Format(time.DateTime), Session: sessionID, RemoteIP: IP_ONLY, RemotePort: PORT_ONLY, Status: status, Method: meth}
	WriteLog(li)
}

func WriteLog(logitem LogItem) {
	LogBook = append(LogBook, logitem)
	fmt.Println(logitem)
}
