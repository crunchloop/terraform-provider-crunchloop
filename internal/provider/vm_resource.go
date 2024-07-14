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
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VmResource{}
var _ resource.ResourceWithImportState = &VmResource{}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

// VmResource defines the resource implementation.
type VmResource struct {
	client *client.ClientWithResponses
}

// VmResourceModel describes the resource data model.
type VmResourceModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	MemoryMegabytes         types.Int64  `tfsdk:"memory_megabytes"`
	Cores                   types.Int64  `tfsdk:"cores"`
	VmiId                   types.Int64  `tfsdk:"vmi_id"`
	HostId                  types.Int64  `tfsdk:"host_id"`
	RootVolumeSizeGigabytes types.Int64  `tfsdk:"root_volume_size_gigabytes"`
}

func (r *VmResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (r *VmResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Vm resource",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Vm identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Vm name",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"memory_megabytes": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Vm memory in megabytes",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"cores": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Vm cores",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"vmi_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Vm Vmi id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"host_id": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Vm Host id",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"root_volume_size_gigabytes": schema.Int64Attribute{
				Required:            true,
				MarkdownDescription: "Vm root volume size in gigabytes",
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *VmResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ClientWithResponses)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *http.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *VmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VmResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	createOptions := client.CreateVmJSONRequestBody{
		Name:                    data.Name.ValueString(),
		MemoryMegabytes:         int(data.MemoryMegabytes.ValueInt64()),
		Cores:                   int(data.Cores.ValueInt64()),
		VmiId:                   int(data.VmiId.ValueInt64()),
		HostId:                  int(data.HostId.ValueInt64()),
		RootVolumeSizeGigabytes: int(data.RootVolumeSizeGigabytes.ValueInt64()),
	}

	vm, err := r.client.CreateVmWithResponse(ctx, createOptions)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed to create vm: %s", err),
		)
		return
	}

	data.vmModelToStateResource(vm.JSON201)

	// Write logs using the tflog package
	// Documentation: https://terraform.io/plugin/log
	tflog.Trace(ctx, "created a resource")

	// Save data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

	// User expects the VM to be running, so wait for it to be running
	err = waitForVmStatus(ctx, r.client, int(*vm.JSON201.Id), "running")
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed while waiting for vm to be running: %s", err),
		)
	}
}

func (r *VmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the vm id
	id, _ := strconv.Atoi(data.Id.ValueString())

	vm, err := r.client.GetVmWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed to get vm: %s", err),
		)
		return
	}

	data.vmModelToStateResource(vm.JSON200)

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VmResourceModel

	// Read Terraform plan data into the model
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Save updated data into Terraform state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VmResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the vm id
	id, _ := strconv.Atoi(data.Id.ValueString())

	_, err := r.client.DeleteVmWithResponse(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed to delete vm: %s", err),
		)
		return
	}

	err = waitForVmDeletion(ctx, r.client, id)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Failed while waiting for vm to get deleted: %s", err),
		)
		return
	}
}

func (r *VmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (d *VmResourceModel) vmModelToStateResource(vm *client.VirtualMachine) {
	d.Id = types.StringValue(strconv.Itoa(int(*vm.Id)))
	d.Name = types.StringValue(*vm.Name)
	d.MemoryMegabytes = types.Int64Value(int64(bytesToMegabytes(*vm.MemoryBytes)))
	d.Cores = types.Int64Value(int64(*vm.Cores))
	d.VmiId = types.Int64Value(int64(*vm.Vmi.Id))
	d.HostId = types.Int64Value(int64(*vm.Host.Id))
	d.RootVolumeSizeGigabytes = types.Int64Value(int64(bytesToGigabytes(int64(*vm.RootVolume.SizeBytes))))
}

func bytesToMegabytes(bytes int) int {
	return bytes / 1024 / 1024
}

func bytesToGigabytes(bytes int64) int64 {
	return bytes / 1024 / 1024 / 1024
}

func waitForVmDeletion(ctx context.Context, client *client.ClientWithResponses, id int) error {
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
