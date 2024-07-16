package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/crunchloop/terraform-provider-crunchloop/internal/client"
	"github.com/crunchloop/terraform-provider-crunchloop/internal/services"
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
	service *services.VmService
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

	client, ok := req.ProviderData.(*client.ClientWithResponses)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.ClientWithResponses, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.service = services.NewVmService(client)
}

func (r *VmStateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VmStateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.updateVmStatus(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	data.vmModelToStateResource(vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmStateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VmStateResourceModel

	// Read Terraform prior state data into the model
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, _ := strconv.Atoi(data.VmId.ValueString())
	vm, err := r.service.GetVm(ctx, int32(id))
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	data.vmModelToStateResource(vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmStateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VmStateResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.updateVmStatus(ctx, &data)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	data.vmModelToStateResource(vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmStateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VmStateResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	tflog.Debug(ctx, fmt.Sprintf("Deleting a crunchloop_vm_state resource only stops managing instance state, The vm is left in its current state.: %s", data.VmId.String()))
}

func (r *VmStateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("vm_id"), req, resp)
}

func (d *VmStateResourceModel) vmModelToStateResource(vm *client.VirtualMachine) {
	d.VmId = types.StringValue(strconv.Itoa(int(*vm.Id)))
	d.Status = types.StringValue(string(*vm.Status))
}

func (r *VmStateResource) updateVmStatus(ctx context.Context, data *VmStateResourceModel) (*client.VirtualMachine, error) {
	id, _ := strconv.Atoi(data.VmId.ValueString())
	vm, err := r.service.GetVm(ctx, int32(id))
	if err != nil {
		return nil, err
	}

	if data.Status.String() == string(client.VirtualMachineStatusRunning) && *vm.Status == client.VirtualMachineStatusRunning {
		tflog.Debug(ctx, fmt.Sprintf("Vm is already running: %s", data.VmId.String()))
		return vm, nil
	}

	if data.Status.String() == string(client.VirtualMachineStatusStopped) && *vm.Status == client.VirtualMachineStatusStopped {
		tflog.Debug(ctx, fmt.Sprintf("Vm is already stopped: %s", data.VmId.String()))
		return vm, nil
	}

	switch data.Status.ValueString() {
	case "stopped":
		vm, err = r.service.StopVm(ctx, int32(id))
	case "running":
		vm, err = r.service.StartVm(ctx, int32(id))
	}

	return vm, err
}
