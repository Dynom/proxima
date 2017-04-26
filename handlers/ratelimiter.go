package handlers

import (
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/juju/ratelimit"
)

func NewRateLimitHandler(l log.Logger, b *ratelimit.Bucket) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			d := b.Take(1)
			if d > 0 {
				l.Log("msg", "Rate limiting", "delay", d)
				time.Sleep(d)
			}

			h.ServeHTTP(w, r)
		})
	}
}
