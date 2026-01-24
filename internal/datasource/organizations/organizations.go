package organizations

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &OrganizationsDataSource{}

func NewDataSource() datasource.DataSource {
	return &OrganizationsDataSource{}
}

type OrganizationsDataSource struct {
	client *client.Client
}

type OrganizationsDataSourceModel struct {
	Organizations []OrganizationModel `tfsdk:"organizations"`
}

type OrganizationModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Logo types.String `tfsdk:"logo"`
}

func (d *OrganizationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organizations"
}

func (d *OrganizationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all organizations from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"organizations": schema.ListNestedAttribute{
				Description: "List of all organizations.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the organization.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the organization.",
							Computed:    true,
						},
						"logo": schema.StringAttribute{
							Description: "The logo URL of the organization.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *OrganizationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = c
}

func (d *OrganizationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state OrganizationsDataSourceModel

	orgs, err := d.client.GetOrganizations(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy Organizations", err.Error())
		return
	}

	state.Organizations = []OrganizationModel{}
	for _, org := range orgs {
		state.Organizations = append(state.Organizations, OrganizationModel{
			ID:   types.StringValue(org.OrganizationID),
			Name: types.StringValue(org.Name),
			Logo: types.StringValue(org.Logo),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
