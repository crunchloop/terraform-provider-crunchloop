package client

import (
	"fmt"
	"net/http"
)

// ClusterAgentsService handles communication with the cluster agents related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/cluster_agents.html
type VmisService struct {
	client *Client
}

type Vmi struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
}

type VmisList struct {
	Vmis []*Vmi `json:"data"`
}

func (s *VmisService) ListVmis() (*VmisList, *http.Response, error) {
	uri := "/api/v1/vmis"

	req, err := s.client.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, nil, err
	}

	vmiList := new(VmisList)
	resp, err := s.client.Do(req, &vmiList)
	if err != nil {
		return nil, resp, err
	}

	return vmiList, resp, nil
}

// GetVm gets a single vm.
func (s *VmisService) GetVmiByName(name string) (*Vmi, *http.Response, error) {
	list, _, err := s.ListVmis()
	if err != nil {
		return nil, nil, err
	}

	for _, vmi := range list.Vmis {
		if vmi.Name == name {
			return vmi, nil, nil
		}
	}

	return nil, nil, fmt.Errorf("vmi not found")
}
