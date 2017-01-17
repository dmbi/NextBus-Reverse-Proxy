package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

var queriesCounter map[string]int

func main() {

	target := "http://webservices.nextbus.com"
	remote, err := url.Parse(target)
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	r := mux.NewRouter()
	r.HandleFunc("/api/{endpoint:.*}", tracker(handler(proxy), "api"))
	r.HandleFunc("/api/stats", handler(proxy))
	r.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = "http"
		if mux.Vars(r)["endpoint"] == "stats" {
			r.URL.Host = r.Host
			r.URL.Path = "/api/stats"
		} else {
			r.Host = "webservices.nextbus.com"
			r.URL.Host = r.Host
			r.URL.Path = "/service/publicXMLFeed"
			r.URL.RawQuery = "command=" + mux.Vars(r)["endpoint"]
		}
		p.ServeHTTP(w, r)
	}

}

func notFound(w http.ResponseWriter, r *http.Request) {
	log.Println("404")
}

func timeTrack(start time.Time, endpoint string) {
	elapsed := time.Since(start).Seconds()
	log.Printf("[%s] request took %fs", endpoint, elapsed)
	for k, v := range queriesCounter {
		log.Printf("%s:%d", k, v)
	}
}

func tracker(fn http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		endpoint := req.URL.Path
		defer timeTrack(time.Now(), endpoint)
		fn(w, req)
		if len(queriesCounter) == 0 {
			queriesCounter = make(map[string]int)
		}
		if val, exists := queriesCounter[endpoint]; exists {
			val = val + 1
			queriesCounter[endpoint] = val
		} else {
			queriesCounter[endpoint] = 1

		}
	}
}
