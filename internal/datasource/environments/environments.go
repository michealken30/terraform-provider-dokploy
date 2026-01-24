package environments

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &EnvironmentsDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &EnvironmentsDataSource{}
}

// EnvironmentsDataSource is the data source implementation.
type EnvironmentsDataSource struct {
	client *client.Client
}

// EnvironmentsDataSourceModel describes the data source data model.
type EnvironmentsDataSourceModel struct {
	ProjectID    types.String       `tfsdk:"project_id"`
	Environments []EnvironmentModel `tfsdk:"environments"`
}

// EnvironmentModel describes a single environment with its services.
type EnvironmentModel struct {
	ID           types.String       `tfsdk:"id"`
	Name         types.String       `tfsdk:"name"`
	Description  types.String       `tfsdk:"description"`
	IsDefault    types.Bool         `tfsdk:"is_default"`
	CreatedAt    types.String       `tfsdk:"created_at"`
	Applications []ApplicationModel `tfsdk:"applications"`
	Compose      []ComposeModel     `tfsdk:"compose"`
	Postgres     []PostgresModel    `tfsdk:"postgres"`
	MySQL        []MySQLModel       `tfsdk:"mysql"`
	Redis        []RedisModel       `tfsdk:"redis"`
}

// ApplicationModel describes an application summary.
type ApplicationModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Status   types.String `tfsdk:"status"`
	ServerID types.String `tfsdk:"server_id"`
}

// ComposeModel describes a compose service summary.
type ComposeModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Status   types.String `tfsdk:"status"`
	ServerID types.String `tfsdk:"server_id"`
}

// PostgresModel describes a Postgres database summary.
type PostgresModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Status   types.String `tfsdk:"status"`
	ServerID types.String `tfsdk:"server_id"`
}

// MySQLModel describes a MySQL database summary.
type MySQLModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Status   types.String `tfsdk:"status"`
	ServerID types.String `tfsdk:"server_id"`
}

// RedisModel describes a Redis database summary.
type RedisModel struct {
	ID       types.String `tfsdk:"id"`
	Name     types.String `tfsdk:"name"`
	Status   types.String `tfsdk:"status"`
	ServerID types.String `tfsdk:"server_id"`
}

func (d *EnvironmentsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_environments"
}

func (d *EnvironmentsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	serviceAttributes := map[string]schema.Attribute{
		"id": schema.StringAttribute{
			Description: "The unique identifier.",
			Computed:    true,
		},
		"name": schema.StringAttribute{
			Description: "The name.",
			Computed:    true,
		},
		"status": schema.StringAttribute{
			Description: "The current status.",
			Computed:    true,
		},
		"server_id": schema.StringAttribute{
			Description: "The server ID where deployed.",
			Computed:    true,
		},
	}

	resp.Schema = schema.Schema{
		Description: "Fetches all environments for a project from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Description: "The project ID to fetch environments for.",
				Required:    true,
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
						"created_at": schema.StringAttribute{
							Description: "The creation timestamp of the environment.",
							Computed:    true,
						},
						"applications": schema.ListNestedAttribute{
							Description: "List of applications in the environment.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: serviceAttributes,
							},
						},
						"compose": schema.ListNestedAttribute{
							Description: "List of compose services in the environment.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: serviceAttributes,
							},
						},
						"postgres": schema.ListNestedAttribute{
							Description: "List of Postgres databases in the environment.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: serviceAttributes,
							},
						},
						"mysql": schema.ListNestedAttribute{
							Description: "List of MySQL databases in the environment.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: serviceAttributes,
							},
						},
						"redis": schema.ListNestedAttribute{
							Description: "List of Redis databases in the environment.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: serviceAttributes,
							},
						},
					},
				},
			},
		},
	}
}

func (d *EnvironmentsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *EnvironmentsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config EnvironmentsDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	environments, err := d.client.GetEnvironmentsByProjectID(ctx, config.ProjectID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Environments",
			err.Error(),
		)
		return
	}

	state := EnvironmentsDataSourceModel{
		ProjectID:    config.ProjectID,
		Environments: []EnvironmentModel{},
	}

	for _, env := range environments {
		envModel := EnvironmentModel{
			ID:           types.StringValue(env.EnvironmentID),
			Name:         types.StringValue(env.Name),
			Description:  types.StringValue(env.Description),
			IsDefault:    types.BoolValue(env.IsDefault),
			CreatedAt:    types.StringValue(env.CreatedAt.Format("2006-01-02T15:04:05Z")),
			Applications: []ApplicationModel{},
			Compose:      []ComposeModel{},
			Postgres:     []PostgresModel{},
			MySQL:        []MySQLModel{},
			Redis:        []RedisModel{},
		}

		for _, app := range env.Applications {
			serverID := ""
			if app.ServerID != nil {
				serverID = *app.ServerID
			}
			envModel.Applications = append(envModel.Applications, ApplicationModel{
				ID:       types.StringValue(app.ApplicationID),
				Name:     types.StringValue(app.Name),
				Status:   types.StringValue(app.ApplicationStatus),
				ServerID: types.StringValue(serverID),
			})
		}

		for _, comp := range env.Compose {
			serverID := ""
			if comp.ServerID != nil {
				serverID = *comp.ServerID
			}
			envModel.Compose = append(envModel.Compose, ComposeModel{
				ID:       types.StringValue(comp.ComposeID),
				Name:     types.StringValue(comp.Name),
				Status:   types.StringValue(comp.ComposeStatus),
				ServerID: types.StringValue(serverID),
			})
		}

		for _, pg := range env.Postgres {
			serverID := ""
			if pg.ServerID != nil {
				serverID = *pg.ServerID
			}
			envModel.Postgres = append(envModel.Postgres, PostgresModel{
				ID:       types.StringValue(pg.PostgresID),
				Name:     types.StringValue(pg.Name),
				Status:   types.StringValue(pg.ApplicationStatus),
				ServerID: types.StringValue(serverID),
			})
		}

		for _, mysql := range env.MySQL {
			serverID := ""
			if mysql.ServerID != nil {
				serverID = *mysql.ServerID
			}
			envModel.MySQL = append(envModel.MySQL, MySQLModel{
				ID:       types.StringValue(mysql.MySQLID),
				Name:     types.StringValue(mysql.Name),
				Status:   types.StringValue(mysql.ApplicationStatus),
				ServerID: types.StringValue(serverID),
			})
		}

		for _, redis := range env.Redis {
			serverID := ""
			if redis.ServerID != nil {
				serverID = *redis.ServerID
			}
			envModel.Redis = append(envModel.Redis, RedisModel{
				ID:       types.StringValue(redis.RedisID),
				Name:     types.StringValue(redis.Name),
				Status:   types.StringValue(redis.ApplicationStatus),
				ServerID: types.StringValue(serverID),
			})
		}

		state.Environments = append(state.Environments, envModel)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
