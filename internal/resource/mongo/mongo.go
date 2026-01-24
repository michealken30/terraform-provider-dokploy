package mongo

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

var (
	_ resource.Resource                = &MongoResource{}
	_ resource.ResourceWithConfigure   = &MongoResource{}
	_ resource.ResourceWithImportState = &MongoResource{}
)

func NewResource() resource.Resource {
	return &MongoResource{}
}

type MongoResource struct {
	client *client.Client
}

type MongoResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	AppName          types.String `tfsdk:"app_name"`
	Description      types.String `tfsdk:"description"`
	EnvironmentID    types.String `tfsdk:"environment_id"`
	ServerID         types.String `tfsdk:"server_id"`
	DatabaseUser     types.String `tfsdk:"database_user"`
	DatabasePassword types.String `tfsdk:"database_password"`
	DockerImage      types.String `tfsdk:"docker_image"`
}

func (r *MongoResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mongo"
}

func (r *MongoResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy MongoDB database within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the mongo database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the mongo database.",
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
				Description: "The description of the mongo database.",
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
				Description: "The Docker image to use (e.g., mongo:7).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *MongoResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MongoResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MongoResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := strings.ToLower(strings.ReplaceAll(plan.Name.ValueString(), " ", "-"))

	databaseUser := appName
	if !plan.DatabaseUser.IsNull() && !plan.DatabaseUser.IsUnknown() && plan.DatabaseUser.ValueString() != "" {
		databaseUser = plan.DatabaseUser.ValueString()
	}

	databasePassword := generateRandomPassword(16)
	if !plan.DatabasePassword.IsNull() && !plan.DatabasePassword.IsUnknown() && plan.DatabasePassword.ValueString() != "" {
		databasePassword = plan.DatabasePassword.ValueString()
	}

	createReq := client.CreateMongoRequest{
		Name:             plan.Name.ValueString(),
		AppName:          appName,
		EnvironmentID:    plan.EnvironmentID.ValueString(),
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

	createResp, err := r.client.CreateMongo(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy MongoDB", "Could not create mongo: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.MongoID)

	mongo, err := r.client.GetMongo(ctx, createResp.MongoID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy MongoDB", "Could not read created mongo: "+err.Error())
		return
	}

	plan.AppName = types.StringValue(mongo.AppName)
	plan.DatabaseUser = types.StringValue(mongo.DatabaseUser)
	plan.DatabasePassword = types.StringValue(mongo.DatabasePassword)
	plan.DockerImage = types.StringValue(mongo.DockerImage)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MongoResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MongoResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mongo, err := r.client.GetMongo(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy MongoDB", "Could not read mongo ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Name = types.StringValue(mongo.Name)
	state.AppName = types.StringValue(mongo.AppName)
	state.EnvironmentID = types.StringValue(mongo.EnvironmentID)
	state.DatabaseUser = types.StringValue(mongo.DatabaseUser)
	state.DatabasePassword = types.StringValue(mongo.DatabasePassword)
	state.DockerImage = types.StringValue(mongo.DockerImage)

	if mongo.Description != "" {
		state.Description = types.StringValue(mongo.Description)
	} else {
		state.Description = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *MongoResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MongoResourceModel
	var state MongoResourceModel

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

	updateReq := client.UpdateMongoRequest{
		MongoID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}
	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
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

	err := r.client.UpdateMongo(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy MongoDB", "Could not update mongo: "+err.Error())
		return
	}

	plan.ID = state.ID
	plan.AppName = state.AppName

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MongoResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MongoResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMongo(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy MongoDB", "Could not delete mongo: "+err.Error())
		return
	}
}

func (r *MongoResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func generateRandomPassword(length int) string {
	b := make([]byte, length)
	if _, err := rand.Read(b); err != nil {
		panic("failed to generate random password: " + err.Error())
	}
	return base64.URLEncoding.EncodeToString(b)[:length]
}
