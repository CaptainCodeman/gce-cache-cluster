package cachecluster

import (
	"context"
	"net"

	"github.com/clockworksoul/smudge"
)

type (
	cluster struct {
		me     net.IP
		config *Config
	}

	statusListener struct {
		self   string
		update updateFunc
	}

	updateFunc func(peers []net.IP)
)

func newCluster(config *Config) (*cluster, error) {
	me, err := GetLocalIP()
	if err != nil {
		return nil, err
	}

	c := &cluster{
		me:     me,
		config: config,
	}

	logger.Debugf("me:%s cluster running", me)
	// logger.Debugf("me:%s cluster running on port %d with heartbeat every %dms", me, config.Cluster.Port, config.Cluster.Heartbeat)
	smudge.SetListenIP(me)
	smudge.SetListenPort(config.Cluster.Port)
	// smudge.SetHeartbeatMillis(config.Cluster.Heartbeat)
	// smudge.SetMinPingTime(500)
	smudge.SetLogThreshold(smudge.LogWarn)
	smudge.SetLogger(logger)

	return c, nil
}

func (c *cluster) listenForUpdates(self string, update updateFunc) {
	smudge.AddStatusListener(&statusListener{self, update})

	if err := c.setInitialNodes(smudge.GetListenPort()); err != nil {
		logger.Debugf("error setting initial nodes %v", err)

		// TODO: this might mean we _don't_ join the cluster so
		// we really need to do more here than just "give up".
		// at the very least we should keep trying to get the
		// list of nodes we might need to join ...
	}

	smudge.Begin() // NOTE: this blocks !
}

func (c *cluster) setInitialNodes(port int) error {
	m, err := newMatcher(&c.config.Match)
	if err != nil {
		return err
	}

	ips, err := m.getIPAddresses(context.TODO())
	if err != nil {
		return err
	}

	p := uint16(port)
	for _, ip := range ips {
		if !ip.Equal(c.me) {
			node, _ := smudge.CreateNodeByIP(ip, p)
			smudge.AddNode(node)
		}
	}

	return nil
}

func (s *statusListener) OnChange(node *smudge.Node, status smudge.NodeStatus) {
	peers := []net.IP{}
	if status == smudge.StatusAlive {
		peers = append(peers, node.IP())
	}

	nodes := smudge.HealthyNodes()
	for _, n := range nodes {
		if !n.IP().Equal(node.IP()) {
			peers = append(peers, n.IP())
		}
	}

	s.update(peers)
}

func GetLocalIP() (net.IP, error) {
	var ip net.IP

	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return ip, err
	}

	for _, a := range addrs {
		if ipnet, ok := a.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				ip = ipnet.IP.To4()
				break
			}
		}
	}

	return ip, nil
}
