package main

import (
	"log"
	"net/http"
	"time"
)

// This needs to implement the http.ResponseWriter interface
type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(statusCode int) {
	w.ResponseWriter.WriteHeader(statusCode)
	w.statusCode = statusCode
}

// Logging is a middleware that logs a request
func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// I need a wrapper to capture the response (e.g. to extract the status code)
		// StatusOK is the default status code
		wrapped := &wrappedWriter{w, http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Println(wrapped.statusCode, r.Method, r.URL.Path, time.Since(start))

	})
}
