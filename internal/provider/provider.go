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
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/certificates"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/deployments"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/destinations"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/environments"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/github"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/gitproviders"
	notificationds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/notification"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/notifications"
	organizationds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/organization"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/organizations"
	projectds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/project"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/projects"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/registries"
	serverds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/server"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/servers"
	sshkeyds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/sshkey"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/sshkeys"
	userds "github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/user"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/datasource/users"

	// Resources
	applicationrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/application"
	backuprs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/backup"
	bitbucketrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/bitbucket"
	bootstraprs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/bootstrap"
	certificaters "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/certificate"
	composers "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/compose"
	destinationrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/destination"
	domainrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/domain"
	environmentrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/environment"
	gitears "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/gitea"
	gitlabrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/gitlab"
	mariadbrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/mariadb"
	mongors "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/mongo"
	mountrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/mount"
	mysqlrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/mysql"
	notificationrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/notification"
	organizationrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/organization"
	portrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/port"
	postgresrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/postgres"
	projectrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/project"
	redirectrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/redirect"
	redisrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/redis"
	registryrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/registry"
	schedulers "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/schedule"
	securityrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/security"
	serverrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/server"
	sshkeyrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/sshkey"
	userpermissionsrs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/userpermissions"
	volumebackuprs "github.com/reserve-protocol/terraform-provider-dokploy/internal/resource/volumebackup"
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
				Description: "The Dokploy API key. Can also be set via DOKPLOY_API_KEY environment variable. " +
					"Optional for bootstrap resource, required for all other resources.",
				Optional:  true,
				Sensitive: true,
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

	// API key is optional - bootstrap resource can work without it
	// Other resources will fail if they need the client and api_key is not set

	// Create API client (may have empty apiKey for bootstrap-only usage)
	apiClient := client.New(host, apiKey)

	// Make the client available to data sources and resources
	resp.DataSourceData = apiClient
	resp.ResourceData = apiClient
}

func (p *DokployProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		bootstraprs.NewResource,
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
		domainrs.NewResource,
		portrs.NewResource,
		mountrs.NewResource,
		securityrs.NewResource,
		redirectrs.NewResource,
		backuprs.NewResource,
		schedulers.NewResource,
		notificationrs.NewResource,
		organizationrs.NewResource,
		gitlabrs.NewResource,
		bitbucketrs.NewResource,
		gitears.NewResource,
		userpermissionsrs.NewResource,
		volumebackuprs.NewResource,
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
		certificates.NewDataSource,
		destinations.NewDataSource,
		registries.NewDataSource,
		users.NewDataSource,
		userds.NewDataSource,
		organizations.NewDataSource,
		organizationds.NewDataSource,
		notifications.NewDataSource,
		notificationds.NewDataSource,
		gitproviders.NewDataSource,
		github.NewDataSource,
		deployments.NewDataSource,
	}
}
