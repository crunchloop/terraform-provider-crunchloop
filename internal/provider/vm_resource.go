package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/crunchloop/terraform-provider-crunchloop/internal/client"
	"github.com/crunchloop/terraform-provider-crunchloop/internal/services"
	"github.com/crunchloop/terraform-provider-crunchloop/internal/utils"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int32planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &VmResource{}
var _ resource.ResourceWithImportState = &VmResource{}

func NewVmResource() resource.Resource {
	return &VmResource{}
}

// VmResource defines the resource implementation.
type VmResource struct {
	service *services.VmService
}

// VmResourceModel describes the resource data model.
type VmResourceModel struct {
	Id                      types.String `tfsdk:"id"`
	Name                    types.String `tfsdk:"name"`
	MemoryMegabytes         types.Int32  `tfsdk:"memory_megabytes"`
	Cores                   types.Int32  `tfsdk:"cores"`
	VmiId                   types.Int32  `tfsdk:"vmi_id"`
	HostId                  types.Int32  `tfsdk:"host_id"`
	RootVolumeSizeGigabytes types.Int32  `tfsdk:"root_volume_size_gigabytes"`
	UserData                types.String `tfsdk:"user_data"`
	SshKey                  types.String `tfsdk:"ssh_key"`
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
				MarkdownDescription: "Identifier",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Name of the Vm",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"memory_megabytes": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "Memory (MiB)",
			},
			"cores": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "Virtual CPU cores",
			},
			"vmi_id": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "Identifier of the VMI to use for the Vm",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"host_id": schema.Int32Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Identifier of the Host where the Vm will be created",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.UseStateForUnknown(),
				},
			},
			"root_volume_size_gigabytes": schema.Int32Attribute{
				Required:            true,
				MarkdownDescription: "Root volume size (GiB)",
				PlanModifiers: []planmodifier.Int32{
					int32planmodifier.RequiresReplace(),
				},
			},
			"ssh_key": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Ssh public key to authenticate with the Vm",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_data": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Cloud init user data shell script, base64 encoded",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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
			fmt.Sprintf("Expected *client.ClientWithResponses, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.service = services.NewVmService(client)
}

func (r *VmResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data VmResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createOptions := client.CreateVmJSONRequestBody{
		Name:                    data.Name.ValueString(),
		MemoryMegabytes:         data.MemoryMegabytes.ValueInt32(),
		Cores:                   data.Cores.ValueInt32(),
		VmiId:                   data.VmiId.ValueInt32(),
		RootVolumeSizeGigabytes: data.RootVolumeSizeGigabytes.ValueInt32(),
	}

	if data.HostId.ValueInt32() != 0 {
		createOptions.HostId = data.HostId.ValueInt32Pointer()
	}

	if data.UserData.ValueString() != "" {
		createOptions.UserData = data.UserData.ValueStringPointer()
	}

	if data.SshKey.ValueString() != "" {
		createOptions.SshKey = data.SshKey.ValueStringPointer()
	}

	vm, err := r.service.CreateVm(ctx, createOptions)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	data.vmModelToStateResource(vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data VmResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	id, _ := strconv.Atoi(data.Id.ValueString())
	vm, err := r.service.GetVm(ctx, int32(id))
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	data.vmModelToStateResource(vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data VmResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateOptions := client.UpdateVmJSONRequestBody{
		MemoryMegabytes: data.MemoryMegabytes.ValueInt32Pointer(),
		Cores:           data.Cores.ValueInt32Pointer(),
	}

	// Parse the vm id
	id, _ := strconv.Atoi(data.Id.ValueString())
	vm, err := r.service.UpdateVm(ctx, int32(id), updateOptions)
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}

	data.vmModelToStateResource(vm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *VmResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data VmResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Parse the vm id
	id, _ := strconv.Atoi(data.Id.ValueString())
	err := r.service.DeleteVm(ctx, int32(id))
	if err != nil {
		resp.Diagnostics.AddError("API Error", err.Error())
		return
	}
}

func (r *VmResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (d *VmResourceModel) vmModelToStateResource(vm *client.VirtualMachine) {
	d.Id = types.StringValue(strconv.Itoa(int(*vm.Id)))
	d.Name = types.StringValue(*vm.Name)
	d.VmiId = types.Int32Value(*vm.Vmi.Id)
	d.HostId = types.Int32Value(*vm.Host.Id)
	d.Cores = types.Int32Value(*vm.Cores)
	d.MemoryMegabytes = types.Int32Value(utils.BytesToMegabytes(*vm.MemoryBytes))
	d.RootVolumeSizeGigabytes = types.Int32Value(utils.BytesToGigabytes(*vm.RootVolume.SizeBytes))
}
