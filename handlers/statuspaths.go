package handlers

import (
	"net/http"

	"github.com/go-kit/kit/log"
)

func NewHTTPStatusPaths(_ log.Logger, paths []string, httpStatus int) func(h http.Handler) http.Handler {
	var pathMap = make(map[string]bool, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}

		pathMap[p] = true
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, exists := pathMap[r.URL.Path]; exists {
				w.WriteHeader(httpStatus)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
