package utils

import (
	"context"
	"fmt"
	"time"

	"github.com/crunchloop/terraform-provider-crunchloop/internal/client"
)

func WaitForVmStatus(ctx context.Context, client *client.ClientWithResponses, id int32, status client.VirtualMachineStatus) error {
	timeout := time.After(5 * time.Minute)    // 5 minutes timeout
	ticker := time.NewTicker(5 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for VM status to become '%s'", status)
		case <-ticker.C:
			vm, err := client.GetVmWithResponse(ctx, id)
			if err != nil {
				return err
			}
			if *vm.JSON200.Status == status {
				return nil
			}
		}
	}
}

func WaitForVmDeletion(ctx context.Context, client *client.ClientWithResponses, id int32) error {
	timeout := time.After(5 * time.Minute)    // 5 minutes timeout
	ticker := time.NewTicker(5 * time.Second) // Check every 10 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return fmt.Errorf("timeout waiting for VM status to get deleted")
		case <-ticker.C:
			vmResponse, err := client.GetVmWithResponse(ctx, id)
			if err != nil {
				return err
			}

			if vmResponse.StatusCode() == 404 {
				return nil
			}
		}
	}
}
