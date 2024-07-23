package mw_wrapper

import "net/http"

type WrappedWriter struct {
	http.ResponseWriter
	StatusCode int
}

func (w *WrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader((statusCode))
	w.StatusCode = statusCode
}

func (w *WrappedWriter) GetStatusCode() int {
	return w.StatusCode
}
