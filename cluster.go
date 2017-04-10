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
		update updateFunc
	}

	updateFunc func(peers []net.IP)
)

func newCluster(config *Config) (*cluster, error) {
	me, err := smudge.GetLocalIP()
	if err != nil {
		return nil, err
	}

	c := &cluster{
		me:     me,
		config: config,
	}

	logger.Debugf("me:%s cluster running on port %d with heartbeat every %dms", me, config.Cluster.Port, config.Cluster.Heartbeat)
	smudge.SetListenPort(config.Cluster.Port)
	smudge.SetHeartbeatMillis(config.Cluster.Heartbeat)
	smudge.SetLogThreshold(smudge.LogError)

	return c, nil
}

func (c *cluster) listenForUpdates(update updateFunc) {
	smudge.AddStatusListener(&statusListener{update})

	if err := c.setInitialNodes(); err != nil {
		logger.Debugf("error setting initial nodes %v", err)

		// TODO: this might mean we _don't_ join the cluster so
		// we really need to do more here than just "give up".
		// at the very least we should keep trying to get the
		// list of nodes we might need to join ...

		return
	}

	smudge.Begin() // NOTE: this blocks !
}

func (c *cluster) setInitialNodes() error {
	m, err := newMatcher(&c.config.Match)
	if err != nil {
		return err
	}

	ips, err := m.getIPAddresses(context.TODO())
	if err != nil {
		return err
	}

	port := uint16(smudge.GetListenPort())
	for _, ip := range ips {
		if !ip.Equal(c.me) {
			node, _ := smudge.CreateNodeByIP(ip, port)
			smudge.AddNode(node)
		}
	}

	return nil
}

func (s *statusListener) OnChange(node *smudge.Node, status smudge.NodeStatus) {
	nodes := smudge.HealthyNodes()
	peers := make([]net.IP, len(nodes))
	for i, node := range nodes {
		peers[i] = node.IP()
	}
	s.update(peers)
}
