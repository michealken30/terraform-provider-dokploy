package server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &ServerDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &ServerDataSource{}
}

// ServerDataSource is the data source implementation.
type ServerDataSource struct {
	client *client.Client
}

// ServerDataSourceModel describes the data source data model.
type ServerDataSourceModel struct {
	ID                  types.String `tfsdk:"id"`
	ServerID            types.String `tfsdk:"server_id"`
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

func (d *ServerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *ServerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single server from Dokploy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the server (same as server_id).",
				Computed:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "The server ID to look up. Mutually exclusive with name.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The server name to look up. Mutually exclusive with server_id.",
				Optional:    true,
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
	}
}

func (d *ServerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ServerDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either server_id or name is provided
	if config.ServerID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either server_id or name must be specified.",
		)
		return
	}

	if !config.ServerID.IsNull() && !config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Only one of server_id or name can be specified, not both.",
		)
		return
	}

	var server *client.Server
	var err error

	if !config.ServerID.IsNull() {
		server, err = d.client.GetServer(ctx, config.ServerID.ValueString())
	} else {
		server, err = d.client.GetServerByName(ctx, config.Name.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Server",
			err.Error(),
		)
		return
	}

	state := ServerDataSourceModel{
		ID:                  types.StringValue(server.ServerID),
		ServerID:            types.StringValue(server.ServerID),
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
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
