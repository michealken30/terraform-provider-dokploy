package security

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
	_ resource.Resource                = &SecurityResource{}
	_ resource.ResourceWithConfigure   = &SecurityResource{}
	_ resource.ResourceWithImportState = &SecurityResource{}
)

func NewResource() resource.Resource {
	return &SecurityResource{}
}

type SecurityResource struct {
	client *client.Client
}

type SecurityResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Username      types.String `tfsdk:"username"`
	Password      types.String `tfsdk:"password"`
	ApplicationID types.String `tfsdk:"application_id"`
}

func (r *SecurityResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_security"
}

func (r *SecurityResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy security entry (basic auth) for an application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the security entry.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "The username for basic authentication.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password for basic authentication.",
				Required:    true,
				Sensitive:   true,
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this security entry belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SecurityResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SecurityResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SecurityResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateSecurityRequest{
		ApplicationID: plan.ApplicationID.ValueString(),
		Username:      plan.Username.ValueString(),
		Password:      plan.Password.ValueString(),
	}

	createResp, err := r.client.CreateSecurity(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Security", "Could not create security entry: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.SecurityID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SecurityResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	security, err := r.client.GetSecurity(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Security", "Could not read security ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(security.SecurityID)
	state.Username = types.StringValue(security.Username)
	state.Password = types.StringValue(security.Password)
	state.ApplicationID = types.StringValue(security.ApplicationID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SecurityResourceModel
	var state SecurityResourceModel

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

	updateReq := client.UpdateSecurityRequest{
		SecurityID: state.ID.ValueString(),
		Username:   plan.Username.ValueString(),
		Password:   plan.Password.ValueString(),
	}

	err := r.client.UpdateSecurity(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Security", "Could not update security entry: "+err.Error())
		return
	}

	// Preserve immutable fields from state
	plan.ID = state.ID
	plan.ApplicationID = state.ApplicationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SecurityResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SecurityResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSecurity(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Security", "Could not delete security entry: "+err.Error())
		return
	}
}

func (r *SecurityResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
