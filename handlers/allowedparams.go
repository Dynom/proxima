package handlers

import (
	"net/http"

	"strings"

	"github.com/go-kit/kit/log"
)

func NewAllowedParams(l log.Logger, allowedParams []string) func(h http.Handler) http.Handler {
	var params = make(map[string]bool, len(allowedParams))
	for _, p := range allowedParams {
		if p == "" {
			continue
		}

		params[p] = true
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestParams := r.URL.Query()

			for p := range requestParams {
				if _, exists := params[p]; !exists {
					l.Log("error", "parameter is not white-listed", "parameter", p, "allowed", strings.Join(allowedParams, ","))
					http.Error(w, "Unregisterd parameter", http.StatusNotAcceptable)
					return
				}
			}

			h.ServeHTTP(w, r)
		})
	}
}
