package provider

import (
	"context"

	"github.com/crunchloop/terraform-provider-crunchloop/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure CrunchloopProvider satisfies various provider interfaces.
var _ provider.Provider = &CrunchloopProvider{}
var _ provider.ProviderWithFunctions = &CrunchloopProvider{}

// CrunchloopProvider defines the provider implementation.
type CrunchloopProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// crunchloopProviderModel describes the provider data model.
type crunchloopProviderModel struct {
	Url types.String `tfsdk:"url"`
}

func (p *CrunchloopProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "crunchloop"
	resp.Version = p.version
}

func (p *CrunchloopProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "URL for the Crunchloop Cloud instance",
				Required:            true,
			},
		},
	}
}

func (p *CrunchloopProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data crunchloopProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Example client configuration for data sources and resources
	client, err := client.NewClient(client.WithBaseURL(data.Url.ValueString()))

	if err != nil {
		tflog.Info(ctx, err.Error())
		resp.Diagnostics.AddError("Failed to create Crunchloop Cloud client", "")
		return
	}

	resp.DataSourceData = client
	resp.ResourceData = client
}

func (p *CrunchloopProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVmResource,
		NewVmStateResource,
	}
}

func (p *CrunchloopProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewHostDataSource,
		NewVmiDataSource,
	}
}

func (p *CrunchloopProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CrunchloopProvider{
			version: version,
		}
	}
}
