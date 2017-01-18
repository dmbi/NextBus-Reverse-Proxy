package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/gorilla/mux"
)

var queriesCounter map[string]int
var slowRequests map[string]string
var pool = newPool()

// Pool configuration
func newPool() *redis.Pool {
	return &redis.Pool{
		MaxIdle:   80,
		MaxActive: 12000,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", ":6379")
		},
	}
}

func main() {

	target := "http://webservices.nextbus.com"
	remote, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	r := mux.NewRouter()
	r.HandleFunc("/api/v1/{endpoint:.*}", counter(handler(proxy)))
	r.HandleFunc("/api/stats", counter(http.HandlerFunc(stats)))
	r.HandleFunc("/api/red", counter(http.HandlerFunc(red)))
	r.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}

//Main http handler
func handler(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = "http"
		r.Host = "webservices.nextbus.com"
		r.URL.Host = r.Host
		r.URL.Path = "/service/publicXMLFeed"
		r.URL.RawQuery = "command=" + mux.Vars(r)["endpoint"]
		p.ServeHTTP(w, r)
	}
}

// Custom 404 Page
func notFound(w http.ResponseWriter, r *http.Request) {
	log.Println("404")
}

//Redis test
func red(res http.ResponseWriter, req *http.Request) {
	conn := pool.Get()
	defer conn.Close()

	pong, _ := redis.Bytes(conn.Do("PING"))
	res.Write(pong)
}

/* Displays endpoints with slow requests and the number of requests per endpoint
Not sure about this implementation though. Might try again later */
func stats(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\n slow_requests: "))
	s, _ := json.MarshalIndent(slowRequests, " ", "  ")
	w.Write(s)
	q, _ := json.MarshalIndent(queriesCounter, " ", "  ")
	w.Write([]byte("\n queries: "))
	w.Write(q)
	w.Write([]byte("\n}"))
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
func counter(fn http.HandlerFunc) http.HandlerFunc {
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
