package redis

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
	_ resource.Resource                = &RedisResource{}
	_ resource.ResourceWithConfigure   = &RedisResource{}
	_ resource.ResourceWithImportState = &RedisResource{}
)

func NewResource() resource.Resource {
	return &RedisResource{}
}

type RedisResource struct {
	client *client.Client
}

type RedisResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	AppName          types.String `tfsdk:"app_name"`
	Description      types.String `tfsdk:"description"`
	EnvironmentID    types.String `tfsdk:"environment_id"`
	ServerID         types.String `tfsdk:"server_id"`
	DatabasePassword types.String `tfsdk:"database_password"`
	DockerImage      types.String `tfsdk:"docker_image"`
}

func (r *RedisResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_redis"
}

func (r *RedisResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy Redis database within an environment.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the redis database.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the redis database.",
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
				Description: "The description of the redis database.",
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
			"database_password": schema.StringAttribute{
				Description: "The database password.",
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
			},
			"docker_image": schema.StringAttribute{
				Description: "The Docker image to use (e.g., redis:7).",
				Optional:    true,
				Computed:    true,
			},
		},
	}
}

func (r *RedisResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RedisResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RedisResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	appName := strings.ToLower(strings.ReplaceAll(plan.Name.ValueString(), " ", "-"))

	databasePassword := generateRandomPassword(16)
	if !plan.DatabasePassword.IsNull() && !plan.DatabasePassword.IsUnknown() && plan.DatabasePassword.ValueString() != "" {
		databasePassword = plan.DatabasePassword.ValueString()
	}

	createReq := client.CreateRedisRequest{
		Name:             plan.Name.ValueString(),
		AppName:          appName,
		EnvironmentID:    plan.EnvironmentID.ValueString(),
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

	createResp, err := r.client.CreateRedis(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Redis", "Could not create redis: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.RedisID)

	redis, err := r.client.GetRedis(ctx, createResp.RedisID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Redis", "Could not read created redis: "+err.Error())
		return
	}

	plan.AppName = types.StringValue(redis.AppName)
	plan.DatabasePassword = types.StringValue(redis.DatabasePassword)
	plan.DockerImage = types.StringValue(redis.DockerImage)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedisResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RedisResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	redis, err := r.client.GetRedis(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Redis", "Could not read redis ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Name = types.StringValue(redis.Name)
	state.AppName = types.StringValue(redis.AppName)
	state.EnvironmentID = types.StringValue(redis.EnvironmentID)
	state.DatabasePassword = types.StringValue(redis.DatabasePassword)
	state.DockerImage = types.StringValue(redis.DockerImage)

	if redis.Description != "" {
		state.Description = types.StringValue(redis.Description)
	} else {
		state.Description = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RedisResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RedisResourceModel
	var state RedisResourceModel

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

	updateReq := client.UpdateRedisRequest{
		RedisID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}
	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}
	if !plan.DatabasePassword.Equal(state.DatabasePassword) {
		dbPass := plan.DatabasePassword.ValueString()
		updateReq.DatabasePassword = &dbPass
	}
	if !plan.DockerImage.Equal(state.DockerImage) {
		dockerImage := plan.DockerImage.ValueString()
		updateReq.DockerImage = &dockerImage
	}

	err := r.client.UpdateRedis(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Redis", "Could not update redis: "+err.Error())
		return
	}

	plan.ID = state.ID
	plan.AppName = state.AppName

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RedisResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RedisResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRedis(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Redis", "Could not delete redis: "+err.Error())
		return
	}
}

func (r *RedisResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func generateRandomPassword(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}
