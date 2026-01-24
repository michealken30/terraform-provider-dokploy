package organization

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &OrganizationDataSource{}

func NewDataSource() datasource.DataSource {
	return &OrganizationDataSource{}
}

type OrganizationDataSource struct {
	client *client.Client
}

type OrganizationDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	OrganizationID types.String `tfsdk:"organization_id"`
	Name           types.String `tfsdk:"name"`
	Logo           types.String `tfsdk:"logo"`
}

func (d *OrganizationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (d *OrganizationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single organization from Dokploy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the organization.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID to look up. Mutually exclusive with name.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The organization name to look up. Mutually exclusive with organization_id.",
				Optional:    true,
				Computed:    true,
			},
			"logo": schema.StringAttribute{
				Description: "The logo URL of the organization.",
				Computed:    true,
			},
		},
	}
}

func (d *OrganizationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *OrganizationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config OrganizationDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.OrganizationID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either organization_id or name must be specified.")
		return
	}

	if !config.OrganizationID.IsNull() && !config.Name.IsNull() {
		resp.Diagnostics.AddError("Conflicting Attributes", "Only one of organization_id or name can be specified, not both.")
		return
	}

	var org *client.Organization
	var err error

	if !config.OrganizationID.IsNull() {
		org, err = d.client.GetOrganization(ctx, config.OrganizationID.ValueString())
	} else {
		// Look up by name
		orgs, fetchErr := d.client.GetOrganizations(ctx)
		if fetchErr != nil {
			resp.Diagnostics.AddError("Unable to Read Dokploy Organizations", fetchErr.Error())
			return
		}
		name := config.Name.ValueString()
		for i := range orgs {
			if orgs[i].Name == name {
				org = &orgs[i]
				break
			}
		}
		if org == nil {
			err = fmt.Errorf("organization with name %s not found", name)
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy Organization", err.Error())
		return
	}

	state := OrganizationDataSourceModel{
		ID:             types.StringValue(org.OrganizationID),
		OrganizationID: types.StringValue(org.OrganizationID),
		Name:           types.StringValue(org.Name),
		Logo:           types.StringValue(org.Logo),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
