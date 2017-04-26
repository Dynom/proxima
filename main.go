package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"time"

	"strings"

	"flag"

	"fmt"

	stdlog "log"

	"github.com/Dynom/proxima/handlers"
	"github.com/go-kit/kit/log"
	"github.com/juju/ratelimit"
)

var (
	allowedHosts            argumentList
	allowedImaginaryParams  string
	allowedImaginaryActions string
	imaginaryURL            string
	pathSegmentToStrip      string
	listenPort              int64
	bucketRate              float64
	bucketSize              int64

	Version = "dev"
)

type argumentList []string

func (l argumentList) String() string {
	return strings.Join(l, ",")
}

func (l *argumentList) Set(value string) error {
	*l = append(*l, value)
	return nil
}

func init() {
	flag.Var(&allowedHosts, "allow-host", "Repeatable flag for hosts to allow for the URL parameter (e.g. \"d2dktr6aauwgqs.cloudfront.net\")")
	flag.StringVar(&imaginaryURL, "imaginary-url", "http://localhost:9000", "URL to imaginary (default: http://localhost:9000)")
	flag.Int64Var(&listenPort, "listen-port", 8080, "Port to listen on")
	flag.Float64Var(&bucketRate, "bucket-rate", 20, "Rate limiter bucket fill rate (req/s)")
	flag.Int64Var(&bucketSize, "bucket-size", 500, "Rate limiter bucket size (burst capacity)")
	flag.StringVar(&allowedImaginaryParams, "allowed-params", "", "A comma seperated list of parameters allows to be sent upstream. If empty, everything is allowed.")
	flag.StringVar(&allowedImaginaryActions, "allowed-actions", "", "A comma seperated list of actions allows to be sent upstream. If empty, everything is allowed.")
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
		"imaginary_backend", imaginaryURL,
	)

	rURL, err := url.Parse(imaginaryURL)
	if err != nil {
		panic(err)
	}

	rlBucket := ratelimit.NewBucketWithRate(bucketRate, bucketSize)
	proxy := newProxy(rURL, logger)

	s := &http.Server{
		Addr:              fmt.Sprintf(":%d", listenPort),
		Handler:           decorateHandler(proxy, rlBucket, logger),
		ReadHeaderTimeout: 2 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
		IdleTimeout:       60 * time.Second,
		MaxHeaderBytes:    1 << 20,
	}

	s.ListenAndServe()
}

func newProxy(backend *url.URL, l log.Logger) *httputil.ReverseProxy {
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

func decorateHandler(h http.Handler, b *ratelimit.Bucket, l log.Logger) http.Handler {
	decorators := []httpHandler{
		handlers.NewValidateURLParameter(l, allowedHosts),
	}

	if allowedImaginaryParams != "" {
		decorators = append(
			decorators,
			handlers.NewAllowedParams(
				l,
				strings.Split(allowedImaginaryParams, ","),
			))
	}

	if allowedImaginaryActions != "" {
		decorators = append(
			decorators,
			handlers.NewAllowedActions(
				l,
				strings.Split(allowedImaginaryActions, ","),
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
		handlers.NewIgnoreFaviconRequests(),
		handlers.NewRateLimitHandler(b, l),
	)

	var handler http.Handler = h
	for _, d := range decorators {
		handler = d(handler)
	}

	return handler
}
