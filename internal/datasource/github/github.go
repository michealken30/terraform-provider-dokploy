package github

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &GithubDataSource{}

func NewDataSource() datasource.DataSource {
	return &GithubDataSource{}
}

type GithubDataSource struct {
	client *client.Client
}

type GithubDataSourceModel struct {
	ID            types.String `tfsdk:"id"`
	GithubID      types.String `tfsdk:"github_id"`
	GitProviderID types.String `tfsdk:"git_provider_id"`
	GithubAppName types.String `tfsdk:"github_app_name"`
	GithubAppID   types.Int64  `tfsdk:"github_app_id"`
}

func (d *GithubDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_github"
}

func (d *GithubDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a GitHub OAuth connection from Dokploy. GitHub providers can only be created via OAuth flow, but can be read via this data source.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier.",
				Computed:    true,
			},
			"github_id": schema.StringAttribute{
				Description: "The GitHub provider ID to look up.",
				Required:    true,
			},
			"git_provider_id": schema.StringAttribute{
				Description: "The git provider ID.",
				Computed:    true,
			},
			"github_app_name": schema.StringAttribute{
				Description: "The GitHub App name.",
				Computed:    true,
			},
			"github_app_id": schema.Int64Attribute{
				Description: "The GitHub App ID.",
				Computed:    true,
			},
		},
	}
}

func (d *GithubDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GithubDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config GithubDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	github, err := d.client.GetGithub(ctx, config.GithubID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy GitHub Provider", err.Error())
		return
	}

	state := GithubDataSourceModel{
		ID:            types.StringValue(github.GithubID),
		GithubID:      types.StringValue(github.GithubID),
		GitProviderID: types.StringValue(github.GitProviderID),
	}

	if github.GithubAppName != nil {
		state.GithubAppName = types.StringValue(*github.GithubAppName)
	} else {
		state.GithubAppName = types.StringNull()
	}

	if github.GithubAppID != nil {
		state.GithubAppID = types.Int64Value(*github.GithubAppID)
	} else {
		state.GithubAppID = types.Int64Null()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
