package main

import (
	"log/slog"
	"net/http"
	"strconv"
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
func logging(next http.Handler, log *slog.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// I need a wrapper to capture the response (e.g. to extract the status code)
		// StatusOK is the default status code sent by the server
		wrapped := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		log.Info("request", "status", strconv.Itoa(wrapped.statusCode), "method", r.Method, "path", r.URL.Path, "time", time.Since(start))

	})
}
