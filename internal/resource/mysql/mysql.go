package mysql

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"

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
	_ resource.Resource                = &MysqlResource{}
	_ resource.ResourceWithConfigure   = &MysqlResource{}
	_ resource.ResourceWithImportState = &MysqlResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &MysqlResource{}
}

// MysqlResource is the resource implementation.
type MysqlResource struct {
	client *client.Client
}

// MysqlResourceModel describes the resource data model.
type MysqlResourceModel struct {
	ID                   types.String `tfsdk:"id"`
	Name                 types.String `tfsdk:"name"`
	AppName              types.String `tfsdk:"app_name"`
	Description          types.String `tfsdk:"description"`
	EnvironmentID        types.String `tfsdk:"environment_id"`
	ServerID             types.String `tfsdk:"server_id"`
	DatabaseName         types.String `tfsdk:"database_name"`
	DatabaseUser         types.String `tfsdk:"database_user"`
	DatabasePassword     types.String `tfsdk:"database_password"`
	DatabaseRootPassword types.String `tfsdk:"database_root_password"`
	DockerImage          types.String `tfsdk:"docker_image"`
}

func (r *MysqlResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mysql"
}

func (r *MysqlResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy MySQL database within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the mysql database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the mysql database.",
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
				Description: "The description of the mysql database.",
				Optional:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The ID of the environment this database belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server to deploy to (optional, uses default if not specified).",
				Optional:    true,
			},
			"database_name": schema.StringAttribute{
				Description: "The name of the database to create.",
				Optional:    true,
				Computed:    true,
			},
			"database_user": schema.StringAttribute{
				Description: "The database user.",
				Optional:    true,
				Computed:    true,
			},
			"database_password": schema.StringAttribute{
				Description: "The database password.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"database_root_password": schema.StringAttribute{
				Description: "The database root password.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"docker_image": schema.StringAttribute{
				Description: "The Docker image to use (e.g., mysql:8).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *MysqlResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MysqlResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MysqlResourceModel

	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate appName from name (lowercase, replace spaces with dashes)
	appName := strings.ToLower(strings.ReplaceAll(plan.Name.ValueString(), " ", "-"))

	// Default database name and user to the app name if not specified
	databaseName := appName
	if !plan.DatabaseName.IsNull() && !plan.DatabaseName.IsUnknown() && plan.DatabaseName.ValueString() != "" {
		databaseName = plan.DatabaseName.ValueString()
	}

	databaseUser := appName
	if !plan.DatabaseUser.IsNull() && !plan.DatabaseUser.IsUnknown() && plan.DatabaseUser.ValueString() != "" {
		databaseUser = plan.DatabaseUser.ValueString()
	}

	// Generate default passwords if not provided
	databasePassword := generateRandomPassword(16)
	if !plan.DatabasePassword.IsNull() && !plan.DatabasePassword.IsUnknown() && plan.DatabasePassword.ValueString() != "" {
		databasePassword = plan.DatabasePassword.ValueString()
	}

	databaseRootPassword := generateRandomPassword(16)
	if !plan.DatabaseRootPassword.IsNull() && !plan.DatabaseRootPassword.IsUnknown() && plan.DatabaseRootPassword.ValueString() != "" {
		databaseRootPassword = plan.DatabaseRootPassword.ValueString()
	}

	// Build create request
	createReq := client.CreateMysqlRequest{
		Name:                 plan.Name.ValueString(),
		AppName:              appName,
		EnvironmentID:        plan.EnvironmentID.ValueString(),
		DatabaseName:         databaseName,
		DatabaseUser:         databaseUser,
		DatabasePassword:     databasePassword,
		DatabaseRootPassword: databaseRootPassword,
	}

	if !plan.Description.IsNull() {
		desc := plan.Description.ValueString()
		createReq.Description = &desc
	}

	if !plan.ServerID.IsNull() {
		serverID := plan.ServerID.ValueString()
		createReq.ServerID = &serverID
	}

	if !plan.DockerImage.IsNull() {
		dockerImage := plan.DockerImage.ValueString()
		createReq.DockerImage = &dockerImage
	}

	// Create the mysql database
	createResp, err := r.client.CreateMysql(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Dokploy MySQL",
			"Could not create mysql: "+err.Error(),
		)
		return
	}

	// Set state
	plan.ID = types.StringValue(createResp.MysqlID)

	// Read back the mysql to get computed fields
	mysql, err := r.client.GetMysql(ctx, createResp.MysqlID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy MySQL",
			"Could not read created mysql: "+err.Error(),
		)
		return
	}

	plan.AppName = types.StringValue(mysql.AppName)
	plan.DatabaseName = types.StringValue(mysql.DatabaseName)
	plan.DatabaseUser = types.StringValue(mysql.DatabaseUser)
	plan.DatabasePassword = types.StringValue(mysql.DatabasePassword)
	plan.DatabaseRootPassword = types.StringValue(mysql.DatabaseRootPassword)
	plan.DockerImage = types.StringValue(mysql.DockerImage)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MysqlResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MysqlResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get mysql from API
	mysql, err := r.client.GetMysql(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy MySQL",
			"Could not read mysql ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(mysql.Name)
	state.AppName = types.StringValue(mysql.AppName)
	state.EnvironmentID = types.StringValue(mysql.EnvironmentID)
	state.DatabaseName = types.StringValue(mysql.DatabaseName)
	state.DatabaseUser = types.StringValue(mysql.DatabaseUser)
	state.DatabasePassword = types.StringValue(mysql.DatabasePassword)
	state.DatabaseRootPassword = types.StringValue(mysql.DatabaseRootPassword)
	state.DockerImage = types.StringValue(mysql.DockerImage)

	// Handle optional fields - keep null if empty to avoid drift
	if mysql.Description != "" {
		state.Description = types.StringValue(mysql.Description)
	} else {
		state.Description = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *MysqlResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MysqlResourceModel
	var state MysqlResourceModel

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
	updateReq := client.UpdateMysqlRequest{
		MysqlID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}

	if !plan.DatabaseName.Equal(state.DatabaseName) {
		dbName := plan.DatabaseName.ValueString()
		updateReq.DatabaseName = &dbName
	}

	if !plan.DatabaseUser.Equal(state.DatabaseUser) {
		dbUser := plan.DatabaseUser.ValueString()
		updateReq.DatabaseUser = &dbUser
	}

	if !plan.DatabasePassword.Equal(state.DatabasePassword) {
		dbPass := plan.DatabasePassword.ValueString()
		updateReq.DatabasePassword = &dbPass
	}

	if !plan.DatabaseRootPassword.Equal(state.DatabaseRootPassword) {
		dbRootPass := plan.DatabaseRootPassword.ValueString()
		updateReq.DatabaseRootPassword = &dbRootPass
	}

	if !plan.DockerImage.Equal(state.DockerImage) {
		dockerImage := plan.DockerImage.ValueString()
		updateReq.DockerImage = &dockerImage
	}

	// Update the mysql
	err := r.client.UpdateMysql(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Dokploy MySQL",
			"Could not update mysql: "+err.Error(),
		)
		return
	}

	// Update state with plan values
	plan.ID = state.ID
	plan.AppName = state.AppName

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MysqlResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MysqlResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMysql(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dokploy MySQL",
			"Could not delete mysql: "+err.Error(),
		)
		return
	}
}

func (r *MysqlResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// generateRandomPassword generates a random password of the specified length
func generateRandomPassword(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
