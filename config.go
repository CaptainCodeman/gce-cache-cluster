package cachecluster

import (
	"gopkg.in/ini.v1"
)

type (
	// CacheConfig hold settings for the groupcache service
	CacheConfig struct {
		Port            int
		GCPercent       int `ini:"gc_percent"`
		PeriodicRelease int
	}

	// ClusterConfig hold settings for the cluster service
	ClusterConfig struct {
		Port      int
		Heartbeat int
	}

	// MatchConfig hold settings for instance matching
	MatchConfig struct {
		Project          string
		Zone             string
		NetworkInterface string
		Tags             []string
		Meta             map[string]string `ini:"meta"`
	}

	// Config holds configuration settings
	Config struct {
		Cache   CacheConfig
		Cluster ClusterConfig
		Match   MatchConfig
	}
)

// LoadConfig loads cachecluster Config from an .ini file
// default filename is "cachecluster.ini"
func LoadConfig(filename string) (*Config, error) {
	if filename == "" {
		filename = "cachecluster.ini"
	}
	cfg, err := ini.Load(filename)
	if err != nil {
		return nil, err
	}
	cfg.BlockMode = false
	cfg.NameMapper = ini.TitleUnderscore

	var c Config
	cfg.MapTo(&c)

	c.Match.Meta = cfg.Section("meta").KeysHash()

	return &c, nil
}
