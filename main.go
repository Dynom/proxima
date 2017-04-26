package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"flag"

	"fmt"

	"github.com/Dynom/proxima/handlers"
	"github.com/go-kit/kit/log"
	"github.com/juju/ratelimit"
)

var (
	allowedHosts           argumentList
	allowedImaginaryParams argumentList
	imaginaryURL           string
	listenPort             int64
	bucketRate             float64
	bucketSize             int64

	Version = "dev"
	logger  = log.With(
		log.NewLogfmtLogger(os.Stderr),
		"ts", log.DefaultTimestampUTC,
		"caller", log.DefaultCaller,
	)
)

func init() {
	flag.Var(&allowedHosts, "allow-hosts", "Repeatable flag (or a comma-separated list) for hosts to allow for the URL parameter (e.g. \"d2dktr6aauwgqs.cloudfront.net\")")
	flag.Var(&allowedImaginaryParams, "allowed-params", "A comma seperated list of parameters allows to be sent upstream. If empty, everything is allowed.")

	flag.StringVar(&imaginaryURL, "imaginary-url", "http://localhost:9000", "URL to imaginary (default: http://localhost:9000)")
	flag.Int64Var(&listenPort, "listen-port", 8080, "Port to listen on")
	flag.Float64Var(&bucketRate, "bucket-rate", 20, "Rate limiter bucket fill rate (req/s)")
	flag.Int64Var(&bucketSize, "bucket-size", 500, "Rate limiter bucket size (burst capacity)")

}

func main() {
	flag.Parse()

	logger.Log(
		"msg", "Starting.",
		"version", Version,
		"allowed_hosts", allowedHosts.String(),
		"allowed_params", allowedImaginaryParams.String(),
		"imaginary_backend", imaginaryURL,
	)

	rURL, err := url.Parse(imaginaryURL)
	if err != nil {
		panic(err)
	}

	rlBucket := ratelimit.NewBucketWithRate(bucketRate, bucketSize)

	proxy := httputil.NewSingleHostReverseProxy(rURL)
	proxy.Transport = &http.Transport{
		DisableCompression:    true,
		DisableKeepAlives:     false,
		IdleConnTimeout:       5 * time.Minute,
		MaxIdleConns:          10000,
		MaxIdleConnsPerHost:   10000,
		ResponseHeaderTimeout: 10 * time.Second,
	}

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", listenPort),
		Handler:           decorateHandler(proxy, rlBucket),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	s.ListenAndServe()
}

type httpHandler func(h http.Handler) http.Handler

func decorateHandler(h http.Handler, b *ratelimit.Bucket) http.Handler {
	decorators := []httpHandler{
		handlers.NewValidateURLParameter(logger, allowedHosts),
	}

	if len(allowedImaginaryParams) > 0 {
		decorators = append(
			decorators,
			handlers.NewAllowedParams(
				logger,
				allowedImaginaryParams,
			))
	}

	// Defining early needed handlers last
	decorators = append(
		decorators,
		handlers.NewIgnoreFaviconRequests(),
		handlers.NewRateLimitHandler(b, logger),
	)

	var handler http.Handler = h
	for _, d := range decorators {
		handler = d(handler)
	}

	return handler
}
