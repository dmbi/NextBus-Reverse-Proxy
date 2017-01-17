package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/gorilla/mux"
)

var queriesCounter map[string]int
var slowRequests map[string]string

func main() {

	target := "http://webservices.nextbus.com"
	remote, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	r := mux.NewRouter()
	r.HandleFunc("/api/{endpoint:.*}", counter(handler(proxy), "api"))
	r.HandleFunc("/api/stats", handler(proxy))
	r.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}

func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = "http"
		if mux.Vars(r)["endpoint"] == "stats" { /// TODO: Not working as intended, needs work
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

// Custom 404 Page
func notFound(w http.ResponseWriter, r *http.Request) {
	log.Println("404")
}

/* Measure how long it took for http request to finish.
Also if above threshold add the endpoint to the slowRequests map. */
func timeTrack(start time.Time, endpoint string) {
	elapsed := time.Since(start).Seconds()
	threshold := 0.5         // 0.5 seconds
	if elapsed > threshold { // Response time higher than threshold
		if len(slowRequests) == 0 { // Initialize slowRequests map
			slowRequests = make(map[string]string)
		}
		e := fmt.Sprintf("%.3f", elapsed) // Float to string, round value
		slowRequests[endpoint] = e
		log.Printf("** SLOW REQUEST - [%s] took %ss **", endpoint, e)
	}
}

/* Counts the number of times that the endpoint has been requested and also
passes the http request to the timeTrack function so we can measure how
long it took for it to finish */
func counter(fn http.HandlerFunc, name string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		endpoint := req.URL.Path
		defer timeTrack(time.Now(), endpoint) // Begins tracking
		fn(w, req)
		if len(queriesCounter) == 0 { // Initialize  queriesCounter map
			queriesCounter = make(map[string]int)
		}
		if val, exists := queriesCounter[endpoint]; exists { // If endpoint already present, increment
			val = val + 1
			queriesCounter[endpoint] = val
		} else { // Else insert endpoint
			queriesCounter[endpoint] = 1
		}
	}
}
