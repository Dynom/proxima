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
			start := time.Now()
			l.Log("request", "start", "path", r.URL.Path, "remote_addr", r.RemoteAddr)
			defer l.Log("request", "end", "duration", time.Since(start))

			h.ServeHTTP(w, r)
		})
	}
}
