package compose

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

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ComposeResource{}
	_ resource.ResourceWithConfigure   = &ComposeResource{}
	_ resource.ResourceWithImportState = &ComposeResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &ComposeResource{}
}

// ComposeResource is the resource implementation.
type ComposeResource struct {
	client *client.Client
}

// ComposeResourceModel describes the resource data model.
type ComposeResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Name          types.String `tfsdk:"name"`
	AppName       types.String `tfsdk:"app_name"`
	Description   types.String `tfsdk:"description"`
	EnvironmentID types.String `tfsdk:"environment_id"`
	ServerID      types.String `tfsdk:"server_id"`
	ComposeFile   types.String `tfsdk:"compose_file"`
	ComposePath   types.String `tfsdk:"compose_path"`
	ComposeType   types.String `tfsdk:"compose_type"`
}

func (r *ComposeResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_compose"
}

func (r *ComposeResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy compose service within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the compose service.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the compose service.",
				Required:    true,
			},
			"app_name": schema.StringAttribute{
				Description: "The internal app name (used for container naming). Computed by Dokploy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Description: "The description of the compose service.",
				Optional:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment this compose service belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server to deploy to (optional, uses default if not specified).",
				Optional:    true,
			},
			"compose_file": schema.StringAttribute{
				Description: "The docker-compose.yml content.",
				Optional:    true,
			},
			"compose_path": schema.StringAttribute{
				Description: "The path to the compose file (default: ./docker-compose.yml).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"compose_type": schema.StringAttribute{
				Description: "The compose type (docker-compose or stack).",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *ComposeResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ComposeResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ComposeResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	createReq := client.CreateComposeRequest{
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

	// ComposeType is required, default to docker-compose
	composeType := "docker-compose"
	if !plan.ComposeType.IsNull() && !plan.ComposeType.IsUnknown() && plan.ComposeType.ValueString() != "" {
		composeType = plan.ComposeType.ValueString()
	}
	createReq.ComposeType = composeType

	// Create the compose service
	createResp, err := r.client.CreateCompose(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Dokploy Compose",
			"Could not create compose: "+err.Error(),
		)
		return
	}

	// Set state
	plan.ID = types.StringValue(createResp.ComposeID)

	// Read back the compose to get computed fields
	compose, err := r.client.GetCompose(ctx, createResp.ComposeID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy Compose",
			"Could not read created compose: "+err.Error(),
		)
		return
	}

	plan.AppName = types.StringValue(compose.AppName)
	plan.ComposePath = types.StringValue(compose.ComposePath)
	plan.ComposeType = types.StringValue(compose.ComposeType)

	// Update compose file if provided
	if !plan.ComposeFile.IsNull() && plan.ComposeFile.ValueString() != "" {
		updateReq := client.UpdateComposeRequest{
			ComposeID:   createResp.ComposeID,
			ComposeFile: stringPtr(plan.ComposeFile.ValueString()),
		}
		err := r.client.UpdateCompose(ctx, updateReq)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Updating Dokploy Compose File",
				"Could not update compose file: "+err.Error(),
			)
			return
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ComposeResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ComposeResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get compose from API
	compose, err := r.client.GetCompose(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy Compose",
			"Could not read compose ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(compose.Name)
	state.AppName = types.StringValue(compose.AppName)
	state.EnvironmentID = types.StringValue(compose.EnvironmentID)
	state.ComposePath = types.StringValue(compose.ComposePath)
	state.ComposeType = types.StringValue(compose.ComposeType)

	// Handle optional fields - keep null if empty to avoid drift
	if compose.Description != "" {
		state.Description = types.StringValue(compose.Description)
	} else {
		state.Description = types.StringNull()
	}
	if compose.ComposeFile != "" {
		state.ComposeFile = types.StringValue(compose.ComposeFile)
	} else {
		state.ComposeFile = types.StringNull()
	}
	if compose.ServerID != nil && *compose.ServerID != "" {
		state.ServerID = types.StringValue(*compose.ServerID)
	} else {
		state.ServerID = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ComposeResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ComposeResourceModel
	var state ComposeResourceModel

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
	updateReq := client.UpdateComposeRequest{
		ComposeID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}

	if !plan.ComposeFile.Equal(state.ComposeFile) {
		composeFile := plan.ComposeFile.ValueString()
		updateReq.ComposeFile = &composeFile
	}

	// Update the compose
	err := r.client.UpdateCompose(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Dokploy Compose",
			"Could not update compose: "+err.Error(),
		)
		return
	}

	// Update state with plan values, preserve computed fields from state
	plan.ID = state.ID
	plan.AppName = state.AppName
	plan.ComposePath = state.ComposePath
	plan.ComposeType = state.ComposeType

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ComposeResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ComposeResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCompose(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dokploy Compose",
			"Could not delete compose: "+err.Error(),
		)
		return
	}
}

func (r *ComposeResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func stringPtr(s string) *string {
	return &s
}
