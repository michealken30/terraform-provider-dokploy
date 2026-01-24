package gitlab

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &GitlabResource{}
	_ resource.ResourceWithConfigure   = &GitlabResource{}
	_ resource.ResourceWithImportState = &GitlabResource{}
)

func NewResource() resource.Resource {
	return &GitlabResource{}
}

type GitlabResource struct {
	client *client.Client
}

type GitlabResourceModel struct {
	ID            types.String `tfsdk:"id"`
	GitProviderID types.String `tfsdk:"git_provider_id"`
	Name          types.String `tfsdk:"name"`
	GitlabURL     types.String `tfsdk:"gitlab_url"`
	AuthID        types.String `tfsdk:"auth_id"`
	ApplicationID types.String `tfsdk:"application_id"`
	RedirectURI   types.String `tfsdk:"redirect_uri"`
	Secret        types.String `tfsdk:"secret"`
	AccessToken   types.String `tfsdk:"access_token"`
	RefreshToken  types.String `tfsdk:"refresh_token"`
	GroupName     types.String `tfsdk:"group_name"`
	ExpiresAt     types.Int64  `tfsdk:"expires_at"`
}

func (r *GitlabResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gitlab"
}

func (r *GitlabResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy GitLab provider integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the GitLab provider (gitlabId).",
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
				Description: "The name of the GitLab provider.",
				Required:    true,
			},
			"gitlab_url": schema.StringAttribute{
				Description: "The GitLab instance URL (e.g., https://gitlab.com).",
				Required:    true,
			},
			"auth_id": schema.StringAttribute{
				Description: "The authentication ID for the provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The GitLab OAuth application ID.",
				Optional:    true,
			},
			"redirect_uri": schema.StringAttribute{
				Description: "The OAuth redirect URI.",
				Optional:    true,
			},
			"secret": schema.StringAttribute{
				Description: "The GitLab OAuth application secret.",
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
			"group_name": schema.StringAttribute{
				Description: "The GitLab group name to restrict repository access.",
				Optional:    true,
			},
			"expires_at": schema.Int64Attribute{
				Description: "The token expiration timestamp.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *GitlabResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *GitlabResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan GitlabResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateGitlabRequest{
		Name:      plan.Name.ValueString(),
		GitlabURL: plan.GitlabURL.ValueString(),
		AuthID:    plan.AuthID.ValueString(),
	}

	if !plan.ApplicationID.IsNull() && !plan.ApplicationID.IsUnknown() {
		v := plan.ApplicationID.ValueString()
		createReq.ApplicationID = &v
	}
	if !plan.RedirectURI.IsNull() && !plan.RedirectURI.IsUnknown() {
		v := plan.RedirectURI.ValueString()
		createReq.RedirectURI = &v
	}
	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() {
		v := plan.Secret.ValueString()
		createReq.Secret = &v
	}
	if !plan.AccessToken.IsNull() && !plan.AccessToken.IsUnknown() {
		v := plan.AccessToken.ValueString()
		createReq.AccessToken = &v
	}
	if !plan.RefreshToken.IsNull() && !plan.RefreshToken.IsUnknown() {
		v := plan.RefreshToken.ValueString()
		createReq.RefreshToken = &v
	}
	if !plan.GroupName.IsNull() && !plan.GroupName.IsUnknown() {
		v := plan.GroupName.ValueString()
		createReq.GroupName = &v
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueInt64()
		createReq.ExpiresAt = &v
	}

	gitlab, err := r.client.CreateGitlab(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy GitLab Provider", err.Error())
		return
	}

	plan.ID = types.StringValue(gitlab.GitlabID)
	plan.GitProviderID = types.StringValue(gitlab.GitProviderID)

	if gitlab.ExpiresAt != nil {
		plan.ExpiresAt = types.Int64Value(*gitlab.ExpiresAt)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state GitlabResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	gitlab, err := r.client.GetGitlab(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy GitLab Provider", err.Error())
		return
	}

	state.ID = types.StringValue(gitlab.GitlabID)
	state.GitProviderID = types.StringValue(gitlab.GitProviderID)
	state.Name = types.StringValue(gitlab.Name)
	state.GitlabURL = types.StringValue(gitlab.GitlabURL)

	if gitlab.ApplicationID != nil {
		state.ApplicationID = types.StringValue(*gitlab.ApplicationID)
	}
	if gitlab.RedirectURI != nil {
		state.RedirectURI = types.StringValue(*gitlab.RedirectURI)
	}
	if gitlab.Secret != nil {
		state.Secret = types.StringValue(*gitlab.Secret)
	}
	if gitlab.AccessToken != nil {
		state.AccessToken = types.StringValue(*gitlab.AccessToken)
	}
	if gitlab.RefreshToken != nil {
		state.RefreshToken = types.StringValue(*gitlab.RefreshToken)
	}
	if gitlab.GroupName != nil {
		state.GroupName = types.StringValue(*gitlab.GroupName)
	}
	if gitlab.ExpiresAt != nil {
		state.ExpiresAt = types.Int64Value(*gitlab.ExpiresAt)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan GitlabResourceModel
	var state GitlabResourceModel

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

	updateReq := client.UpdateGitlabRequest{
		GitlabID:      state.ID.ValueString(),
		GitProviderID: state.GitProviderID.ValueString(),
		Name:          plan.Name.ValueString(),
		GitlabURL:     plan.GitlabURL.ValueString(),
	}

	if !plan.ApplicationID.IsNull() && !plan.ApplicationID.IsUnknown() {
		v := plan.ApplicationID.ValueString()
		updateReq.ApplicationID = &v
	}
	if !plan.RedirectURI.IsNull() && !plan.RedirectURI.IsUnknown() {
		v := plan.RedirectURI.ValueString()
		updateReq.RedirectURI = &v
	}
	if !plan.Secret.IsNull() && !plan.Secret.IsUnknown() {
		v := plan.Secret.ValueString()
		updateReq.Secret = &v
	}
	if !plan.AccessToken.IsNull() && !plan.AccessToken.IsUnknown() {
		v := plan.AccessToken.ValueString()
		updateReq.AccessToken = &v
	}
	if !plan.RefreshToken.IsNull() && !plan.RefreshToken.IsUnknown() {
		v := plan.RefreshToken.ValueString()
		updateReq.RefreshToken = &v
	}
	if !plan.GroupName.IsNull() && !plan.GroupName.IsUnknown() {
		v := plan.GroupName.ValueString()
		updateReq.GroupName = &v
	}
	if !plan.ExpiresAt.IsNull() && !plan.ExpiresAt.IsUnknown() {
		v := plan.ExpiresAt.ValueInt64()
		updateReq.ExpiresAt = &v
	}

	if err := r.client.UpdateGitlab(ctx, updateReq); err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy GitLab Provider", err.Error())
		return
	}

	plan.ID = state.ID
	plan.GitProviderID = state.GitProviderID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *GitlabResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state GitlabResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGitProvider(ctx, state.GitProviderID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy GitLab Provider", err.Error())
		return
	}
}

func (r *GitlabResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
