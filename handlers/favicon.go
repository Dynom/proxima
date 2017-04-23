package handlers

import (
	"net/http"
	"strings"
)

func NewIgnoreFaviconRequests() func(h http.Handler) http.Handler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.RequestURI, "/favicon") {
				http.NotFound(w, r)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
