package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &ProjectDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &ProjectDataSource{}
}

// ProjectDataSource is the data source implementation.
type ProjectDataSource struct {
	client *client.Client
}

// ProjectDataSourceModel describes the data source data model.
type ProjectDataSourceModel struct {
	ID             types.String       `tfsdk:"id"`
	ProjectID      types.String       `tfsdk:"project_id"`
	Name           types.String       `tfsdk:"name"`
	Description    types.String       `tfsdk:"description"`
	OrganizationID types.String       `tfsdk:"organization_id"`
	CreatedAt      types.String       `tfsdk:"created_at"`
	Environments   []EnvironmentModel `tfsdk:"environments"`
}

// EnvironmentModel describes a single environment.
type EnvironmentModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func (d *ProjectDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (d *ProjectDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single project from Dokploy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the project (same as project_id).",
				Computed:    true,
			},
			"project_id": schema.StringAttribute{
				Description: "The project ID to look up. Mutually exclusive with name.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The project name to look up. Mutually exclusive with project_id.",
				Optional:    true,
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the project.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID the project belongs to.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp of the project.",
				Computed:    true,
			},
			"environments": schema.ListNestedAttribute{
				Description: "List of environments in the project.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the environment.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the environment.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "The description of the environment.",
							Computed:    true,
						},
						"is_default": schema.BoolAttribute{
							Description: "Whether this is the default environment.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ProjectDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ProjectDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either project_id or name is provided
	if config.ProjectID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Missing Required Attribute",
			"Either project_id or name must be specified.",
		)
		return
	}

	if !config.ProjectID.IsNull() && !config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Conflicting Attributes",
			"Only one of project_id or name can be specified, not both.",
		)
		return
	}

	var project *client.Project
	var err error

	if !config.ProjectID.IsNull() {
		project, err = d.client.GetProject(ctx, config.ProjectID.ValueString())
	} else {
		project, err = d.client.GetProjectByName(ctx, config.Name.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Project",
			err.Error(),
		)
		return
	}

	state := ProjectDataSourceModel{
		ID:             types.StringValue(project.ProjectID),
		ProjectID:      types.StringValue(project.ProjectID),
		Name:           types.StringValue(project.Name),
		Description:    types.StringValue(project.Description),
		OrganizationID: types.StringValue(project.OrganizationID),
		CreatedAt:      types.StringValue(project.CreatedAt.Format("2006-01-02T15:04:05Z")),
		Environments:   []EnvironmentModel{},
	}

	for _, env := range project.Environments {
		state.Environments = append(state.Environments, EnvironmentModel{
			ID:          types.StringValue(env.EnvironmentID),
			Name:        types.StringValue(env.Name),
			Description: types.StringValue(env.Description),
			IsDefault:   types.BoolValue(env.IsDefault),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
