package destinations

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &DestinationsDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &DestinationsDataSource{}
}

// DestinationsDataSource is the data source implementation.
type DestinationsDataSource struct {
	client *client.Client
}

// DestinationsDataSourceModel describes the data source data model.
type DestinationsDataSourceModel struct {
	Destinations []DestinationModel `tfsdk:"destinations"`
}

// DestinationModel describes a single backup destination (without sensitive keys).
type DestinationModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Bucket         types.String `tfsdk:"bucket"`
	Region         types.String `tfsdk:"region"`
	Endpoint       types.String `tfsdk:"endpoint"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *DestinationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destinations"
}

func (d *DestinationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all backup destinations from Dokploy. Use this data source to discover destination IDs for importing existing resources.",
		Attributes: map[string]schema.Attribute{
			"destinations": schema.ListNestedAttribute{
				Description: "List of all backup destinations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the destination.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the destination.",
							Computed:    true,
						},
						"bucket": schema.StringAttribute{
							Description: "The S3 bucket name.",
							Computed:    true,
						},
						"region": schema.StringAttribute{
							Description: "The AWS region.",
							Computed:    true,
						},
						"endpoint": schema.StringAttribute{
							Description: "The S3-compatible endpoint URL.",
							Computed:    true,
						},
						"organization_id": schema.StringAttribute{
							Description: "The organization ID the destination belongs to.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The creation timestamp of the destination.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DestinationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DestinationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state DestinationsDataSourceModel

	destinations, err := d.client.GetDestinations(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Destinations",
			err.Error(),
		)
		return
	}

	// Initialize to empty slice to avoid null in state
	state.Destinations = []DestinationModel{}

	for _, dest := range destinations {
		state.Destinations = append(state.Destinations, DestinationModel{
			ID:             types.StringValue(dest.DestinationID),
			Name:           types.StringValue(dest.Name),
			Bucket:         types.StringValue(dest.Bucket),
			Region:         types.StringValue(dest.Region),
			Endpoint:       types.StringValue(dest.Endpoint),
			OrganizationID: types.StringValue(dest.OrganizationID),
			CreatedAt:      types.StringValue(dest.CreatedAt.Format("2006-01-02T15:04:05Z")),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
