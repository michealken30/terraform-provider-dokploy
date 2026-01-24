package volumebackup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &VolumeBackupResource{}
	_ resource.ResourceWithConfigure   = &VolumeBackupResource{}
	_ resource.ResourceWithImportState = &VolumeBackupResource{}
)

func NewResource() resource.Resource {
	return &VolumeBackupResource{}
}

type VolumeBackupResource struct {
	client *client.Client
}

type VolumeBackupResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	VolumeName      types.String `tfsdk:"volume_name"`
	Prefix          types.String `tfsdk:"prefix"`
	CronExpression  types.String `tfsdk:"cron_expression"`
	DestinationID   types.String `tfsdk:"destination_id"`
	ServiceType     types.String `tfsdk:"service_type"`
	AppName         types.String `tfsdk:"app_name"`
	ServiceName     types.String `tfsdk:"service_name"`
	TurnOff         types.Bool   `tfsdk:"turn_off"`
	KeepLatestCount types.Int64  `tfsdk:"keep_latest_count"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	ApplicationID   types.String `tfsdk:"application_id"`
	PostgresID      types.String `tfsdk:"postgres_id"`
	MysqlID         types.String `tfsdk:"mysql_id"`
	MariadbID       types.String `tfsdk:"mariadb_id"`
	MongoID         types.String `tfsdk:"mongo_id"`
	RedisID         types.String `tfsdk:"redis_id"`
	ComposeID       types.String `tfsdk:"compose_id"`
}

func (r *VolumeBackupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_volume_backup"
}

func (r *VolumeBackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy volume backup configuration. Backs up Docker volumes to a storage destination on a schedule.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the volume backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the volume backup configuration.",
				Required:    true,
			},
			"volume_name": schema.StringAttribute{
				Description: "The name of the Docker volume to back up.",
				Required:    true,
			},
			"prefix": schema.StringAttribute{
				Description: "The prefix for backup files in the destination.",
				Required:    true,
			},
			"cron_expression": schema.StringAttribute{
				Description: "Cron expression for the backup schedule (e.g., '0 2 * * *' for daily at 2 AM).",
				Required:    true,
			},
			"destination_id": schema.StringAttribute{
				Description: "The ID of the storage destination for backups.",
				Required:    true,
			},
			"service_type": schema.StringAttribute{
				Description: "The type of service owning the volume. Valid values: 'application', 'postgres', 'mysql', 'mariadb', 'mongo', 'redis', 'compose'.",
				Optional:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("application", "postgres", "mysql", "mariadb", "mongo", "redis", "compose"),
				},
			},
			"app_name": schema.StringAttribute{
				Description: "The application name (used for identifying the volume).",
				Optional:    true,
			},
			"service_name": schema.StringAttribute{
				Description: "The service name within a compose stack.",
				Optional:    true,
			},
			"turn_off": schema.BoolAttribute{
				Description: "Whether to stop the service during backup for consistency.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"keep_latest_count": schema.Int64Attribute{
				Description: "Number of recent backups to keep. Older backups are automatically deleted.",
				Optional:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the backup schedule is enabled.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"application_id": schema.StringAttribute{
				Description: "The application ID if backing up an application volume.",
				Optional:    true,
			},
			"postgres_id": schema.StringAttribute{
				Description: "The PostgreSQL service ID if backing up a PostgreSQL volume.",
				Optional:    true,
			},
			"mysql_id": schema.StringAttribute{
				Description: "The MySQL service ID if backing up a MySQL volume.",
				Optional:    true,
			},
			"mariadb_id": schema.StringAttribute{
				Description: "The MariaDB service ID if backing up a MariaDB volume.",
				Optional:    true,
			},
			"mongo_id": schema.StringAttribute{
				Description: "The MongoDB service ID if backing up a MongoDB volume.",
				Optional:    true,
			},
			"redis_id": schema.StringAttribute{
				Description: "The Redis service ID if backing up a Redis volume.",
				Optional:    true,
			},
			"compose_id": schema.StringAttribute{
				Description: "The compose service ID if backing up a compose volume.",
				Optional:    true,
			},
		},
	}
}

func (r *VolumeBackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *VolumeBackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan VolumeBackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateVolumeBackupRequest{
		Name:           plan.Name.ValueString(),
		VolumeName:     plan.VolumeName.ValueString(),
		Prefix:         plan.Prefix.ValueString(),
		CronExpression: plan.CronExpression.ValueString(),
		DestinationID:  plan.DestinationID.ValueString(),
	}

	if !plan.ServiceType.IsNull() && !plan.ServiceType.IsUnknown() {
		v := plan.ServiceType.ValueString()
		createReq.ServiceType = &v
	}
	if !plan.AppName.IsNull() && !plan.AppName.IsUnknown() {
		v := plan.AppName.ValueString()
		createReq.AppName = &v
	}
	if !plan.ServiceName.IsNull() && !plan.ServiceName.IsUnknown() {
		v := plan.ServiceName.ValueString()
		createReq.ServiceName = &v
	}
	if !plan.TurnOff.IsNull() && !plan.TurnOff.IsUnknown() {
		v := plan.TurnOff.ValueBool()
		createReq.TurnOff = &v
	}
	if !plan.KeepLatestCount.IsNull() && !plan.KeepLatestCount.IsUnknown() {
		v := int(plan.KeepLatestCount.ValueInt64())
		createReq.KeepLatestCount = &v
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		createReq.Enabled = &v
	}
	if !plan.ApplicationID.IsNull() && !plan.ApplicationID.IsUnknown() {
		v := plan.ApplicationID.ValueString()
		createReq.ApplicationID = &v
	}
	if !plan.PostgresID.IsNull() && !plan.PostgresID.IsUnknown() {
		v := plan.PostgresID.ValueString()
		createReq.PostgresID = &v
	}
	if !plan.MysqlID.IsNull() && !plan.MysqlID.IsUnknown() {
		v := plan.MysqlID.ValueString()
		createReq.MysqlID = &v
	}
	if !plan.MariadbID.IsNull() && !plan.MariadbID.IsUnknown() {
		v := plan.MariadbID.ValueString()
		createReq.MariadbID = &v
	}
	if !plan.MongoID.IsNull() && !plan.MongoID.IsUnknown() {
		v := plan.MongoID.ValueString()
		createReq.MongoID = &v
	}
	if !plan.RedisID.IsNull() && !plan.RedisID.IsUnknown() {
		v := plan.RedisID.ValueString()
		createReq.RedisID = &v
	}
	if !plan.ComposeID.IsNull() && !plan.ComposeID.IsUnknown() {
		v := plan.ComposeID.ValueString()
		createReq.ComposeID = &v
	}

	backup, err := r.client.CreateVolumeBackup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Volume Backup", err.Error())
		return
	}

	plan.ID = types.StringValue(backup.VolumeBackupID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeBackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state VolumeBackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.GetVolumeBackup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Volume Backup", err.Error())
		return
	}

	state.ID = types.StringValue(backup.VolumeBackupID)
	state.Name = types.StringValue(backup.Name)
	state.VolumeName = types.StringValue(backup.VolumeName)
	state.Prefix = types.StringValue(backup.Prefix)
	state.CronExpression = types.StringValue(backup.CronExpression)
	state.DestinationID = types.StringValue(backup.DestinationID)

	if backup.ServiceType != nil {
		state.ServiceType = types.StringValue(*backup.ServiceType)
	}
	if backup.AppName != nil {
		state.AppName = types.StringValue(*backup.AppName)
	}
	if backup.ServiceName != nil {
		state.ServiceName = types.StringValue(*backup.ServiceName)
	}
	if backup.TurnOff != nil {
		state.TurnOff = types.BoolValue(*backup.TurnOff)
	}
	if backup.KeepLatestCount != nil {
		state.KeepLatestCount = types.Int64Value(int64(*backup.KeepLatestCount))
	}
	if backup.Enabled != nil {
		state.Enabled = types.BoolValue(*backup.Enabled)
	}
	if backup.ApplicationID != nil {
		state.ApplicationID = types.StringValue(*backup.ApplicationID)
	}
	if backup.PostgresID != nil {
		state.PostgresID = types.StringValue(*backup.PostgresID)
	}
	if backup.MysqlID != nil {
		state.MysqlID = types.StringValue(*backup.MysqlID)
	}
	if backup.MariadbID != nil {
		state.MariadbID = types.StringValue(*backup.MariadbID)
	}
	if backup.MongoID != nil {
		state.MongoID = types.StringValue(*backup.MongoID)
	}
	if backup.RedisID != nil {
		state.RedisID = types.StringValue(*backup.RedisID)
	}
	if backup.ComposeID != nil {
		state.ComposeID = types.StringValue(*backup.ComposeID)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeBackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan VolumeBackupResourceModel
	var state VolumeBackupResourceModel

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

	updateReq := client.UpdateVolumeBackupRequest{
		VolumeBackupID: state.ID.ValueString(),
		Name:           plan.Name.ValueString(),
		VolumeName:     plan.VolumeName.ValueString(),
		Prefix:         plan.Prefix.ValueString(),
		CronExpression: plan.CronExpression.ValueString(),
		DestinationID:  plan.DestinationID.ValueString(),
	}

	if !plan.ServiceType.IsNull() && !plan.ServiceType.IsUnknown() {
		v := plan.ServiceType.ValueString()
		updateReq.ServiceType = &v
	}
	if !plan.AppName.IsNull() && !plan.AppName.IsUnknown() {
		v := plan.AppName.ValueString()
		updateReq.AppName = &v
	}
	if !plan.ServiceName.IsNull() && !plan.ServiceName.IsUnknown() {
		v := plan.ServiceName.ValueString()
		updateReq.ServiceName = &v
	}
	if !plan.TurnOff.IsNull() && !plan.TurnOff.IsUnknown() {
		v := plan.TurnOff.ValueBool()
		updateReq.TurnOff = &v
	}
	if !plan.KeepLatestCount.IsNull() && !plan.KeepLatestCount.IsUnknown() {
		v := int(plan.KeepLatestCount.ValueInt64())
		updateReq.KeepLatestCount = &v
	}
	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		v := plan.Enabled.ValueBool()
		updateReq.Enabled = &v
	}
	if !plan.ApplicationID.IsNull() && !plan.ApplicationID.IsUnknown() {
		v := plan.ApplicationID.ValueString()
		updateReq.ApplicationID = &v
	}
	if !plan.PostgresID.IsNull() && !plan.PostgresID.IsUnknown() {
		v := plan.PostgresID.ValueString()
		updateReq.PostgresID = &v
	}
	if !plan.MysqlID.IsNull() && !plan.MysqlID.IsUnknown() {
		v := plan.MysqlID.ValueString()
		updateReq.MysqlID = &v
	}
	if !plan.MariadbID.IsNull() && !plan.MariadbID.IsUnknown() {
		v := plan.MariadbID.ValueString()
		updateReq.MariadbID = &v
	}
	if !plan.MongoID.IsNull() && !plan.MongoID.IsUnknown() {
		v := plan.MongoID.ValueString()
		updateReq.MongoID = &v
	}
	if !plan.RedisID.IsNull() && !plan.RedisID.IsUnknown() {
		v := plan.RedisID.ValueString()
		updateReq.RedisID = &v
	}
	if !plan.ComposeID.IsNull() && !plan.ComposeID.IsUnknown() {
		v := plan.ComposeID.ValueString()
		updateReq.ComposeID = &v
	}

	if err := r.client.UpdateVolumeBackup(ctx, updateReq); err != nil {
		resp.Diagnostics.AddError("Error Updating Volume Backup", err.Error())
		return
	}

	plan.ID = state.ID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *VolumeBackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state VolumeBackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteVolumeBackup(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Volume Backup", err.Error())
		return
	}
}

func (r *VolumeBackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
