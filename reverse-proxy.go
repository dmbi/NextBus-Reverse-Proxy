package main

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func report(r *http.Request) {
	r.Host = "nextbus.com"
	r.URL.Host = r.Host
	r.URL.Scheme = "http"
}

func main() {
	proxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Scheme: "http",
		Host:   "nextbus.com",
	})
	proxy.Director = report
	http.Handle("/", proxy)
	fmt.Println("Server running...")
	http.ListenAndServe(":8080", nil)
}
