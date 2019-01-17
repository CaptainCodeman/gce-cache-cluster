package cachecluster

import (
	"fmt"
	"net"
	"sort"
	"strings"
	"time"

	"runtime/debug"

	"github.com/golang/groupcache"
)

type (
	cache struct {
		*groupcache.HTTPPool
		me     net.IP
		prev   string
		config *CacheConfig
	}
)

// New creates a new auto-clusered cache instance
func New(config *Config) (*cache, error) {
	cluster, err := newCluster(config)
	if err != nil {
		logger.Debugf("error creating cluster %v", err)
		return nil, err
	}

	self := fmt.Sprintf("http://%s:%d", cluster.me, config.Cache.Port)
	logger.Debugf("groupcache on %s", self)

	pool := groupcache.NewHTTPPoolOpts(self, nil)
	cache := &cache{
		HTTPPool: pool,
		me:       cluster.me,
		config:   &config.Cache,
	}

	if config.Cache.GCPercent > 0 {
		debug.SetGCPercent(config.Cache.GCPercent)
	}

	if config.Cache.PeriodicRelease > 0 {
		go periodMemoryRelease(config.Cache.PeriodicRelease)
	}

	// start a separate goroutine because a) this blocks and b) we want
	// to return as soon as possible so whatever service is running on
	// this instance can actually start without needing to wait for the
	// clustering to happen

	go cluster.listenForUpdates(self, cache.updatePeers)

	return cache, nil
}

// Periodic release of unused memory back to the OS.
func periodMemoryRelease(interval int) {
	ticker := time.Tick(time.Duration(interval) * time.Second)
	for range ticker {
		debug.FreeOSMemory()
	}
}

func (c *cache) updatePeers(addresses []net.IP) {
	peers := make([]string, len(addresses))
	for i, addr := range addresses {
		peers[i] = fmt.Sprintf("http://%s:%d", addr, c.config.Port)
	}

	sort.Slice(peers, func(i, j int) bool { return peers[i] < peers[j] })

	list := strings.Join(peers, ", ")
	if list == c.prev {
		return
	}

	logger.Debugf("%s set peers %s", c.me, list)
	c.Set(peers...)
	c.prev = list
}

func (c *cache) ListenOn() string {
	return fmt.Sprintf("0.0.0.0:%d", c.config.Port)
}
