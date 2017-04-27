package handlers

import (
	"net/http"
	"net/url"

	"github.com/go-kit/kit/log"
)

func NewValidateURLParameter(l log.Logger, allowedHosts []string) func(h http.Handler) http.Handler {
	var hosts = make(map[string]bool, len(allowedHosts))
	for _, host := range allowedHosts {
		hosts[host] = true
	}

	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			queryURL := r.URL.Query().Get("url")
			if queryURL != "" && !isValidQueryURL(queryURL, hosts) {
				l.Log("error", "domain not registered", "QS", r.URL.RawQuery, "URL", queryURL)
				http.Error(w, "Unregisterd domain", http.StatusNotAcceptable)
				return
			}

			h.ServeHTTP(w, r)
		})
	}
}

func isValidQueryURL(i string, hosts map[string]bool) bool {
	if i == "" || len(i) > 2048 {
		return false
	}

	qURL, err := url.Parse(i)
	if err != nil {
		return false
	}

	_, exists := hosts[qURL.Host]
	return exists

}
