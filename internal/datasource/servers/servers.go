package servers

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &ServersDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &ServersDataSource{}
}

// ServersDataSource is the data source implementation.
type ServersDataSource struct {
	client *client.Client
}

// ServersDataSourceModel describes the data source data model.
type ServersDataSourceModel struct {
	Servers []ServerModel `tfsdk:"servers"`
}

// ServerModel describes a single server.
type ServerModel struct {
	ID                  types.String `tfsdk:"id"`
	Name                types.String `tfsdk:"name"`
	Description         types.String `tfsdk:"description"`
	IPAddress           types.String `tfsdk:"ip_address"`
	Port                types.Int64  `tfsdk:"port"`
	Username            types.String `tfsdk:"username"`
	SSHKeyID            types.String `tfsdk:"ssh_key_id"`
	ServerStatus        types.String `tfsdk:"server_status"`
	ServerType          types.String `tfsdk:"server_type"`
	EnableDockerCleanup types.Bool   `tfsdk:"enable_docker_cleanup"`
	OrganizationID      types.String `tfsdk:"organization_id"`
	CreatedAt           types.String `tfsdk:"created_at"`
}

func (d *ServersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

func (d *ServersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all servers from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"servers": schema.ListNestedAttribute{
				Description: "List of all servers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the server.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the server.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the server.",
							Computed:    true,
						},
						"ip_address": schema.StringAttribute{
							Description: "The IP address of the server.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The SSH port of the server.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "The SSH username for the server.",
							Computed:    true,
						},
						"ssh_key_id": schema.StringAttribute{
							Description: "The SSH key ID used to connect to the server.",
							Computed:    true,
						},
						"server_status": schema.StringAttribute{
							Description: "The current status of the server.",
							Computed:    true,
						},
						"server_type": schema.StringAttribute{
							Description: "The type of the server (e.g., deploy).",
							Computed:    true,
						},
						"enable_docker_cleanup": schema.BoolAttribute{
							Description: "Whether Docker cleanup is enabled.",
							Computed:    true,
						},
						"organization_id": schema.StringAttribute{
							Description: "The organization ID the server belongs to.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The creation timestamp of the server.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ServersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ServersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ServersDataSourceModel

	servers, err := d.client.GetServers(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Servers",
			err.Error(),
		)
		return
	}

	// Initialize to empty slice to avoid null in state
	state.Servers = []ServerModel{}

	for _, server := range servers {
		state.Servers = append(state.Servers, ServerModel{
			ID:                  types.StringValue(server.ServerID),
			Name:                types.StringValue(server.Name),
			Description:         types.StringValue(server.Description),
			IPAddress:           types.StringValue(server.IPAddress),
			Port:                types.Int64Value(int64(server.Port)),
			Username:            types.StringValue(server.Username),
			SSHKeyID:            types.StringValue(server.SSHKeyID),
			ServerStatus:        types.StringValue(server.ServerStatus),
			ServerType:          types.StringValue(server.ServerType),
			EnableDockerCleanup: types.BoolValue(server.EnableDockerCleanup),
			OrganizationID:      types.StringValue(server.OrganizationID),
			CreatedAt:           types.StringValue(server.CreatedAt.Format("2006-01-02T15:04:05Z")),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
