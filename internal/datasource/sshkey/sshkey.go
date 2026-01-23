package sshkey

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &SSHKeyDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &SSHKeyDataSource{}
}

// SSHKeyDataSource is the data source implementation.
type SSHKeyDataSource struct {
	client *client.Client
}

// SSHKeyDataSourceModel describes the data source data model.
type SSHKeyDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	SSHKeyID       types.String `tfsdk:"ssh_key_id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	PublicKey      types.String `tfsdk:"public_key"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	LastUsedAt     types.String `tfsdk:"last_used_at"`
}

func (d *SSHKeyDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *SSHKeyDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single SSH key from Dokploy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the SSH key (same as ssh_key_id).",
				Computed:    true,
			},
			"ssh_key_id": schema.StringAttribute{
				Description: "The SSH key ID to look up. Mutually exclusive with name.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The SSH key name to look up. Mutually exclusive with ssh_key_id.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the SSH key.",
				Computed:    true,
			},
			"public_key": schema.StringAttribute{
				Description: "The public key.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID the SSH key belongs to.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp of the SSH key.",
				Computed:    true,
			},
			"last_used_at": schema.StringAttribute{
				Description: "The last used timestamp of the SSH key.",
				Computed:    true,
			},
		},
	}
}

func (d *SSHKeyDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SSHKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SSHKeyDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either ssh_key_id or name is provided
	if config.SSHKeyID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either ssh_key_id or name must be specified.",
		)
		return
	}

	if !config.SSHKeyID.IsNull() && !config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Only one of ssh_key_id or name can be specified, not both.",
		)
		return
	}

	var key *client.SSHKey
	var err error

	if !config.SSHKeyID.IsNull() {
		key, err = d.client.GetSSHKey(ctx, config.SSHKeyID.ValueString())
	} else {
		key, err = d.client.GetSSHKeyByName(ctx, config.Name.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy SSH Key",
			err.Error(),
		)
		return
	}

	state := SSHKeyDataSourceModel{
		ID:             types.StringValue(key.SSHKeyID),
		SSHKeyID:       types.StringValue(key.SSHKeyID),
		Name:           types.StringValue(key.Name),
		Description:    types.StringValue(key.Description),
		PublicKey:      types.StringValue(key.PublicKey),
		OrganizationID: types.StringValue(key.OrganizationID),
		CreatedAt:      types.StringValue(key.CreatedAt.Format("2006-01-02T15:04:05Z")),
		LastUsedAt:     types.StringValue(key.LastUsedAt.Format("2006-01-02T15:04:05Z")),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
