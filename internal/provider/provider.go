package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"

	// Data sources
	applicationds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/application"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/environments"
	projectds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/project"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/projects"
	serverds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/server"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/servers"
	sshkeyds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/sshkey"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/sshkeys"

	// Resources
	applicationrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/application"
	certificaters "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/certificate"
	composers "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/compose"
	destinationrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/destination"
	environmentrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/environment"
	mariadbrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/mariadb"
	mongors "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/mongo"
	mysqlrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/mysql"
	postgresrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/postgres"
	projectrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/project"
	redisrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/redis"
	registryrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/registry"
	serverrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/server"
	sshkeyrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/sshkey"
)

// Ensure DokployProvider satisfies various provider interfaces.
var _ provider.Provider = &DokployProvider{}

// DokployProvider defines the provider implementation.
type DokployProvider struct {
	version string
}

// DokployProviderModel describes the provider data model.
type DokployProviderModel struct {
	Host   types.String `tfsdk:"host"`
	APIKey types.String `tfsdk:"api_key"`
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &DokployProvider{
			version: version,
		}
	}
}

func (p *DokployProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "dokploy"
	resp.Version = p.version
}

func (p *DokployProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with Dokploy infrastructure management platform.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The Dokploy server URL. Can also be set via DOKPLOY_HOST environment variable.",
				Optional:    true,
			},
			"api_key": schema.StringAttribute{
				Description: "The Dokploy API key. Can also be set via DOKPLOY_API_KEY environment variable.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *DokployProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config DokployProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use environment variables as fallback
	host := os.Getenv("DOKPLOY_HOST")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	apiKey := os.Getenv("DOKPLOY_API_KEY")
	if !config.APIKey.IsNull() {
		apiKey = config.APIKey.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddError(
			"Missing Host Configuration",
			"The provider requires a host to be configured. "+
				"Set the host value in the provider configuration or use the DOKPLOY_HOST environment variable.",
		)
		return
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"Missing API Key Configuration",
			"The provider requires an API key to be configured. "+
				"Set the api_key value in the provider configuration or use the DOKPLOY_API_KEY environment variable.",
		)
		return
	}

	// Create API client
	apiClient := client.New(host, apiKey)

	// Make the client available to data sources and resources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *DokployProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		projectrs.NewResource,
		serverrs.NewResource,
		environmentrs.NewResource,
		applicationrs.NewResource,
		composers.NewResource,
		postgresrs.NewResource,
		mysqlrs.NewResource,
		mariadbrs.NewResource,
		mongors.NewResource,
		redisrs.NewResource,
		sshkeyrs.NewResource,
		registryrs.NewResource,
		certificaters.NewResource,
		destinationrs.NewResource,
	}
}

func (p *DokployProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		projects.NewDataSource,
		projectds.NewDataSource,
		servers.NewDataSource,
		serverds.NewDataSource,
		sshkeys.NewDataSource,
		sshkeyds.NewDataSource,
		environments.NewDataSource,
		applicationds.NewDataSource,
	}
}
