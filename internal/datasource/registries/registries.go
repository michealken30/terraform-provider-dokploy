package registries

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &RegistriesDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &RegistriesDataSource{}
}

// RegistriesDataSource is the data source implementation.
type RegistriesDataSource struct {
	client *client.Client
}

// RegistriesDataSourceModel describes the data source data model.
type RegistriesDataSourceModel struct {
	Registries []RegistryModel `tfsdk:"registries"`
}

// RegistryModel describes a single container registry (without password for security).
type RegistryModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Username       types.String `tfsdk:"username"`
	URL            types.String `tfsdk:"url"`
	ImagePrefix    types.String `tfsdk:"image_prefix"`
	RegistryType   types.String `tfsdk:"type"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *RegistriesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registries"
}

func (d *RegistriesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all container registries from Dokploy. Use this data source to discover registry IDs for importing existing resources.",
		Attributes: map[string]schema.Attribute{
			"registries": schema.ListNestedAttribute{
				Description: "List of all container registries.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the registry.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the registry.",
							Computed:    true,
						},
						"username": schema.StringAttribute{
							Description: "The username for registry authentication.",
							Computed:    true,
						},
						"url": schema.StringAttribute{
							Description: "The registry URL.",
							Computed:    true,
						},
						"image_prefix": schema.StringAttribute{
							Description: "The image prefix for the registry.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The registry type (selfHosted, docker, github, etc.).",
							Computed:    true,
						},
						"organization_id": schema.StringAttribute{
							Description: "The organization ID the registry belongs to.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The creation timestamp of the registry.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *RegistriesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *RegistriesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state RegistriesDataSourceModel

	registries, err := d.client.GetRegistries(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Registries",
			err.Error(),
		)
		return
	}

	// Initialize to empty slice to avoid null in state
	state.Registries = []RegistryModel{}

	for _, reg := range registries {
		imagePrefix := ""
		if reg.ImagePrefix != nil {
			imagePrefix = *reg.ImagePrefix
		}

		state.Registries = append(state.Registries, RegistryModel{
			ID:             types.StringValue(reg.RegistryID),
			Name:           types.StringValue(reg.RegistryName),
			Username:       types.StringValue(reg.Username),
			URL:            types.StringValue(reg.RegistryURL),
			ImagePrefix:    types.StringValue(imagePrefix),
			RegistryType:   types.StringValue(reg.RegistryType),
			OrganizationID: types.StringValue(reg.OrganizationID),
			CreatedAt:      types.StringValue(reg.CreatedAt.Format("2006-01-02T15:04:05Z")),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
