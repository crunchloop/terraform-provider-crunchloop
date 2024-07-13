package client

import (
	"fmt"
	"net/http"
)

// ClusterAgentsService handles communication with the cluster agents related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/cluster_agents.html
type HostsService struct {
	client *Client
}

type Host struct {
	Id     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

type HostList struct {
	Hosts []*Host `json:"data"`
}

func (s *HostsService) ListHosts() (*HostList, *http.Response, error) {
	uri := "/api/v1/hosts"

	req, err := s.client.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, nil, err
	}

	hostList := new(HostList)
	resp, err := s.client.Do(req, &hostList)
	if err != nil {
		return nil, resp, err
	}

	return hostList, resp, nil
}

// GetVm gets a single vm.
func (s *HostsService) GetHostByName(name string) (*Host, *http.Response, error) {
	list, _, err := s.ListHosts()
	if err != nil {
		return nil, nil, err
	}

	for _, host := range list.Hosts {
		if host.Name == name {
			return host, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("host not found")
}
