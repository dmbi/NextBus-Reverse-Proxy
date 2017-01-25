# NextBus Reverse Proxy

NextBus Reverse Proxy is a reverse-proxy service for San Francisco’s public transportation using the NextBus XML feed.
It uses:
  - Golang
  - MongoDB
  - Redis
  - Docker
 
*Disclaimer:
My first attempt at Golang, MongoDB, Redis and Docker.
Probably not the best way to do this but the only way I know how at the moment.
Suggestions welcome!*

## Configuration
In the `config.json` file you can change the MongoDB and Redis address and port. The default value is the service name for use with docker-compose. If you are running the application without using docker-compose you should at least change both adresses to localhost.
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
  
### Examples
   - `api/routeList/sf-muni`
   - `api/routeConfig/sf-muni/E`
   - `api/predictByStopId/sf-muni/15184`
   - `api/predictByStop/sf-muni/E/5184`
   - `api/predictionsForMultiStops/sf-muni/N|6997`		
   - `api/schedule/agencyList/E`
   - `api/vehicleLocations/sf-muni/E/0`
   
  
 
