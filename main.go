package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"flag"

	"fmt"

	stdlog "log"

	"github.com/Dynom/proxima/handlers"
	"github.com/go-kit/kit/log"
	"github.com/juju/ratelimit"
)

var (
	allowedHosts            argumentList
	allowedImaginaryParams  argumentList
	allowedImaginaryActions argumentList
	imaginaryURL            string
	pathSegmentToStrip      string
	listenPort              int64
	bucketRate              float64
	bucketSize              int64

	Version = "dev"
)

func init() {
	flag.Var(&allowedHosts, "allow-hosts", "Repeatable flag (or a comma-separated list) for hosts to allow for the URL parameter (e.g. \"d2dktr6aauwgqs.cloudfront.net\")")
	flag.Var(&allowedImaginaryParams, "allowed-params", "A comma seperated list of parameters allows to be sent upstream. If empty, everything is allowed.")
	flag.Var(&allowedImaginaryActions, "allowed-actions", "A comma seperated list of actions allows to be sent upstream. If empty, everything is allowed.")

	flag.StringVar(&imaginaryURL, "imaginary-url", "http://localhost:9000", "URL to imaginary (default: http://localhost:9000)")
	flag.Int64Var(&listenPort, "listen-port", 8080, "Port to listen on")
	flag.Float64Var(&bucketRate, "bucket-rate", 20, "Rate limiter bucket fill rate (req/s)")
	flag.Int64Var(&bucketSize, "bucket-size", 500, "Rate limiter bucket size (burst capacity)")
	flag.StringVar(&pathSegmentToStrip, "root-path-strip", "", "A section of the (left most) path to strip (e.g.: \"/static\"). Start with a /.")
}

func main() {
	flag.Parse()

	logger := log.With(
		log.NewLogfmtLogger(os.Stderr),
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)

	logger.Log(
		"msg", "Starting.",
		"version", Version,
		"allowed_hosts", allowedHosts.String(),
		"allowed_params", allowedImaginaryParams.String(),
		"allowed_actions", allowedImaginaryActions.String(),
		"path_to_strip", pathSegmentToStrip,
		"imaginary_backend", imaginaryURL,
	)

	rURL, err := url.Parse(imaginaryURL)
	if err != nil {
		panic(err)
	}

	rlBucket := ratelimit.NewBucketWithRate(bucketRate, bucketSize)
	proxy := newProxy(logger, rURL)

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", listenPort),
		Handler:           decorateHandler(logger, proxy, rlBucket),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	s.ListenAndServe()
}

func newProxy(l log.Logger, backend *url.URL) *httputil.ReverseProxy {
	proxy := httputil.NewSingleHostReverseProxy(backend)
	proxy.ErrorLog = stdlog.New(log.NewStdlibAdapter(l), "", stdlog.LstdFlags)
	proxy.Transport = &http.Transport{
		DisableCompression:    true,
		DisableKeepAlives:     false,
		IdleConnTimeout:       5 * time.Minute,
		MaxIdleConns:          10000,
		MaxIdleConnsPerHost:   10000,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	return proxy
}

type httpHandler func(h http.Handler) http.Handler

func decorateHandler(l log.Logger, h http.Handler, b *ratelimit.Bucket) http.Handler {
	decorators := []httpHandler{
		handlers.NewValidateURLParameter(l, allowedHosts),
	}

	if len(allowedImaginaryParams) > 0 {
		decorators = append(
			decorators,
			handlers.NewAllowedParams(
				l,
				allowedImaginaryParams,
			))
	}

	if len(allowedImaginaryActions) > 0 {
		decorators = append(
			decorators,
			handlers.NewAllowedActions(
				l,
				allowedImaginaryActions,
			))
	}

	if pathSegmentToStrip != "" {
		decorators = append(
			decorators,
			handlers.NewPathStrip(
				l,
				pathSegmentToStrip,
			))
	}

	// Defining early needed handlers last
	decorators = append(
		decorators,

		// Checking on the route "health", doesn't support path-segment-stripping!
		handlers.NewHTTPStatusPaths(l, []string{"health"}, http.StatusOK),

		handlers.NewIgnoreFaviconRequests(),
		handlers.NewRateLimitHandler(l, b),
	)

	var handler http.Handler = h
	for _, d := range decorators {
		handler = d(handler)
	}

	return handler
}
