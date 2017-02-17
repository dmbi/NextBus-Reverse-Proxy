# NextBus Reverse Proxy

NextBus Reverse Proxy is a reverse-proxy service for San Francisco’s public transportation using the NextBus XML feed.
It uses:
  - Golang
  - MongoDB
  - Redis
  - Docker (<a href="https://hub.docker.com/r/dmbi17/nextbus-reverse-proxy/">Repository</a>)
 
*Disclaimer:
My first attempt at Golang, MongoDB, Redis and Docker.
Probably not the best way to do this but the only way I know how at the moment.
Suggestions welcome!*

Reading more about the <a href="http://www.nextbus.com/xmlFeedDocs/NextBusXMLFeed.pdf">NextBus XML Feed</a> is recommended.

## Running the application - Locally
If you are going to run the application you need to:
 - <a href="https://golang.org/doc/install">Install Golang</a>
 - <a href="https://golang.org/doc/install#testing">Setup your $GOPATH</a>
 - <a href="https://redis.io/download"> Install Redis</a>
 - <a href="https://docs.mongodb.com/manual/installation/">Install MongoDB</a>
 
Download the source code and it's dependencies
 ```bash
#The application
$ go get github.com/dmbi/NextBus-Reverse-Proxy

#Dependencies
$ go get gopkg.in/mgo.v2
$ go get gopkg.in/mgo.v2/bson
$ go get github.com/gorilla/mux
$ go get github.com/garyburd/redigo/redis

#Run it
$ cd $GOPATH/src/github.com/dmbi/NextBus-Reverse-Proxy
$ ./Run.sh
 ```
## Running the application - Distributed mode
By using <a href="https://docs.docker.com/compose/overview/">docker-compose</a> we can use a single command to run all the required services at the same time.
There is no need to install Redis and MongoDB separately.

 - Start by <a href="https://docs.docker.com/compose/install/">installing docker-compose</a>.
 - Install Golang and setup your $GOPATH as explained above.
 - Get the application (If you haven't before) and run docker-compose
```bash
$ go get github.com/dmbi/NextBus-Reverse-Proxy
$ cd $GOPATH/src/github.com/dmbi/NextBus-Reverse-Proxy
$ docker-compose build
$ docker-compose up
```
## Configuration
In the `config.json` file you can change the MongoDB and Redis address and port. The default value is the service name for use with docker-compose. If you are running the application without using docker-compose you should at least change both addresses to localhost.
In this file you can also change the threshold for the slow_requests stats. The default value is 1.0 (1 second).

## API Endpoints
The application address is, by default, `127.0.0.1:8080/`. All of these, except the `stats` endpoint, will return the original XML response. 
These are the allowed endpoints:
  - `api/stats` Exposes statistics: `slow_requests` - Lists the endpoints which had response time higher a certain threshold along with the time taken. `queries` - Lists all the endpoints queried by the user along with the number of requests for each. Response is in JSON.
  - `api/agencyList` Lists all agencies.
  - `api/routeList/{agency}` Lists all the routes for the agency tag supplied.
  - `api/routeConfig/{agency}/{route}` Lists all the stops for the route tag supplied.
  - `api/predictByStopId/{agency}/{stopId}` Lists arrival/departure predictions for a stop. A `/{route}` tag  can be appended if predictions for only one route are desired. Append `&useShortTitles=true` to have the XML feed return short titles intended for display devices with small screens.
  - `api/predictByStop/{agency}/{route}/{stop}` Same as predictByStopId but using the `{stop}` tag instead. `{route}` tag is required because `{stop}` tag is associated with a route.  Append `&useShortTitles=true` to have the XML feed return short titles intended for display devices with small screens.
  - `api/predictionsForMultiStops/{agency}/{stops}` Lists arrival/departure predictions for multi-stops. The format of the `{stops}` tag is *route|stop* . Append more `{/stops}` for more stops.
  - `api/schedule/{agency}/{route}` Obtain the schedule information for a given `{agency}` and `{route}` tags
  - `api/messages/{agency}/{route}` List the active messages for the selected route. Append `{/route}`for more routes.
  - `api/vehicleLocations/{agency}/{route}/{time}` Lists vehicle locations for the selected `{route}`. `{time}` tag is in msec since the 1970 epoch time. If you specify a time of 0, then data for the last 15 minutes is provided.
  
Get `{agency}` tags using `agencyList`, `{route}` tags using `routeList`, `{stop}` and `{stopId}` tags using `routeConfig`.
  
### Examples
   - `api/routeList/sf-muni`
   - `api/routeConfig/sf-muni/E`
   - `api/predictByStopId/sf-muni/15184`
   - `api/predictByStop/sf-muni/E/5184`
   - `api/predictionsForMultiStops/sf-muni/N|6997`		
   - `api/schedule/sf-muni/E`
   - `api/vehicleLocations/sf-muni/E/0`
   
## MongoDB and Statistics
The statistics are written in a MongoDB database. The main goal here is that all instances of the proxy have a common data store so that each instance displays the same statistics.

`slow_requests` lists the endpoints which had response time higher a certain threshold along with the time taken. Every request that the reverse proxy handles is timed and if the it takes long than a certain user-defined threshold it is written on the slow_requests database collection.

`queries` lists all the endpoints queried by the user along with the number of requests for each. Endpoints are added to a map and a counter is incremented each time the endpoint is queried. Everything is then written on the queries database collection.

## Caching
Cache is provided by Redis. 
Currently only the connection to the Redis server is implemented, so no actual caching is taking place.
It's on the TODO list. 
