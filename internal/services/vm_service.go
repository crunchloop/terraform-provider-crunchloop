package services

import (
	"context"
	"fmt"

	"github.com/crunchloop/terraform-provider-crunchloop/internal/client"
	"github.com/crunchloop/terraform-provider-crunchloop/internal/utils"
)

type VmService struct {
	client *client.ClientWithResponses
}

func NewVmService(client *client.ClientWithResponses) *VmService {
	return &VmService{
		client: client,
	}
}

func (s *VmService) CreateVm(ctx context.Context, options client.CreateVmJSONRequestBody) (*client.VirtualMachine, error) {
	createResponse, err := s.client.CreateVmWithResponse(ctx, options)
	if err != nil {
		return nil, fmt.Errorf("failed to create. Error: %s", err)
	}

	if createResponse.StatusCode() != 201 {
		return nil, fmt.Errorf("failed to create vm. Response body: %s", createResponse.Body)
	}

	err = utils.WaitForVmStatus(ctx, s.client, *createResponse.JSON201.Id, "running")
	if err != nil {
		return nil, err
	}

	vm, err := s.GetVm(ctx, *createResponse.JSON201.Id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *VmService) GetVm(ctx context.Context, id int32) (*client.VirtualMachine, error) {
	response, err := s.client.GetVmWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to create. Error: %s", err)
	}

	if response.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to get vm. Response body: %s", response.Body)
	}

	return response.JSON200, nil
}

func (s *VmService) DeleteVm(ctx context.Context, id int32) error {
	response, err := s.client.DeleteVmWithResponse(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete. Error: %s", err)
	}

	if response.StatusCode() != 204 {
		return fmt.Errorf("failed to delete vm. Response body: %s", response.Body)
	}

	return nil
}

func (s *VmService) UpdateVm(ctx context.Context, id int32, options client.UpdateVmJSONRequestBody) (*client.VirtualMachine, error) {
	vm, err := s.GetVm(ctx, id)
	if err != nil {
		return nil, err
	}

	// We are ready to update the vm now.
	updateResponse, err := s.client.UpdateVmWithResponse(ctx, id, options)
	if err != nil {
		return nil, fmt.Errorf("failed to update. Error: %s", err)
	}

	if updateResponse.StatusCode() != 200 {
		return nil, fmt.Errorf("failed to update vm. Response body: %s", updateResponse.Body)
	}

	// After we issue an update, the vm is going to transition to `updating` state
	// and eventually will be back to `currentStatus` state, we need to wait for that
	// state before moving forward.
	err = utils.WaitForVmStatus(ctx, s.client, *vm.Id, *vm.Status)
	if err != nil {
		return nil, err
	}

	vm, err = s.GetVm(ctx, id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *VmService) StopVm(ctx context.Context, id int32) (*client.VirtualMachine, error) {
	_, err := s.client.StopVmWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to stop vm: %s", err)
	}

	err = utils.WaitForVmStatus(ctx, s.client, id, "stopped")
	if err != nil {
		return nil, fmt.Errorf("failed while waiting for vm to be stopped: %s", err)
	}

	vm, err := s.GetVm(ctx, id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}

func (s *VmService) StartVm(ctx context.Context, id int32) (*client.VirtualMachine, error) {
	_, err := s.client.StartVmWithResponse(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to start vm: %s", err)
	}

	err = utils.WaitForVmStatus(ctx, s.client, id, "running")
	if err != nil {
		return nil, fmt.Errorf("failed while waiting for vm to be running: %s", err)
	}

	vm, err := s.GetVm(ctx, id)
	if err != nil {
		return nil, err
	}

	return vm, nil
}
