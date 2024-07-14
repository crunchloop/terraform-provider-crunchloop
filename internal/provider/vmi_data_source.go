package provider

import (
	"context"
	"fmt"
	"strconv"

	"github.com/crunchloop/terraform-provider-crunchloop/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ datasource.DataSource = &VmiDataSource{}

func NewVmiDataSource() datasource.DataSource {
	return &VmiDataSource{}
}

// VmiDataSource defines the data source implementation.
type VmiDataSource struct {
	client *client.ClientWithResponses
}

// VmiDataSourceModel describes the data source data model.
type VmiDataSourceModel struct {
	Id   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
}

func (d *VmiDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vmi"
}

func (d *VmiDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		// This description is used by the documentation generator and the language server.
		MarkdownDescription: "Vmi data source",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Vmi identifier",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Vmi name",
				Required:            true,
			},
		},
	}
}

func (d *VmiDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ClientWithResponses)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

func (d *VmiDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data VmiDataSourceModel

	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	response, err := d.client.ListVmisWithResponse(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"API Error",
			fmt.Sprintf("Unable to read vmi, got error: %s", err),
		)
		return
	}

	for _, vmi := range *response.JSON200.Data {
		if *vmi.Name == data.Name.ValueString() {
			data.Id = types.StringValue(strconv.Itoa(int(*vmi.Id)))
			data.Name = types.StringValue(*vmi.Name)

			// Write logs using the tflog package
			// Documentation: https://terraform.io/plugin/log
			tflog.Trace(ctx, "read Vmi data source")

			// Save data into Terraform state
			resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

			return
		}
	}

	resp.Diagnostics.AddError(
		"API Error",
		fmt.Sprintf("Vmi with name %s was not found", data.Name.ValueString()),
	)
}
