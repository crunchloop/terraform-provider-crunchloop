package client

import (
	"fmt"
	"net/http"
)

// ClusterAgentsService handles communication with the cluster agents related
// methods of the GitLab API.
//
// GitLab API docs: https://docs.gitlab.com/ee/api/cluster_agents.html
type VmsService struct {
	client *Client
}

type CreateVolumeOptions struct {
	SizeBytes int64 `json:"size_bytes"`
}

type CreateVmOptions struct {
	Name                    string `json:"name"`
	Status                  string `json:"status"`
	MemoryMegabytes         int64  `json:"memory_megabytes"`
	Cores                   int64  `json:"cores"`
	VmiId                   int64  `json:"vmi_id"`
	HostId                  int64  `json:"host_id"`
	RootVolumeSizeGigabytes int64  `json:"root_volume_size_gigabytes"`
}

type Vm struct {
	Id          int64  `json:"id"`
	Name        string `json:"name"`
	MemoryBytes int64  `json:"memory_bytes"`
	Cores       int64  `json:"cores"`
	Provider    string `json:"provider"`
	Status      string `json:"status"`
	Host        Host   `json:"host"`
	Vmi         Vmi    `json:"vmi"`

	RootVolume struct {
		SizeBytes int64 `json:"size_bytes"`
	} `json:"root_volume"`
}

// CreateVm a new vm.
func (s *VmsService) CreateVm(opt *CreateVmOptions) (*Vm, *http.Response, error) {
	uri := "api/v1/vms"

	req, err := s.client.NewRequest(http.MethodPost, uri, opt)
	if err != nil {
		return nil, nil, err
	}

	vm := new(Vm)
	resp, err := s.client.Do(req, &vm)
	if err != nil {
		return nil, resp, err
	}

	return vm, resp, nil
}

// CreateVm a new vm.
func (s *VmsService) DeleteVm(id int) (*http.Response, error) {
	uri := fmt.Sprintf("api/v1/vms/%d", int(id))

	req, err := s.client.NewRequest(http.MethodDelete, uri, nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.Do(req, nil)

	return resp, err
}

// GetVm gets a single vm.
func (s *VmsService) GetVm(id int) (*Vm, *http.Response, error) {
	uri := fmt.Sprintf("/api/v1/vms/%d", int(id))

	req, err := s.client.NewRequest(http.MethodGet, uri, nil)
	if err != nil {
		return nil, nil, err
	}

	vm := new(Vm)
	resp, err := s.client.Do(req, &vm)
	if err != nil {
		return nil, resp, err
	}

	return vm, resp, nil
}

// StartVm starts a single vm.
func (s *VmsService) StartVm(id int) (*Vm, *http.Response, error) {
	uri := fmt.Sprintf("api/v1/vms/%d/start", int(id))

	req, err := s.client.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, nil, err
	}

	vm := new(Vm)
	resp, err := s.client.Do(req, &vm)
	if err != nil {
		return nil, resp, err
	}

	return vm, resp, nil
}

// StopVm stops a single vm.
func (s *VmsService) StopVm(id int) (*Vm, *http.Response, error) {
	uri := fmt.Sprintf("api/v1/vms/%d/stop", int(id))

	req, err := s.client.NewRequest(http.MethodPost, uri, nil)
	if err != nil {
		return nil, nil, err
	}

	vm := new(Vm)
	resp, err := s.client.Do(req, &vm)
	if err != nil {
		return nil, resp, err
	}

	return vm, resp, nil
}
