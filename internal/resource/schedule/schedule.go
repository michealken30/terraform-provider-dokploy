package schedule

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &ScheduleResource{}
	_ resource.ResourceWithConfigure   = &ScheduleResource{}
	_ resource.ResourceWithImportState = &ScheduleResource{}
)

func NewResource() resource.Resource {
	return &ScheduleResource{}
}

type ScheduleResource struct {
	client *client.Client
}

type ScheduleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	CronExpression types.String `tfsdk:"cron_expression"`
	Command        types.String `tfsdk:"command"`
	ShellType      types.String `tfsdk:"shell_type"`
	ScheduleType   types.String `tfsdk:"schedule_type"`
	AppName        types.String `tfsdk:"app_name"`
	ServiceName    types.String `tfsdk:"service_name"`
	Script         types.String `tfsdk:"script"`
	ApplicationID  types.String `tfsdk:"application_id"`
	ComposeID      types.String `tfsdk:"compose_id"`
	ServerID       types.String `tfsdk:"server_id"`
	Enabled        types.Bool   `tfsdk:"enabled"`
	Timezone       types.String `tfsdk:"timezone"`
}

func (r *ScheduleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_schedule"
}

func (r *ScheduleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy scheduled task (cron job) for an application, compose service, or server.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the schedule.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the scheduled task.",
				Required:    true,
			},
			"cron_expression": schema.StringAttribute{
				Description: "Cron expression for the schedule (e.g., '*/5 * * * *' for every 5 minutes).",
				Required:    true,
			},
			"command": schema.StringAttribute{
				Description: "The command to execute.",
				Required:    true,
			},
			"shell_type": schema.StringAttribute{
				Description: "The shell to use for executing the command. Valid values: 'bash', 'sh'. Defaults to 'bash'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("bash"),
				Validators: []validator.String{
					stringvalidator.OneOf("bash", "sh"),
				},
			},
			"schedule_type": schema.StringAttribute{
				Description: "The type of schedule. Valid values: 'application', 'compose', 'server', 'dokploy-server'. Defaults to 'application'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("application"),
				Validators: []validator.String{
					stringvalidator.OneOf("application", "compose", "server", "dokploy-server"),
				},
			},
			"app_name": schema.StringAttribute{
				Description: "The application name for the schedule. Used for container identification.",
				Optional:    true,
			},
			"service_name": schema.StringAttribute{
				Description: "The service name within a compose stack. Required when schedule_type is 'compose'.",
				Optional:    true,
			},
			"script": schema.StringAttribute{
				Description: "An optional script to execute instead of a command.",
				Optional:    true,
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this schedule belongs to. Required when schedule_type is 'application'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose_id": schema.StringAttribute{
				Description: "The ID of the compose service this schedule belongs to. Required when schedule_type is 'compose'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"server_id": schema.StringAttribute{
				Description: "The ID of the server this schedule belongs to. Required when schedule_type is 'server'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the schedule is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"timezone": schema.StringAttribute{
				Description: "The timezone for the schedule (e.g., 'America/New_York').",
				Optional:    true,
			},
		},
	}
}

