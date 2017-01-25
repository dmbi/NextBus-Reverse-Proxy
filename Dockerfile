FROM golang:latest

# Install dependencies
RUN go get gopkg.in/mgo.v2
RUN go get gopkg.in/mgo.v2/bson
RUN go get github.com/gorilla/mux
RUN go get github.com/garyburd/redigo/redis

# Copy our sources
ADD . /go/src/github.com/dmbi/NextBus-Reverse-Proxy

# Set workdir
WORKDIR /go/src/github.com/dmbi/NextBus-Reverse-Proxy

# Install api binary globally within container
RUN go install github.com/dmbi/NextBus-Reverse-Proxy

# Set binary as entrypoint
ENTRYPOINT /go/bin/NextBus-Reverse-Proxy

# Expose default port (8080)
EXPOSE 8080
