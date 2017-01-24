package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"gopkg.in/mgo.v2" //MongoDB

	"github.com/garyburd/redigo/redis" //Caching
	"github.com/gorilla/mux"           //HTTP Routing
)

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

func connect() (session *mgo.Session) {
	connectURL := "localhost"
	session, err := mgo.Dial(connectURL)
	if err != nil {
		fmt.Printf("Can't connect to mongo, go error %v\n", err)
		os.Exit(1)
	}
	session.SetSafe(&mgo.Safe{})
	return session
}

func main() {

	session := connect()
	defer session.Close()
	session.DB("stats").DropDatabase() // Clear previous stats

	target := "http://webservices.nextbus.com"
	remote, err := url.Parse(target)
	if err != nil {
		panic(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(remote)
	r := mux.NewRouter()
	r.HandleFunc("/api/agencyList", counter(handler(proxy, "agencyList")))
	r.HandleFunc("/api/routeList/{a}", counter(handler(proxy, "routeList")))
	r.HandleFunc("/api/routeConfig/{a}/{r}", counter(handler(proxy, "routeConfig")))
	r.HandleFunc("/api/predictByStopId/{a}/{stopId:.*}", counter(handler(proxy, "predictByStopId")))
	r.HandleFunc("/api/predictByStop/{a}/{r}/{s}", counter(handler(proxy, "predictByStop")))
	r.HandleFunc("/api/predictionsForMultiStops/{a}/{s:.*}", counter(handler(proxy, "predictionsForMultiStops")))
	r.HandleFunc("/api/schedule/{a}/{r}", counter(handler(proxy, "schedule")))
	r.HandleFunc("/api/messages/{a}/{r:.*}", counter(handler(proxy, "messages")))
	r.HandleFunc("/api/vehicleLocations/{a}/{r}/{t}", counter(handler(proxy, "vehicleLocations")))
	r.HandleFunc("/api/stats", counter(http.HandlerFunc(stats)))
	r.HandleFunc("/api/red", counter(http.HandlerFunc(red)))
	r.NotFoundHandler = http.HandlerFunc(notFound)
	http.Handle("/", r)
	http.ListenAndServe(":8080", r)
}

//Main http handler
func handler(p *httputil.ReverseProxy, endpoint string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.URL.Scheme = "http"
		r.Host = "webservices.nextbus.com"
		r.URL.Host = r.Host
		r.URL.Path = "/service/publicXMLFeed"
		r.URL.RawQuery = "command=" + endpoint
		switch endpoint {
		case "routeList":
			r.URL.RawQuery = r.URL.RawQuery + "&a=" + mux.Vars(r)["a"]
		case "routeConfig":
			r.URL.RawQuery = r.URL.RawQuery + "&a=" + mux.Vars(r)["a"] + "&r=" + mux.Vars(r)["r"]
		case "predictByStopId":
			r.URL.RawQuery = "command=predictions" + "&a=" + mux.Vars(r)["a"] + "&stopId=" + mux.Vars(r)["stopId"]
			if strings.Contains(mux.Vars(r)["stopId"], "/") {
				split := strings.Split(mux.Vars(r)["stopId"], "/")
				mux.Vars(r)["stopId"] = split[0]
				r.URL.RawQuery = "command=predictions" + "&a=" + mux.Vars(r)["a"] + "&stopId=" + mux.Vars(r)["stopId"] + "&r=" + split[1]
			}
		case "predictByStop":
			r.URL.RawQuery = "command=predictions" + "&a=" + mux.Vars(r)["a"] + "&r=" + mux.Vars(r)["r"] + "&s=" + mux.Vars(r)["s"]
		case "predictionsForMultiStops":
			r.URL.RawQuery = r.URL.RawQuery + "&a=" + mux.Vars(r)["a"]
			split := strings.Split(mux.Vars(r)["s"], "/")
			for i := range split {
				r.URL.RawQuery = r.URL.RawQuery + "&stops=" + split[i]
			}
		case "schedule":
			r.URL.RawQuery = r.URL.RawQuery + "&a=" + mux.Vars(r)["a"] + "&r=" + mux.Vars(r)["r"]
		case "messages":
			r.URL.RawQuery = r.URL.RawQuery + "&a=" + mux.Vars(r)["a"]
			split := strings.Split(mux.Vars(r)["r"], "/")
			for i := range split {
				r.URL.RawQuery = r.URL.RawQuery + "&r=" + split[i]
			}
		case "vehicleLocations":
			r.URL.RawQuery = r.URL.RawQuery + "&a=" + mux.Vars(r)["a"] + "&r=" + mux.Vars(r)["r"] + "&t=" + mux.Vars(r)["t"]
		}
		p.ServeHTTP(w, r)
	}
}

// Custom 404 Page
func notFound(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Incorrect usage, please read the documentation."))
	log.Println("404")
}

//Redis test
func red(res http.ResponseWriter, req *http.Request) {
	conn := pool.Get()
	defer conn.Close()

	pong, _ := redis.Bytes(conn.Do("PING"))
	res.Write(pong)
}
