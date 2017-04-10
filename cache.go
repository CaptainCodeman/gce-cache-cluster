package cachecluster

import (
	"fmt"
	"net"
	"sort"
	"strings"

	"github.com/golang/groupcache"
)

type (
	cache struct {
		*groupcache.HTTPPool
		me     net.IP
		peers  []string
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

	// start a separate goroutine because a) this blocks and b) we want
	// to return as soon as possible so whatever service is running on
	// this instance can actually start without needing to wait for the
	// clustering to happen

	go cluster.listenForUpdates(cache.updatePeers)

	return cache, nil
}

func (c *cache) updatePeers(peerAddresses []net.IP) {
	peers := make([]string, len(peerAddresses))
	for i, addr := range peerAddresses {
		peers[i] = fmt.Sprintf("http://%s:%d", addr, c.config.Port)
	}

	sort.Slice(peers, func(i, j int) bool { return peers[i] < peers[j] })

	if strings.Join(c.peers, ", ") == strings.Join(peers, ", ") {
		return
	}

	logger.Debugf("%s set peers %s", c.me, strings.Join(peers, ", "))
	c.Set(peers...)
	c.peers = peers
}

func (c *cache) ListenOn() string {
	return fmt.Sprintf("0.0.0.0:%d", c.config.Port)
}