func (r *ScheduleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ScheduleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan ScheduleResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateScheduleRequest{
		Name:           plan.Name.ValueString(),
		CronExpression: plan.CronExpression.ValueString(),
		Command:        plan.Command.ValueString(),
		ShellType:      plan.ShellType.ValueString(),
		ScheduleType:   plan.ScheduleType.ValueString(),
		Enabled:        plan.Enabled.ValueBool(),
	}

	if !plan.AppName.IsNull() && !plan.AppName.IsUnknown() {
		appName := plan.AppName.ValueString()
		createReq.AppName = &appName
	}

	if !plan.ServiceName.IsNull() && !plan.ServiceName.IsUnknown() {
		serviceName := plan.ServiceName.ValueString()
		createReq.ServiceName = &serviceName
	}

	if !plan.Script.IsNull() && !plan.Script.IsUnknown() {
		script := plan.Script.ValueString()
		createReq.Script = &script
	}

	if !plan.ApplicationID.IsNull() && !plan.ApplicationID.IsUnknown() {
		appID := plan.ApplicationID.ValueString()
		createReq.ApplicationID = &appID
	}

	if !plan.ComposeID.IsNull() && !plan.ComposeID.IsUnknown() {
		composeID := plan.ComposeID.ValueString()
		createReq.ComposeID = &composeID
	}

	if !plan.ServerID.IsNull() && !plan.ServerID.IsUnknown() {
		serverID := plan.ServerID.ValueString()
		createReq.ServerID = &serverID
	}

	if !plan.Timezone.IsNull() && !plan.Timezone.IsUnknown() {
		tz := plan.Timezone.ValueString()
		createReq.Timezone = &tz
	}

	createResp, err := r.client.CreateSchedule(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Schedule", "Could not create schedule: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.ScheduleID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ScheduleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state ScheduleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	schedule, err := r.client.GetSchedule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Schedule", "Could not read schedule ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(schedule.ScheduleID)
	state.Name = types.StringValue(schedule.Name)
	state.CronExpression = types.StringValue(schedule.CronExpression)
	state.Command = types.StringValue(schedule.Command)
	state.ShellType = types.StringValue(schedule.ShellType)
	state.ScheduleType = types.StringValue(schedule.ScheduleType)
	state.AppName = types.StringValue(schedule.AppName)
	state.Enabled = types.BoolValue(schedule.Enabled)

	if schedule.ServiceName != nil {
		state.ServiceName = types.StringValue(*schedule.ServiceName)
	} else {
		state.ServiceName = types.StringNull()
	}

	if schedule.Script != nil {
		state.Script = types.StringValue(*schedule.Script)
	} else {
		state.Script = types.StringNull()
	}

	if schedule.ApplicationID != nil {
		state.ApplicationID = types.StringValue(*schedule.ApplicationID)
	} else {
		state.ApplicationID = types.StringNull()
	}

	if schedule.ComposeID != nil {
		state.ComposeID = types.StringValue(*schedule.ComposeID)
	} else {
		state.ComposeID = types.StringNull()
	}

	if schedule.ServerID != nil {
		state.ServerID = types.StringValue(*schedule.ServerID)
	} else {
		state.ServerID = types.StringNull()
	}

	if schedule.Timezone != nil {
		state.Timezone = types.StringValue(*schedule.Timezone)
	} else {
		state.Timezone = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *ScheduleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan ScheduleResourceModel
	var state ScheduleResourceModel

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

	updateReq := client.UpdateScheduleRequest{
		ScheduleID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}

	if !plan.CronExpression.Equal(state.CronExpression) {
		cron := plan.CronExpression.ValueString()
		updateReq.CronExpression = &cron
	}

	if !plan.Command.Equal(state.Command) {
		cmd := plan.Command.ValueString()
		updateReq.Command = &cmd
	}

	if !plan.ShellType.Equal(state.ShellType) {
		shell := plan.ShellType.ValueString()
		updateReq.ShellType = &shell
	}

	if !plan.Script.Equal(state.Script) {
		if !plan.Script.IsNull() && !plan.Script.IsUnknown() {
			script := plan.Script.ValueString()
			updateReq.Script = &script
		}
	}

	if !plan.Enabled.Equal(state.Enabled) {
		enabled := plan.Enabled.ValueBool()
		updateReq.Enabled = &enabled
	}

	if !plan.Timezone.Equal(state.Timezone) {
		if !plan.Timezone.IsNull() && !plan.Timezone.IsUnknown() {
			tz := plan.Timezone.ValueString()
			updateReq.Timezone = &tz
		}
	}

	err := r.client.UpdateSchedule(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Schedule", "Could not update schedule: "+err.Error())
		return
	}

	// Preserve immutable fields from state
	plan.ID = state.ID
	plan.ApplicationID = state.ApplicationID
	plan.ComposeID = state.ComposeID
	plan.ServerID = state.ServerID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *ScheduleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state ScheduleResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSchedule(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Schedule", "Could not delete schedule: "+err.Error())
		return
	}
}

func (r *ScheduleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
