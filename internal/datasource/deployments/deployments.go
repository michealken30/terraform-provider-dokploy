package deployments

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &DeploymentsDataSource{}

func NewDataSource() datasource.DataSource {
	return &DeploymentsDataSource{}
}

type DeploymentsDataSource struct {
	client *client.Client
}

type DeploymentsDataSourceModel struct {
	ApplicationID types.String      `tfsdk:"application_id"`
	ComposeID     types.String      `tfsdk:"compose_id"`
	Deployments   []DeploymentModel `tfsdk:"deployments"`
}

type DeploymentModel struct {
	ID            types.String `tfsdk:"id"`
	Title         types.String `tfsdk:"title"`
	Status        types.String `tfsdk:"status"`
	Description   types.String `tfsdk:"description"`
	ApplicationID types.String `tfsdk:"application_id"`
	ComposeID     types.String `tfsdk:"compose_id"`
	CreatedAt     types.String `tfsdk:"created_at"`
}

func (d *DeploymentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_deployments"
}

func (d *DeploymentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches deployment history from Dokploy for an application or compose service.",
		Attributes: map[string]schema.Attribute{
			"application_id": schema.StringAttribute{
				Description: "The application ID to fetch deployments for. Mutually exclusive with compose_id.",
				Optional:    true,
			},
			"compose_id": schema.StringAttribute{
				Description: "The compose ID to fetch deployments for. Mutually exclusive with application_id.",
				Optional:    true,
			},
			"deployments": schema.ListNestedAttribute{
				Description: "List of deployments.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the deployment.",
							Computed:    true,
						},
						"title": schema.StringAttribute{
							Description: "The title of the deployment.",
							Computed:    true,
						},
						"status": schema.StringAttribute{
							Description: "The status of the deployment.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the deployment.",
							Computed:    true,
						},
						"application_id": schema.StringAttribute{
							Description: "The application ID.",
							Computed:    true,
						},
						"compose_id": schema.StringAttribute{
							Description: "The compose ID.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The creation timestamp.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *DeploymentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *DeploymentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config DeploymentsDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.ApplicationID.IsNull() && config.ComposeID.IsNull() {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either application_id or compose_id must be specified.")
		return
	}

	if !config.ApplicationID.IsNull() && !config.ComposeID.IsNull() {
		resp.Diagnostics.AddError("Conflicting Attributes", "Only one of application_id or compose_id can be specified, not both.")
		return
	}

	var deployments []client.Deployment
	var err error

	if !config.ApplicationID.IsNull() {
		deployments, err = d.client.GetDeployments(ctx, config.ApplicationID.ValueString(), "application")
	} else {
		deployments, err = d.client.GetDeployments(ctx, config.ComposeID.ValueString(), "compose")
	}

	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy Deployments", err.Error())
		return
	}

	state := DeploymentsDataSourceModel{
		ApplicationID: config.ApplicationID,
		ComposeID:     config.ComposeID,
		Deployments:   []DeploymentModel{},
	}

	for _, dep := range deployments {
		model := DeploymentModel{
			ID:     types.StringValue(dep.DeploymentID),
			Status: types.StringValue(dep.Status),
		}

		if dep.Title != nil {
			model.Title = types.StringValue(*dep.Title)
		} else {
			model.Title = types.StringNull()
		}

		if dep.Description != nil {
			model.Description = types.StringValue(*dep.Description)
		} else {
			model.Description = types.StringNull()
		}

		if dep.ApplicationID != nil {
			model.ApplicationID = types.StringValue(*dep.ApplicationID)
		} else {
			model.ApplicationID = types.StringNull()
		}

		if dep.ComposeID != nil {
			model.ComposeID = types.StringValue(*dep.ComposeID)
		} else {
			model.ComposeID = types.StringNull()
		}

		if dep.CreatedAt != nil {
			model.CreatedAt = types.StringValue(*dep.CreatedAt)
		} else {
			model.CreatedAt = types.StringNull()
		}

		state.Deployments = append(state.Deployments, model)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
