package mariadb

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

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &MariadbResource{}
	_ resource.ResourceWithConfigure   = &MariadbResource{}
	_ resource.ResourceWithImportState = &MariadbResource{}
)

func NewResource() resource.Resource {
	return &MariadbResource{}
}

type MariadbResource struct {
	client *client.Client
}

type MariadbResourceModel struct {
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

func (r *MariadbResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mariadb"
}

func (r *MariadbResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy MariaDB database within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the mariadb database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the mariadb database.",
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
				Description: "The description of the mariadb database.",
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
				Description: "The Docker image to use (e.g., mariadb:11).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *MariadbResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Resource Configure Type", fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	r.client = client
}

func (r *MariadbResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MariadbResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := strings.ToLower(strings.ReplaceAll(plan.Name.ValueString(), " ", "-"))

	databaseName := appName
	if !plan.DatabaseName.IsNull() && !plan.DatabaseName.IsUnknown() && plan.DatabaseName.ValueString() != "" {
		databaseName = plan.DatabaseName.ValueString()
	}

	databaseUser := appName
	if !plan.DatabaseUser.IsNull() && !plan.DatabaseUser.IsUnknown() && plan.DatabaseUser.ValueString() != "" {
		databaseUser = plan.DatabaseUser.ValueString()
	}

	databasePassword := generateRandomPassword(16)
	if !plan.DatabasePassword.IsNull() && !plan.DatabasePassword.IsUnknown() && plan.DatabasePassword.ValueString() != "" {
		databasePassword = plan.DatabasePassword.ValueString()
	}

	databaseRootPassword := generateRandomPassword(16)
	if !plan.DatabaseRootPassword.IsNull() && !plan.DatabaseRootPassword.IsUnknown() && plan.DatabaseRootPassword.ValueString() != "" {
		databaseRootPassword = plan.DatabaseRootPassword.ValueString()
	}

	createReq := client.CreateMariadbRequest{
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

	createResp, err := r.client.CreateMariadb(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy MariaDB", "Could not create mariadb: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.MariadbID)

	mariadb, err := r.client.GetMariadb(ctx, createResp.MariadbID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy MariaDB", "Could not read created mariadb: "+err.Error())
		return
	}

	plan.AppName = types.StringValue(mariadb.AppName)
	plan.DatabaseName = types.StringValue(mariadb.DatabaseName)
	plan.DatabaseUser = types.StringValue(mariadb.DatabaseUser)
	plan.DatabasePassword = types.StringValue(mariadb.DatabasePassword)
	plan.DatabaseRootPassword = types.StringValue(mariadb.DatabaseRootPassword)
	plan.DockerImage = types.StringValue(mariadb.DockerImage)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MariadbResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MariadbResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mariadb, err := r.client.GetMariadb(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy MariaDB", "Could not read mariadb ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Name = types.StringValue(mariadb.Name)
	state.AppName = types.StringValue(mariadb.AppName)
	state.EnvironmentID = types.StringValue(mariadb.EnvironmentID)
	state.DatabaseName = types.StringValue(mariadb.DatabaseName)
	state.DatabaseUser = types.StringValue(mariadb.DatabaseUser)
	state.DatabasePassword = types.StringValue(mariadb.DatabasePassword)
	state.DatabaseRootPassword = types.StringValue(mariadb.DatabaseRootPassword)
	state.DockerImage = types.StringValue(mariadb.DockerImage)

	if mariadb.Description != "" {
		state.Description = types.StringValue(mariadb.Description)
	} else {
		state.Description = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *MariadbResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MariadbResourceModel
	var state MariadbResourceModel

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

	updateReq := client.UpdateMariadbRequest{
		MariadbID: state.ID.ValueString(),
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

	err := r.client.UpdateMariadb(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy MariaDB", "Could not update mariadb: "+err.Error())
		return
	}

	plan.ID = state.ID
	plan.AppName = state.AppName

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MariadbResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MariadbResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMariadb(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy MariaDB", "Could not delete mariadb: "+err.Error())
		return
	}
}

func (r *MariadbResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func generateRandomPassword(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random password: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}
