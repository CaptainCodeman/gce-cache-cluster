package main

import (
	// "bytes"
	"fmt"
	"time"

	"net/http"

	"code.cloudfoundry.org/bytefmt"
	"github.com/golang/groupcache"
	"github.com/wblakecaldwell/profiler"
)

type (
	server struct {
		sources *groupcache.Group
	}
)

func NewServer(sources *groupcache.Group) *http.Server {
	s := &server{
		sources: sources,
	}

	m := http.NewServeMux()

	m.HandleFunc("/", s.indexHandler)
	m.HandleFunc("/img/", s.imageHandler)

	// expose stats about the groupcache on this instance
	m.HandleFunc("/profiler/info.html", profiler.MemStatsHTMLHandler)
	m.HandleFunc("/profiler/info", profiler.ProfilingInfoJSONHandler)
	m.HandleFunc("/profiler/start", profiler.StartProfilingHandler)
	m.HandleFunc("/profiler/stop", profiler.StopProfilingHandler)

	profiler.RegisterExtraServiceInfoRetriever(s.extraServiceInfo)

	srv := http.Server{
		Addr:         ":8080",
		Handler:      m,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  time.Second * 620,
	}

	return &srv
}

func (s *server) indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "test server")
}

func (s *server) imageHandler(w http.ResponseWriter, r *http.Request) {
	key := r.URL.Path[5:]

	var bv groupcache.ByteView
	if err := s.sources.Get(nil, key, groupcache.ByteViewSink(&bv)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	reader := bv.Reader()

	/*
			var data []byte
			if err := s.sources.Get(nil, key, groupcache.AllocatingByteSliceSink(&data)); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

		  reader := bytes.NewReader(data)
	*/

	http.ServeContent(w, r, r.URL.Path, time.Now().UTC(), reader)
}

func (s *server) extraServiceInfo() map[string]interface{} {
	extraInfo := make(map[string]interface{})

	name := "source"
	cache := s.sources.CacheStats(groupcache.MainCache)
	extraInfo[name+".CacheStats.Bytes"] = bytefmt.ByteSize(uint64(cache.Bytes))
	extraInfo[name+".CacheStats.Evictions"] = cache.Evictions
	extraInfo[name+".CacheStats.Gets"] = cache.Gets
	extraInfo[name+".CacheStats.Hits"] = cache.Hits
	extraInfo[name+".CacheStats.Items"] = cache.Items

	stats := s.sources.Stats
	extraInfo[name+".Stats.CacheHits"] = stats.CacheHits
	extraInfo[name+".Stats.Gets"] = stats.Gets
	extraInfo[name+".Stats.Loads"] = stats.Loads
	extraInfo[name+".Stats.LoadsDeduped"] = stats.LoadsDeduped
	extraInfo[name+".Stats.LocalLoadErrs"] = stats.LocalLoadErrs
	extraInfo[name+".Stats.LocalLoads"] = stats.LocalLoads
	extraInfo[name+".Stats.PeerErrors"] = stats.PeerErrors
	extraInfo[name+".Stats.PeerLoads"] = stats.PeerLoads
	extraInfo[name+".Stats.ServerRequests"] = stats.ServerRequests

	return extraInfo
}
