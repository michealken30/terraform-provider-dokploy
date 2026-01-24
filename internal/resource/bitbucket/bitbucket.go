package bitbucket

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
	_ resource.Resource                = &BitbucketResource{}
	_ resource.ResourceWithConfigure   = &BitbucketResource{}
	_ resource.ResourceWithImportState = &BitbucketResource{}
)

func NewResource() resource.Resource {
	return &BitbucketResource{}
}

type BitbucketResource struct {
	client *client.Client
}

type BitbucketResourceModel struct {
	ID                     types.String `tfsdk:"id"`
	GitProviderID          types.String `tfsdk:"git_provider_id"`
	Name                   types.String `tfsdk:"name"`
	AuthID                 types.String `tfsdk:"auth_id"`
	BitbucketUsername      types.String `tfsdk:"bitbucket_username"`
	BitbucketEmail         types.String `tfsdk:"bitbucket_email"`
	AppPassword            types.String `tfsdk:"app_password"`
	ApiToken               types.String `tfsdk:"api_token"`
	BitbucketWorkspaceName types.String `tfsdk:"bitbucket_workspace_name"`
	OrganizationID         types.String `tfsdk:"organization_id"`
}

func (r *BitbucketResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bitbucket"
}

func (r *BitbucketResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy Bitbucket provider integration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the Bitbucket provider (bitbucketId).",
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
				Description: "The name of the Bitbucket provider.",
				Required:    true,
			},
			"auth_id": schema.StringAttribute{
				Description: "The authentication ID for the provider.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"bitbucket_username": schema.StringAttribute{
				Description: "The Bitbucket username.",
				Optional:    true,
			},
			"bitbucket_email": schema.StringAttribute{
				Description: "The Bitbucket account email.",
				Optional:    true,
			},
			"app_password": schema.StringAttribute{
				Description: "The Bitbucket app password for authentication.",
				Optional:    true,
				Sensitive:   true,
			},
			"api_token": schema.StringAttribute{
				Description: "The Bitbucket API token.",
				Optional:    true,
				Sensitive:   true,
			},
			"bitbucket_workspace_name": schema.StringAttribute{
				Description: "The Bitbucket workspace name to restrict repository access.",
				Optional:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID to associate this provider with.",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *BitbucketResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BitbucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BitbucketResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateBitbucketRequest{
		Name:   plan.Name.ValueString(),
		AuthID: plan.AuthID.ValueString(),
	}

	if !plan.BitbucketUsername.IsNull() && !plan.BitbucketUsername.IsUnknown() {
		v := plan.BitbucketUsername.ValueString()
		createReq.BitbucketUsername = &v
	}
	if !plan.BitbucketEmail.IsNull() && !plan.BitbucketEmail.IsUnknown() {
		v := plan.BitbucketEmail.ValueString()
		createReq.BitbucketEmail = &v
	}
	if !plan.AppPassword.IsNull() && !plan.AppPassword.IsUnknown() {
		v := plan.AppPassword.ValueString()
		createReq.AppPassword = &v
	}
	if !plan.ApiToken.IsNull() && !plan.ApiToken.IsUnknown() {
		v := plan.ApiToken.ValueString()
		createReq.ApiToken = &v
	}
	if !plan.BitbucketWorkspaceName.IsNull() && !plan.BitbucketWorkspaceName.IsUnknown() {
		v := plan.BitbucketWorkspaceName.ValueString()
		createReq.BitbucketWorkspaceName = &v
	}

	bitbucket, err := r.client.CreateBitbucket(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Bitbucket Provider", err.Error())
		return
	}

	plan.ID = types.StringValue(bitbucket.BitbucketID)
	plan.GitProviderID = types.StringValue(bitbucket.GitProviderID)

	if bitbucket.OrganizationID != nil {
		plan.OrganizationID = types.StringValue(*bitbucket.OrganizationID)
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BitbucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BitbucketResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bitbucket, err := r.client.GetBitbucket(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Bitbucket Provider", err.Error())
		return
	}

	state.ID = types.StringValue(bitbucket.BitbucketID)
	state.GitProviderID = types.StringValue(bitbucket.GitProviderID)
	state.Name = types.StringValue(bitbucket.Name)

	if bitbucket.BitbucketUsername != nil {
		state.BitbucketUsername = types.StringValue(*bitbucket.BitbucketUsername)
	}
	if bitbucket.BitbucketEmail != nil {
		state.BitbucketEmail = types.StringValue(*bitbucket.BitbucketEmail)
	}
	if bitbucket.AppPassword != nil {
		state.AppPassword = types.StringValue(*bitbucket.AppPassword)
	}
	if bitbucket.ApiToken != nil {
		state.ApiToken = types.StringValue(*bitbucket.ApiToken)
	}
	if bitbucket.BitbucketWorkspaceName != nil {
		state.BitbucketWorkspaceName = types.StringValue(*bitbucket.BitbucketWorkspaceName)
	}
	if bitbucket.OrganizationID != nil {
		state.OrganizationID = types.StringValue(*bitbucket.OrganizationID)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *BitbucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BitbucketResourceModel
	var state BitbucketResourceModel

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

	updateReq := client.UpdateBitbucketRequest{
		BitbucketID:   state.ID.ValueString(),
		GitProviderID: state.GitProviderID.ValueString(),
		Name:          plan.Name.ValueString(),
	}

	if !plan.BitbucketUsername.IsNull() && !plan.BitbucketUsername.IsUnknown() {
		v := plan.BitbucketUsername.ValueString()
		updateReq.BitbucketUsername = &v
	}
	if !plan.BitbucketEmail.IsNull() && !plan.BitbucketEmail.IsUnknown() {
		v := plan.BitbucketEmail.ValueString()
		updateReq.BitbucketEmail = &v
	}
	if !plan.AppPassword.IsNull() && !plan.AppPassword.IsUnknown() {
		v := plan.AppPassword.ValueString()
		updateReq.AppPassword = &v
	}
	if !plan.ApiToken.IsNull() && !plan.ApiToken.IsUnknown() {
		v := plan.ApiToken.ValueString()
		updateReq.ApiToken = &v
	}
	if !plan.BitbucketWorkspaceName.IsNull() && !plan.BitbucketWorkspaceName.IsUnknown() {
		v := plan.BitbucketWorkspaceName.ValueString()
		updateReq.BitbucketWorkspaceName = &v
	}
	if !plan.OrganizationID.IsNull() && !plan.OrganizationID.IsUnknown() {
		v := plan.OrganizationID.ValueString()
		updateReq.OrganizationID = &v
	}

	if err := r.client.UpdateBitbucket(ctx, updateReq); err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Bitbucket Provider", err.Error())
		return
	}

	plan.ID = state.ID
	plan.GitProviderID = state.GitProviderID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BitbucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BitbucketResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteGitProvider(ctx, state.GitProviderID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Bitbucket Provider", err.Error())
		return
	}
}

func (r *BitbucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
