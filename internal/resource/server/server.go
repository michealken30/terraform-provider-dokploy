package server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &ServerResource{}
	_ resource.ResourceWithConfigure   = &ServerResource{}
	_ resource.ResourceWithImportState = &ServerResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &ServerResource{}
}

// ServerResource is the resource implementation.
type ServerResource struct {
	client *client.Client
}

// ServerResourceModel describes the resource data model.
type ServerResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	Description types.String `tfsdk:"description"`
	IPAddress   types.String `tfsdk:"ip_address"`
	Port        types.Int64  `tfsdk:"port"`
	Username    types.String `tfsdk:"username"`
	SSHKeyID    types.String `tfsdk:"ssh_key_id"`
	ServerType  types.String `tfsdk:"server_type"`
}

func (r *ServerResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (r *ServerResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the server.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the server.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the server.",
				Optional:    true,
			},
			"ip_address": schema.StringAttribute{
				Description: "The IP address of the server.",
				Required:    true,
			},
			"port": schema.Int64Attribute{
				Description: "The SSH port of the server.",
				Required:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Description: "The SSH username for the server.",
				Required:    true,
			},
			"ssh_key_id": schema.StringAttribute{
				Description: "The ID of the SSH key to use for the server.",
				Required:    true,
			},
			"server_type": schema.StringAttribute{
				Description: "The type of server: 'deploy' or 'build'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("deploy"),
			},
		},
	}
}

func (r *ServerResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ServerResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ServerResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build create request
	sshKeyID := plan.SSHKeyID.ValueString()
	createReq := client.CreateServerRequest{
		Name:       plan.Name.ValueString(),
		IPAddress:  plan.IPAddress.ValueString(),
		Port:       int(plan.Port.ValueInt64()),
		Username:   plan.Username.ValueString(),
		SSHKeyID:   &sshKeyID,
		ServerType: plan.ServerType.ValueString(),
	}

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		createReq.Description = &desc
	}

	// Create the server
	createResp, err := r.client.CreateServer(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Dokploy Server",
			"Could not create server: "+err.Error(),
		)
		return
	}

	// Set state
	plan.ID = types.StringValue(createResp.ServerID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ServerResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get server from API
	server, err := r.client.GetServer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy Server",
			"Could not read server ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(server.Name)
	state.Description = types.StringValue(server.Description)
	state.IPAddress = types.StringValue(server.IPAddress)
	state.Port = types.Int64Value(int64(server.Port))
	state.Username = types.StringValue(server.Username)
	state.SSHKeyID = types.StringValue(server.SSHKeyID)
	state.ServerType = types.StringValue(server.ServerType)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ServerResourceModel
	var state ServerResourceModel

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
	updateReq := client.UpdateServerRequest{
		ServerID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}

	if !plan.IPAddress.Equal(state.IPAddress) {
		ip := plan.IPAddress.ValueString()
		updateReq.IPAddress = &ip
	}

	if !plan.Port.Equal(state.Port) {
		port := int(plan.Port.ValueInt64())
		updateReq.Port = &port
	}

	if !plan.Username.Equal(state.Username) {
		username := plan.Username.ValueString()
		updateReq.Username = &username
	}

	if !plan.SSHKeyID.Equal(state.SSHKeyID) {
		sshKeyID := plan.SSHKeyID.ValueString()
		updateReq.SSHKeyID = &sshKeyID
	}

	// Update the server
	err := r.client.UpdateServer(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Dokploy Server",
			"Could not update server: "+err.Error(),
		)
		return
	}

	// Update state with plan values
	plan.ID = state.ID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ServerResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ServerResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteServer(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dokploy Server",
			"Could not delete server: "+err.Error(),
		)
		return
	}
}

func (r *ServerResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
