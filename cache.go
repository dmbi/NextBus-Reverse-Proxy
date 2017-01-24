package main

import (
	"net/http"

	"github.com/garyburd/redigo/redis"
)

type Agency struct {
	Tag         string
	Title       string
	ShortTitle  string
	RegionTitle string
}

//Redis test
func redisTest(res http.ResponseWriter, req *http.Request) {
	conn := pool.Get()
	defer conn.Close()

	pong, _ := redis.Bytes(conn.Do("PING"))
	res.Write(pong)
}
