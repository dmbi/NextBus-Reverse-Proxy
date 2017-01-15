package main

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "nextbus.com",
	})
	http.Handle("/", proxy)
	http.ListenAndServe(":8082", nil)
}
