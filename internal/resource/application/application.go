package application

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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ApplicationResource{}
	_ resource.ResourceWithConfigure   = &ApplicationResource{}
	_ resource.ResourceWithImportState = &ApplicationResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &ApplicationResource{}
}

// ApplicationResource is the resource implementation.
type ApplicationResource struct {
	client *client.Client
}

// ApplicationResourceModel describes the resource data model.
type ApplicationResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ServerID      types.String `tfsdk:"server_id"`
}

func (r *ApplicationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (r *ApplicationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy application within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the application.",
				Required:    true,
			},
			"app_name": schema.StringAttribute{
				Description: "The internal app name (used for container naming). Computed by Dokploy based on the name.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the application.",
				Optional:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment this application belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server to deploy to (optional, uses default if not specified).",
				Optional:    true,
			},
		},
	}
}

func (r *ApplicationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *ApplicationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ApplicationResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := client.CreateApplicationRequest{
		Name:          plan.Name.ValueString(),
		EnvironmentID: plan.EnvironmentID.ValueString(),
	}

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		createReq.Description = &desc
	}

	if !plan.ServerID.IsNull() {
		serverID := plan.ServerID.ValueString()
		createReq.ServerID = &serverID
	}

	// Create the application
	createResp, err := r.client.CreateApplication(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Dokploy Application",
			"Could not create application: "+err.Error(),
		)
		return
	}

	// Set state
	plan.ID = types.StringValue(createResp.ApplicationID)

	// Read back the application to get computed fields
	app, err := r.client.GetApplication(ctx, createResp.ApplicationID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy Application",
			"Could not read created application: "+err.Error(),
		)
		return
	}

	plan.AppName = types.StringValue(app.AppName)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApplicationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ApplicationResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get application from API
	app, err := r.client.GetApplication(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy Application",
			"Could not read application ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(app.Name)
	state.AppName = types.StringValue(app.AppName)
	// Preserve null if description is empty and was originally null
	if app.Description == "" && state.Description.IsNull() {
		// Keep it null
	} else {
		state.Description = types.StringValue(app.Description)
	}
	state.EnvironmentID = types.StringValue(app.EnvironmentID)
	// Preserve null if server_id is nil and was originally null
	if app.ServerID != nil {
		state.ServerID = types.StringValue(*app.ServerID)
	} else if !state.ServerID.IsNull() {
		state.ServerID = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ApplicationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ApplicationResourceModel
	var state ApplicationResourceModel

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

	// Build update request
	updateReq := client.UpdateApplicationRequest{
		ApplicationID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}

	// Update the application
	err := r.client.UpdateApplication(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Dokploy Application",
			"Could not update application: "+err.Error(),
		)
		return
	}

	// Update state with plan values
	plan.ID = state.ID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ApplicationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ApplicationResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteApplication(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dokploy Application",
			"Could not delete application: "+err.Error(),
		)
		return
	}
}

func (r *ApplicationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
