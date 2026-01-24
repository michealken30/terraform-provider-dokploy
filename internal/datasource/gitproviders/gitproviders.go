package gitproviders

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &GitProvidersDataSource{}

func NewDataSource() datasource.DataSource {
	return &GitProvidersDataSource{}
}

type GitProvidersDataSource struct {
	client *client.Client
}

type GitProvidersDataSourceModel struct {
	GitProviders []GitProviderModel `tfsdk:"git_providers"`
}

type GitProviderModel struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	ProviderType types.String `tfsdk:"provider_type"`
	// Type-specific IDs
	GitlabID    types.String `tfsdk:"gitlab_id"`
	BitbucketID types.String `tfsdk:"bitbucket_id"`
	GiteaID     types.String `tfsdk:"gitea_id"`
	GithubID    types.String `tfsdk:"github_id"`
}

func (d *GitProvidersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git_providers"
}

func (d *GitProvidersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all git providers from Dokploy (GitLab, Bitbucket, Gitea, GitHub).",
		Attributes: map[string]schema.Attribute{
			"git_providers": schema.ListNestedAttribute{
				Description: "List of all git providers.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The git provider ID.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the git provider.",
							Computed:    true,
						},
						"provider_type": schema.StringAttribute{
							Description: "The type of git provider (gitlab, bitbucket, gitea, github).",
							Computed:    true,
						},
						"gitlab_id": schema.StringAttribute{
							Description: "The GitLab-specific ID (if provider_type is gitlab).",
							Computed:    true,
						},
						"bitbucket_id": schema.StringAttribute{
							Description: "The Bitbucket-specific ID (if provider_type is bitbucket).",
							Computed:    true,
						},
						"gitea_id": schema.StringAttribute{
							Description: "The Gitea-specific ID (if provider_type is gitea).",
							Computed:    true,
						},
						"github_id": schema.StringAttribute{
							Description: "The GitHub-specific ID (if provider_type is github).",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *GitProvidersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *GitProvidersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state GitProvidersDataSourceModel

	providers, err := d.client.GetGitProviders(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy Git Providers", err.Error())
		return
	}

	state.GitProviders = []GitProviderModel{}
	for _, p := range providers {
		model := GitProviderModel{
			ID:           types.StringValue(p.GitProviderID),
			Name:         types.StringValue(p.Name),
			ProviderType: types.StringValue(p.ProviderType),
			GitlabID:     types.StringNull(),
			BitbucketID:  types.StringNull(),
			GiteaID:      types.StringNull(),
			GithubID:     types.StringNull(),
		}

		if p.GitlabID != nil {
			model.GitlabID = types.StringValue(*p.GitlabID)
		}
		if p.BitbucketID != nil {
			model.BitbucketID = types.StringValue(*p.BitbucketID)
		}
		if p.GiteaID != nil {
			model.GiteaID = types.StringValue(*p.GiteaID)
		}
		if p.GithubID != nil {
			model.GithubID = types.StringValue(*p.GithubID)
		}

		state.GitProviders = append(state.GitProviders, model)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
