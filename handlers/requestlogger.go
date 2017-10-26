package handlers

import (
	"net/http"

	"time"

	"github.com/go-kit/kit/log"
)

// NewRequestLogger logs each request with start and ending times
func NewRequestLogger(l log.Logger) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			l.Log("request", "start", "path", r.URL.Path, "remote_addr", r.RemoteAddr)

			start := time.Now()
			h.ServeHTTP(w, r)
			l.Log("request", "end", "duration", time.Since(start))
		})
	}
}
