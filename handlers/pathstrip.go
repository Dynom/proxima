package handlers

import (
	"net/http"

	"strings"

	"github.com/go-kit/kit/log"
)

// NewPathStrip strips the path from the request URL, paths always start with a /.
func NewPathStrip(_ log.Logger, path string) func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, path) {
				r.URL.Path = strings.TrimPrefix(r.URL.Path, path)
			}

			h.ServeHTTP(w, r)
		})
	}
}
