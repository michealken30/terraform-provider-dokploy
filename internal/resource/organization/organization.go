package organization

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
	_ resource.Resource                = &OrganizationResource{}
	_ resource.ResourceWithConfigure   = &OrganizationResource{}
	_ resource.ResourceWithImportState = &OrganizationResource{}
)

func NewResource() resource.Resource {
	return &OrganizationResource{}
}

type OrganizationResource struct {
	client *client.Client
}

type OrganizationResourceModel struct {
	ID   types.String `tfsdk:"id"`
	Name types.String `tfsdk:"name"`
	Logo types.String `tfsdk:"logo"`
}

func (r *OrganizationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_organization"
}

func (r *OrganizationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy organization.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the organization.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the organization.",
				Required:    true,
			},
			"logo": schema.StringAttribute{
				Description: "The logo URL or base64 encoded image for the organization.",
				Optional:    true,
			},
		},
	}
}

func (r *OrganizationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *OrganizationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan OrganizationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateOrganizationRequest{
		Name: plan.Name.ValueString(),
	}

	if !plan.Logo.IsNull() && !plan.Logo.IsUnknown() {
		logo := plan.Logo.ValueString()
		createReq.Logo = &logo
	}

	createResp, err := r.client.CreateOrganization(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Organization", "Could not create organization: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.OrganizationID)

	// Read back to get computed values
	org, err := r.client.GetOrganization(ctx, createResp.OrganizationID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Organization", "Could not read created organization: "+err.Error())
		return
	}

	plan.Name = types.StringValue(org.Name)
	if org.Logo != "" {
		plan.Logo = types.StringValue(org.Logo)
	} else {
		plan.Logo = types.StringNull()
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OrganizationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state OrganizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	org, err := r.client.GetOrganization(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Organization", "Could not read organization ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(org.OrganizationID)
	state.Name = types.StringValue(org.Name)
	if org.Logo != "" {
		state.Logo = types.StringValue(org.Logo)
	} else {
		state.Logo = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *OrganizationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan OrganizationResourceModel
	var state OrganizationResourceModel

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

	updateReq := client.UpdateOrganizationRequest{
		OrganizationID: state.ID.ValueString(),
		Name:           plan.Name.ValueString(),
	}

	if !plan.Logo.IsNull() && !plan.Logo.IsUnknown() {
		logo := plan.Logo.ValueString()
		updateReq.Logo = &logo
	}

	err := r.client.UpdateOrganization(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Organization", "Could not update organization: "+err.Error())
		return
	}

	plan.ID = state.ID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *OrganizationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state OrganizationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteOrganization(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Organization", "Could not delete organization: "+err.Error())
		return
	}
}

func (r *OrganizationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
