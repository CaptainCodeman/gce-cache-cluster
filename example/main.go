package main

import (
	"net/http"

	"github.com/captaincodeman/gce-cache-cluster"
)

var logger cachecluster.Logging

func init() {
	logger, _ = cachecluster.NewStackdriverLogging()
}

func main() {
	// configure and start the groupcache server
	config, _ := cachecluster.LoadConfig("")
	cache, _ := cachecluster.New(config)

	go http.ListenAndServe(cache.ListenOn(), cache)

	// run the actual server (that uses groupcache)
	storage, _ := NewGoogleStorage("my-bucket")
	sources := NewSourceCache(storage, 64)
	server := NewServer(sources)

	http.ListenAndServe(":8080", server)
}
