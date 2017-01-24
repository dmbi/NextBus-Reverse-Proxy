package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"gopkg.in/mgo.v2/bson"
)

var queriesCounter map[string]int

type Slow struct {
	Endpoint string
	Time     string
}

type Query struct {
	Endpoint string
	Count    int
}

//Displays stats -> slow_requests and queries
func stats(w http.ResponseWriter, r *http.Request) {
	session := connect()
	defer session.Close()
	collectionS := session.DB("stats").C("slow_requests")
	collectionQ := session.DB("stats").C("queries")

	var resultsS []Slow
	err := collectionS.Find(nil).Sort("endpoint").All(&resultsS)
	if err != nil {
		panic(err)
	}

	var resultsQ []Query
	err2 := collectionQ.Find(nil).Sort("endpoint").All(&resultsQ)
	if err2 != nil {
		panic(err2)
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte("{\n"))
	w.Write([]byte(` "slow_requests":`))
	w.Write([]byte{'{'})
	for i, slow := range resultsS {
		if i > 0 {
			w.Write([]byte{','})
		}
		w.Write([]byte(fmt.Sprintf("\n  \"%s\":\"%v\"", slow.Endpoint, slow.Time)))
	}
	w.Write([]byte("\n },\n"))
	w.Write([]byte(` "queries":`))
	w.Write([]byte{'{'})
	for i, query := range resultsQ {
		if i > 0 {
			w.Write([]byte{','})
		}
		w.Write([]byte(fmt.Sprintf("\n  \"%s\":\"%v\"", query.Endpoint, query.Count)))
	}
	w.Write([]byte("\n }\n}"))
}

/* Measure how long it took for http request to finish.
Also if above threshold add the endpoint to the slowRequests map. */
func timeTrack(start time.Time, endpoint string) {
	config := LoadConfig("./config.json")
	session := connect()
	defer session.Close()
	collection := session.DB("stats").C("slow_requests")
	elapsed := time.Since(start).Seconds()
	if elapsed > config.Threshold { // Response time higher than threshold
		e := fmt.Sprintf("%.1f", elapsed) + "s"
		_, err := collection.Upsert(bson.M{"endpoint": endpoint}, &Slow{endpoint, e})
		if err != nil {
			panic(err)
		}
		log.Printf("** SLOW REQUEST - [%s] took %s **", endpoint, e)
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

		session := connect()
		defer session.Close()
		collection := session.DB("stats").C("queries")

		if len(queriesCounter) == 0 { // Initialize  queriesCounter map
			queriesCounter = make(map[string]int)
		}
		if val, exists := queriesCounter[endpoint]; exists { // If endpoint already present, increment
			val = val + 1
			queriesCounter[endpoint] = val
			_, err := collection.Upsert(bson.M{"endpoint": endpoint}, &Query{endpoint, val})
			if err != nil {
				panic(err)
			}
		} else { // Else insert endpoint
			queriesCounter[endpoint] = 1
			_, err := collection.Upsert(bson.M{"endpoint": endpoint}, &Query{endpoint, 1})
			if err != nil {
				log.Fatal(err)
			}
		}

	}
}
