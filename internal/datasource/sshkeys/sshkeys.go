package sshkeys

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &SSHKeysDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &SSHKeysDataSource{}
}

// SSHKeysDataSource is the data source implementation.
type SSHKeysDataSource struct {
	client *client.Client
}

// SSHKeysDataSourceModel describes the data source data model.
type SSHKeysDataSourceModel struct {
	SSHKeys []SSHKeyModel `tfsdk:"ssh_keys"`
}

// SSHKeyModel describes a single SSH key (without private key for security).
type SSHKeyModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	PublicKey      types.String `tfsdk:"public_key"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
	LastUsedAt     types.String `tfsdk:"last_used_at"`
}

func (d *SSHKeysDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_keys"
}

func (d *SSHKeysDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all SSH keys from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"ssh_keys": schema.ListNestedAttribute{
				Description: "List of all SSH keys.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the SSH key.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the SSH key.",
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
				},
			},
		},
	}
}

func (d *SSHKeysDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SSHKeysDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state SSHKeysDataSourceModel

	keys, err := d.client.GetSSHKeys(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy SSH Keys",
			err.Error(),
		)
		return
	}

	for _, key := range keys {
		state.SSHKeys = append(state.SSHKeys, SSHKeyModel{
			ID:             types.StringValue(key.SSHKeyID),
			Name:           types.StringValue(key.Name),
			Description:    types.StringValue(key.Description),
			PublicKey:      types.StringValue(key.PublicKey),
			OrganizationID: types.StringValue(key.OrganizationID),
			CreatedAt:      types.StringValue(key.CreatedAt.Format("2006-01-02T15:04:05Z")),
			LastUsedAt:     types.StringValue(key.LastUsedAt.Format("2006-01-02T15:04:05Z")),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
