package handlers

import (
	"net/http"

	"strings"

	"github.com/go-kit/kit/log"
)

func NewAllowedActions(l log.Logger, allowedActions []string) func(h http.Handler) http.Handler {
	var actions = make(map[string]bool, len(allowedActions))
	for _, p := range allowedActions {
		if p == "" {
			continue
		}

		actions[p] = true
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			action := r.URL.Path[1:]
			if _, exists := actions[action]; !exists {
				l.Log("error", "action is not white-listed", "action", action, "allowed", strings.Join(allowedActions, ","))
				http.Error(w, "Unregisterd action", http.StatusNotAcceptable)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}
