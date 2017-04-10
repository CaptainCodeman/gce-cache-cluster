package cachecluster

import (
	"context"
	"net"

	"io/ioutil"
	"net/http"

	"cloud.google.com/go/compute/metadata"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/compute/v1"
)

type (
	matcher struct {
		service *compute.Service
		config  *MatchConfig
	}
)

func newMatcher(config *MatchConfig) (*matcher, error) {
	ctx := context.Background()

	var err error
	var hc *http.Client

	config.Project, err = metadata.ProjectID()
	config.Zone, err = metadata.Zone()

	b, err := ioutil.ReadFile("service-account.json")
	if err != nil {
		return nil, err
	}
	jc, err := google.JWTConfigFromJSON(b, compute.ComputeReadonlyScope)
	if err != nil {
		return nil, err
	}
	hc = jc.Client(ctx)

	service, err := compute.New(hc)
	if err != nil {
		return nil, err
	}

	im := matcher{
		config:  config,
		service: service,
	}

	return &im, nil
}

func (m *matcher) getIPAddresses(ctx context.Context) ([]net.IP, error) {
	instances, err := m.getInstances(ctx)
	if err != nil {
		return nil, err
	}

	results := make([]net.IP, 0, len(instances))
	for _, instance := range instances {
		for _, nic := range instance.NetworkInterfaces {
			if nic.Name == m.config.NetworkInterface || len(instance.NetworkInterfaces) == 1 {
				ip := net.ParseIP(nic.NetworkIP)
				results = append(results, ip)
			}
		}
	}
	return results, nil
}

func (m *matcher) getInstances(ctx context.Context) ([]*compute.Instance, error) {
	results := make([]*compute.Instance, 0)

	logger.Debugf("list instances in %s / %s", m.config.Project, m.config.Zone)
	req := m.service.Instances.List(m.config.Project, m.config.Zone)
	req.Filter("status eq RUNNING")

	if err := req.Pages(ctx, func(page *compute.InstanceList) error {
		for _, instance := range page.Items {
			logger.Debugf("found %s", instance.Name)
			if m.matchesConfig(instance) {
				results = append(results, instance)
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}

	return results, nil
}

func (m *matcher) matchesConfig(instance *compute.Instance) bool {
	for _, tag := range m.config.Tags {
		found := false
		for _, itag := range instance.Tags.Items {
			if tag == itag {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}

	for k, v := range m.config.Meta {
		found := false
		for _, imeta := range instance.Metadata.Items {
			if k == imeta.Key && v == *imeta.Value {
				found = true
				break
			}
		}
		if !found {
			return false
		}
	}
	return true
}
