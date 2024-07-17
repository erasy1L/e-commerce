package server

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func ProxyRequestHandler(url *url.URL) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		proxy := httputil.NewSingleHostReverseProxy(url)

		fmt.Printf("Request received at %s at %s\n", r.URL, time.Now().UTC())
		// Update the headers to allow for SSL redirection
		r.URL.Host = url.Host
		r.URL.Scheme = url.Scheme
		r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
		r.Host = url.Host

		fmt.Printf("Redirecting request to %s at %s\n", r.URL, time.Now().UTC())
		proxy.ServeHTTP(w, r)
	})
}
