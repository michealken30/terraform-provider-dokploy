package postgres

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
	_ resource.Resource                = &PostgresResource{}
	_ resource.ResourceWithConfigure   = &PostgresResource{}
	_ resource.ResourceWithImportState = &PostgresResource{}
)

// NewResource is a helper function to simplify the provider implementation.
func NewResource() resource.Resource {
	return &PostgresResource{}
}

// PostgresResource is the resource implementation.
type PostgresResource struct {
	client *client.Client
}

// PostgresResourceModel describes the resource data model.
type PostgresResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	AppName          types.String `tfsdk:"app_name"`
	Description      types.String `tfsdk:"description"`
	EnvironmentID    types.String `tfsdk:"environment_id"`
	ServerID         types.String `tfsdk:"server_id"`
	DatabaseName     types.String `tfsdk:"database_name"`
	DatabaseUser     types.String `tfsdk:"database_user"`
	DatabasePassword types.String `tfsdk:"database_password"`
	DockerImage      types.String `tfsdk:"docker_image"`
}

func (r *PostgresResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_postgres"
}

func (r *PostgresResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy PostgreSQL database within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the postgres database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the postgres database.",
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
				Description: "The description of the postgres database.",
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
			"docker_image": schema.StringAttribute{
				Description: "The Docker image to use (e.g., postgres:16).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *PostgresResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PostgresResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PostgresResourceModel

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

	// Generate a default password if not provided
	databasePassword := generateRandomPassword(16)
	if !plan.DatabasePassword.IsNull() && !plan.DatabasePassword.IsUnknown() && plan.DatabasePassword.ValueString() != "" {
		databasePassword = plan.DatabasePassword.ValueString()
	}

	// Build create request
	createReq := client.CreatePostgresRequest{
		Name:             plan.Name.ValueString(),
		AppName:          appName,
		EnvironmentID:    plan.EnvironmentID.ValueString(),
		DatabaseName:     databaseName,
		DatabaseUser:     databaseUser,
		DatabasePassword: databasePassword,
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

	// Create the postgres database
	createResp, err := r.client.CreatePostgres(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Creating Dokploy Postgres",
			"Could not create postgres: "+err.Error(),
		)
		return
	}

	// Set state
	plan.ID = types.StringValue(createResp.PostgresID)

	// Read back the postgres to get computed fields
	pg, err := r.client.GetPostgres(ctx, createResp.PostgresID)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy Postgres",
			"Could not read created postgres: "+err.Error(),
		)
		return
	}

	plan.AppName = types.StringValue(pg.AppName)
	plan.DatabaseName = types.StringValue(pg.DatabaseName)
	plan.DatabaseUser = types.StringValue(pg.DatabaseUser)
	plan.DatabasePassword = types.StringValue(pg.DatabasePassword)
	plan.DockerImage = types.StringValue(pg.DockerImage)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PostgresResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get postgres from API
	pg, err := r.client.GetPostgres(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Dokploy Postgres",
			"Could not read postgres ID "+state.ID.ValueString()+": "+err.Error(),
		)
		return
	}

	// Update state
	state.Name = types.StringValue(pg.Name)
	state.AppName = types.StringValue(pg.AppName)
	state.EnvironmentID = types.StringValue(pg.EnvironmentID)
	state.DatabaseName = types.StringValue(pg.DatabaseName)
	state.DatabaseUser = types.StringValue(pg.DatabaseUser)
	state.DatabasePassword = types.StringValue(pg.DatabasePassword)
	state.DockerImage = types.StringValue(pg.DockerImage)

	// Handle optional fields - keep null if empty to avoid drift
	if pg.Description != "" {
		state.Description = types.StringValue(pg.Description)
	} else {
		state.Description = types.StringNull()
	}
	if pg.ServerID != nil && *pg.ServerID != "" {
		state.ServerID = types.StringValue(*pg.ServerID)
	} else {
		state.ServerID = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PostgresResourceModel
	var state PostgresResourceModel

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
	updateReq := client.UpdatePostgresRequest{
		PostgresID: state.ID.ValueString(),
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

	if !plan.DockerImage.Equal(state.DockerImage) {
		dockerImage := plan.DockerImage.ValueString()
		updateReq.DockerImage = &dockerImage
	}

	// Update the postgres
	err := r.client.UpdatePostgres(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Updating Dokploy Postgres",
			"Could not update postgres: "+err.Error(),
		)
		return
	}

	// Update state with plan values
	plan.ID = state.ID
	plan.AppName = state.AppName

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PostgresResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PostgresResourceModel

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePostgres(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting Dokploy Postgres",
			"Could not delete postgres: "+err.Error(),
		)
		return
	}
}

func (r *PostgresResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

// generateRandomPassword generates a random password of the specified length
func generateRandomPassword(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random password: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}
