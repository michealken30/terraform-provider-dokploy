package gitea

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &GiteaResource{}
	_ resource.ResourceWithConfigure   = &GiteaResource{}
	_ resource.ResourceWithImportState = &GiteaResource{}
)

func NewResource() resource.Resource {
	return &GiteaResource{}
}

type GiteaResource struct {
	client *client.Client
}

type GiteaResourceModel struct {
	ID               types.String `tfsdk:"id"`
	GitProviderID    types.String `tfsdk:"git_provider_id"`
	Name             types.String `tfsdk:"name"`
	GiteaURL         types.String `tfsdk:"gitea_url"`
	RedirectURI      types.String `tfsdk:"redirect_uri"`
	ClientID         types.String `tfsdk:"client_id"`
	ClientSecret     types.String `tfsdk:"client_secret"`
	AccessToken      types.String `tfsdk:"access_token"`
	RefreshToken     types.String `tfsdk:"refresh_token"`
	ExpiresAt        types.Int64  `tfsdk:"expires_at"`
	Scopes           types.String `tfsdk:"scopes"`
	GiteaUsername    types.String `tfsdk:"gitea_username"`
	OrganizationName types.String `tfsdk:"organization_name"`
}

func (r *GiteaResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitea"
}

func (r *GiteaResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy Gitea provider integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Gitea provider (giteaId).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"git_provider_id": schema.StringAttribute{
				Description: "The git provider identifier used for deletion.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the Gitea provider.",
				Required:    true,
			},
			"gitea_url": schema.StringAttribute{
				Description: "The Gitea instance URL (e.g., https://gitea.example.com).",
				Required:    true,
			},
			"redirect_uri": schema.StringAttribute{
				Description: "The OAuth redirect URI.",
				Optional:    true,
			},
			"client_id": schema.StringAttribute{
				Description: "The Gitea OAuth application client ID.",
				Optional:    true,
			},
			"client_secret": schema.StringAttribute{
				Description: "The Gitea OAuth application client secret.",
				Optional:    true,
				Sensitive:   true,
			},
			"access_token": schema.StringAttribute{
				Description: "The OAuth access token.",
				Optional:    true,
				Sensitive:   true,
			},
			"refresh_token": schema.StringAttribute{
				Description: "The OAuth refresh token.",
				Optional:    true,
				Sensitive:   true,
			},
			"expires_at": schema.Int64Attribute{
				Description: "The token expiration timestamp.",
				Optional:    true,
				Computed:    true,
			},
			"scopes": schema.StringAttribute{
				Description: "The OAuth scopes requested.",
				Optional:    true,
			},
			"gitea_username": schema.StringAttribute{
				Description: "The Gitea username.",
				Optional:    true,
			},
			"organization_name": schema.StringAttribute{
				Description: "The Gitea organization name to restrict repository access.",
				Optional:    true,
			},
		},
	}
}

func (r *GiteaResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = c
}

