package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	p := new(Proxy)
	host := "webservices.nextbus.com/service/publicXMLFeed?command="
	u, err := url.Parse(fmt.Sprintf("http://%v/", host))
	if err != nil {
		log.Printf("Error parsing URL")
	}

	targetQuery := u.RawQuery
	p.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			req.Host = host
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = singleJoiningSlash(u.Path, req.URL.Path)
			if targetQuery == "" || req.URL.RawQuery == "" {
				req.URL.RawQuery = targetQuery + req.URL.RawQuery
			} else {
				req.URL.RawQuery = targetQuery + "&" + req.URL.RawQuery
			}
		},
	}

	http.Handle("/", p)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

type Proxy struct {
	proxy *httputil.ReverseProxy
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
