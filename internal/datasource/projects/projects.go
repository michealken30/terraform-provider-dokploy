package projects

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &ProjectsDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &ProjectsDataSource{}
}

// ProjectsDataSource is the data source implementation.
type ProjectsDataSource struct {
	client *client.Client
}

// ProjectsDataSourceModel describes the data source data model.
type ProjectsDataSourceModel struct {
	Projects []ProjectModel `tfsdk:"projects"`
}

// ProjectModel describes a single project.
type ProjectModel struct {
	ID             types.String       `tfsdk:"id"`
	Name           types.String       `tfsdk:"name"`
	Description    types.String       `tfsdk:"description"`
	OrganizationID types.String       `tfsdk:"organization_id"`
	CreatedAt      types.String       `tfsdk:"created_at"`
	Environments   []EnvironmentModel `tfsdk:"environments"`
}

// EnvironmentModel describes a single environment (summary).
type EnvironmentModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IsDefault   types.Bool   `tfsdk:"is_default"`
}

func (d *ProjectsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

func (d *ProjectsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all projects from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"projects": schema.ListNestedAttribute{
				Description: "List of all projects.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the project.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the project.",
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
				},
			},
		},
	}
}

func (d *ProjectsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProjectsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state ProjectsDataSourceModel

	projects, err := d.client.GetProjects(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Projects",
			err.Error(),
		)
		return
	}

	for _, project := range projects {
		projectModel := ProjectModel{
			ID:             types.StringValue(project.ProjectID),
			Name:           types.StringValue(project.Name),
			Description:    types.StringValue(project.Description),
			OrganizationID: types.StringValue(project.OrganizationID),
			CreatedAt:      types.StringValue(project.CreatedAt.Format("2006-01-02T15:04:05Z")),
			Environments:   []EnvironmentModel{},
		}

		for _, env := range project.Environments {
			projectModel.Environments = append(projectModel.Environments, EnvironmentModel{
				ID:          types.StringValue(env.EnvironmentID),
				Name:        types.StringValue(env.Name),
				Description: types.StringValue(env.Description),
				IsDefault:   types.BoolValue(env.IsDefault),
			})
		}

		state.Projects = append(state.Projects, projectModel)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
