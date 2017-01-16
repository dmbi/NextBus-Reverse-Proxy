package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	p := new(Proxy)
	host := "webservices.nextbus.com"
	u, err := url.Parse(fmt.Sprintf("http://%v/api", host))
	if err != nil {
		log.Printf("Error parsing URL")
	}

	p.proxy = &httputil.ReverseProxy{
		Director: func(req *http.Request) {
			values := req.URL.Query()
			values.Add("command", "agencyList")
			values.Add("command", "")
			req.Host = host
			req.URL.Scheme = u.Scheme
			req.URL.Host = u.Host
			req.URL.Path = "/service/publicXMLFeed"
			req.URL.RawQuery = values.Encode()
			log.Println(req.URL.RawQuery)
		},
	}

	http.Handle("/", p)
	log.Fatal(http.ListenAndServe("localhost:8080", nil))
}

type Proxy struct {
	proxy *httputil.ReverseProxy
}

func (p *Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	p.proxy.ServeHTTP(w, r)
}
