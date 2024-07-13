package provider

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/crunchloop/terraform-provider-crunchloop/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VmStateResource{}
var _ resource.ResourceWithImportState = &VmStateResource{}

func NewVmStateResource() resource.Resource {
	return &VmStateResource{}
}

// VmStateResource defines the resource implementation.
type VmStateResource struct {
	client *client.Client
}

// VmStateResourceModel describes the resource data model.
type VmStateResourceModel struct {
	VmId   types.String `tfsdk:"vm_id"`
	Status types.String `tfsdk:"status"`
}

func (r *VmStateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm_state"
}

func (r *VmStateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Vm state resource",

		Attributes: map[string]schema.Attribute{
			"vm_id": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Vm identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Vm status",
			},
		},
	}
}

func (r *VmStateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VmStateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VmStateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// For the purposes of this example code, hardcoding a response value to
	// save into the Terraform state.
	// data.Id = types.StringValue("example-id")

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VmStateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the vm id
	id, _ := strconv.Atoi(data.VmId.ValueString())

	vm, _, err := r.client.VmsService.GetVm(id)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to get VM: %s", err))
		return
	}

	data.VmId = types.StringValue(strconv.Itoa(int(vm.Id)))
	data.Status = types.StringValue(vm.Status)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VmStateResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the vm id
	id, _ := strconv.Atoi(data.VmId.ValueString())

	// Get current status of vm
	vm, _, err := r.client.VmsService.GetVm(id)
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to get VM: %s", err))
		return
	}

	if data.Status.String() == "running" && vm.Status == "running" {
		tflog.Debug(ctx, fmt.Sprintf("Vm is already running: %s", data.VmId.String()))
		return
	}

	if data.Status.String() == "stopped" && vm.Status == "stopped" {
		tflog.Debug(ctx, fmt.Sprintf("Vm is already stopped: %s", data.VmId.String()))
		return
	}

	switch data.Status.ValueString() {
	case "stopped":
		_, _, err = r.client.VmsService.StopVm(id)
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to stop VM: %s", err))
			return
		}
	case "running":
		_, _, err = r.client.VmsService.StartVm(id)
		if err != nil {
			resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to start VM: %s", err))
			return
		}
	}

	err = waitForVmStatus(ctx, r.client, id, data.Status.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("API Error", fmt.Sprintf("Failed to update VM status: %s", err))
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmStateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VmStateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	tflog.Debug(ctx, fmt.Sprintf("Deleting a crunchloop_vm_state resource only stops managing instance state, The vm is left in its current state.: %s", data.VmId.String()))
}

func (r *VmStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("vm_id"), req, resp)
}

func waitForVmStatus(ctx context.Context, client *client.Client, id int, status string) error {
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
			vm, _, err := client.VmsService.GetVm(id)
			if err != nil {
				return err
			}
			if vm.Status == status {
				return nil
			}
		}
	}
}
