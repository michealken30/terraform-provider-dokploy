package redirect

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &RedirectResource{}
	_ resource.ResourceWithConfigure   = &RedirectResource{}
	_ resource.ResourceWithImportState = &RedirectResource{}
)

func NewResource() resource.Resource {
	return &RedirectResource{}
}

type RedirectResource struct {
	client *client.Client
}

type RedirectResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Regex         types.String `tfsdk:"regex"`
	Replacement   types.String `tfsdk:"replacement"`
	Permanent     types.Bool   `tfsdk:"permanent"`
	ApplicationID types.String `tfsdk:"application_id"`
}

func (r *RedirectResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redirect"
}

func (r *RedirectResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy redirect rule for an application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the redirect.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"regex": schema.StringAttribute{
				Description: "The regex pattern to match URLs against.",
				Required:    true,
			},
			"replacement": schema.StringAttribute{
				Description: "The replacement URL pattern. Can include capture groups from the regex.",
				Required:    true,
			},
			"permanent": schema.BoolAttribute{
				Description: "Whether this is a permanent (301) or temporary (302) redirect. Defaults to false.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this redirect belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RedirectResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RedirectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedirectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateRedirectRequest{
		Regex:         plan.Regex.ValueString(),
		Replacement:   plan.Replacement.ValueString(),
		Permanent:     plan.Permanent.ValueBool(),
		ApplicationID: plan.ApplicationID.ValueString(),
	}

	createResp, err := r.client.CreateRedirect(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Redirect", "Could not create redirect: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.RedirectID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedirectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redirect, err := r.client.GetRedirect(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Redirect", "Could not read redirect ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(redirect.RedirectID)
	state.Regex = types.StringValue(redirect.Regex)
	state.Replacement = types.StringValue(redirect.Replacement)
	state.Permanent = types.BoolValue(redirect.Permanent)
	state.ApplicationID = types.StringValue(redirect.ApplicationID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RedirectResourceModel
	var state RedirectResourceModel

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

	updateReq := client.UpdateRedirectRequest{
		RedirectID:  state.ID.ValueString(),
		Regex:       plan.Regex.ValueString(),
		Replacement: plan.Replacement.ValueString(),
		Permanent:   plan.Permanent.ValueBool(),
	}

	err := r.client.UpdateRedirect(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Redirect", "Could not update redirect: "+err.Error())
		return
	}

	// Preserve immutable fields from state
	plan.ID = state.ID
	plan.ApplicationID = state.ApplicationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedirectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedirectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRedirect(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Redirect", "Could not delete redirect: "+err.Error())
		return
	}
}

func (r *RedirectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