func (r *GiteaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GiteaResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateGiteaRequest{
		Name:     plan.Name.ValueString(),
		GiteaURL: plan.GiteaURL.ValueString(),
	}

	if !plan.RedirectURI.IsNull() && !plan.RedirectURI.IsUnknown() {
		v := plan.RedirectURI.ValueString()
		createReq.RedirectURI = &v
	}
	if !plan.ClientID.IsNull() && !plan.ClientID.IsUnknown() {
		v := plan.ClientID.ValueString()
		createReq.ClientID = &v
	}
	if !plan.ClientSecret.IsNull() && !plan.ClientSecret.IsUnknown() {
		v := plan.ClientSecret.ValueString()
		createReq.ClientSecret = &v
	}
	if !plan.AccessToken.IsNull() && !plan.AccessToken.IsUnknown() {
		v := plan.AccessToken.ValueString()
		createReq.AccessToken = &v
	}
	if !plan.RefreshToken.IsNull() && !plan.RefreshToken.IsUnknown() {
		v := plan.RefreshToken.ValueString()
		createReq.RefreshToken = &v
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueInt64()
		createReq.ExpiresAt = &v
	}
	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		v := plan.Scopes.ValueString()
		createReq.Scopes = &v
	}
	if !plan.GiteaUsername.IsNull() && !plan.GiteaUsername.IsUnknown() {
		v := plan.GiteaUsername.ValueString()
		createReq.GiteaUsername = &v
	}
	if !plan.OrganizationName.IsNull() && !plan.OrganizationName.IsUnknown() {
		v := plan.OrganizationName.ValueString()
		createReq.OrganizationName = &v
	}

	gitea, err := r.client.CreateGitea(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Gitea Provider", err.Error())
		return
	}

	plan.ID = types.StringValue(gitea.GiteaID)
	plan.GitProviderID = types.StringValue(gitea.GitProviderID)

	if gitea.ExpiresAt != nil {
		plan.ExpiresAt = types.Int64Value(*gitea.ExpiresAt)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GiteaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GiteaResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	gitea, err := r.client.GetGitea(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Gitea Provider", err.Error())
		return
	}

	state.ID = types.StringValue(gitea.GiteaID)
	state.GitProviderID = types.StringValue(gitea.GitProviderID)
	state.Name = types.StringValue(gitea.Name)
	state.GiteaURL = types.StringValue(gitea.GiteaURL)

	if gitea.RedirectURI != nil {
		state.RedirectURI = types.StringValue(*gitea.RedirectURI)
	}
	if gitea.ClientID != nil {
		state.ClientID = types.StringValue(*gitea.ClientID)
	}
	if gitea.ClientSecret != nil {
		state.ClientSecret = types.StringValue(*gitea.ClientSecret)
	}
	if gitea.AccessToken != nil {
		state.AccessToken = types.StringValue(*gitea.AccessToken)
	}
	if gitea.RefreshToken != nil {
		state.RefreshToken = types.StringValue(*gitea.RefreshToken)
	}
	if gitea.ExpiresAt != nil {
		state.ExpiresAt = types.Int64Value(*gitea.ExpiresAt)
	}
	if gitea.Scopes != nil {
		state.Scopes = types.StringValue(*gitea.Scopes)
	}
	if gitea.GiteaUsername != nil {
		state.GiteaUsername = types.StringValue(*gitea.GiteaUsername)
	}
	if gitea.OrganizationName != nil {
		state.OrganizationName = types.StringValue(*gitea.OrganizationName)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GiteaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GiteaResourceModel
	var state GiteaResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateReq := client.UpdateGiteaRequest{
		GiteaID:       state.ID.ValueString(),
		GitProviderID: state.GitProviderID.ValueString(),
		Name:          plan.Name.ValueString(),
		GiteaURL:      plan.GiteaURL.ValueString(),
	}

	if !plan.RedirectURI.IsNull() && !plan.RedirectURI.IsUnknown() {
		v := plan.RedirectURI.ValueString()
		updateReq.RedirectURI = &v
	}
	if !plan.ClientID.IsNull() && !plan.ClientID.IsUnknown() {
		v := plan.ClientID.ValueString()
		updateReq.ClientID = &v
	}
	if !plan.ClientSecret.IsNull() && !plan.ClientSecret.IsUnknown() {
		v := plan.ClientSecret.ValueString()
		updateReq.ClientSecret = &v
	}
	if !plan.AccessToken.IsNull() && !plan.AccessToken.IsUnknown() {
		v := plan.AccessToken.ValueString()
		updateReq.AccessToken = &v
	}
	if !plan.RefreshToken.IsNull() && !plan.RefreshToken.IsUnknown() {
		v := plan.RefreshToken.ValueString()
		updateReq.RefreshToken = &v
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueInt64()
		updateReq.ExpiresAt = &v
	}
	if !plan.Scopes.IsNull() && !plan.Scopes.IsUnknown() {
		v := plan.Scopes.ValueString()
		updateReq.Scopes = &v
	}
	if !plan.GiteaUsername.IsNull() && !plan.GiteaUsername.IsUnknown() {
		v := plan.GiteaUsername.ValueString()
		updateReq.GiteaUsername = &v
	}
	if !plan.OrganizationName.IsNull() && !plan.OrganizationName.IsUnknown() {
		v := plan.OrganizationName.ValueString()
		updateReq.OrganizationName = &v
	}

	if err := r.client.UpdateGitea(ctx, updateReq); err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Gitea Provider", err.Error())
		return
	}

	plan.ID = state.ID
	plan.GitProviderID = state.GitProviderID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GiteaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GiteaResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGitProvider(ctx, state.GitProviderID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Gitea Provider", err.Error())
		return
	}
}

func (r *GiteaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
