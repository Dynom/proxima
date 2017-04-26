package handlers

import (
	"net/http"

	"github.com/go-kit/kit/log"
)

func NewHTTPStatusPaths(_ log.Logger, paths []string, httpStatus int) func(h http.Handler) http.Handler {
	var actions = make(map[string]bool, len(paths))
	for _, p := range paths {
		if p == "" {
			continue
		}

		actions[p] = true
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			action := r.URL.Path[1:]
			if _, exists := actions[action]; action != "" && exists {
				w.WriteHeader(httpStatus)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
