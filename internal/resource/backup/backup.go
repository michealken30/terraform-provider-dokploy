package backup

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
	_ resource.Resource                = &BackupResource{}
	_ resource.ResourceWithConfigure   = &BackupResource{}
	_ resource.ResourceWithImportState = &BackupResource{}
)

func NewResource() resource.Resource {
	return &BackupResource{}
}

type BackupResource struct {
	client *client.Client
}

type BackupResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Schedule        types.String `tfsdk:"schedule"`
	Enabled         types.Bool   `tfsdk:"enabled"`
	Prefix          types.String `tfsdk:"prefix"`
	DestinationID   types.String `tfsdk:"destination_id"`
	KeepLatestCount types.Int64  `tfsdk:"keep_latest_count"`
	Database        types.String `tfsdk:"database"`
	DatabaseType    types.String `tfsdk:"database_type"`
	BackupType      types.String `tfsdk:"backup_type"`
	PostgresID      types.String `tfsdk:"postgres_id"`
	MysqlID         types.String `tfsdk:"mysql_id"`
	MariadbID       types.String `tfsdk:"mariadb_id"`
	MongoID         types.String `tfsdk:"mongo_id"`
	ComposeID       types.String `tfsdk:"compose_id"`
	ServiceName     types.String `tfsdk:"service_name"`
}

func (r *BackupResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (r *BackupResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy backup configuration for a database or compose service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the backup.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"schedule": schema.StringAttribute{
				Description: "Cron expression for the backup schedule (e.g., '0 0 * * *' for daily at midnight).",
				Required:    true,
			},
			"enabled": schema.BoolAttribute{
				Description: "Whether the backup is enabled. Defaults to true.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
			},
			"prefix": schema.StringAttribute{
				Description: "Prefix for backup file names.",
				Required:    true,
			},
			"destination_id": schema.StringAttribute{
				Description: "The ID of the backup destination (S3 bucket, etc.).",
				Required:    true,
			},
			"keep_latest_count": schema.Int64Attribute{
				Description: "Number of recent backups to keep. Older backups will be deleted.",
				Optional:    true,
			},
			"database": schema.StringAttribute{
				Description: "The database name to backup.",
				Required:    true,
			},
			"database_type": schema.StringAttribute{
				Description: "The type of database. Valid values: 'postgres', 'mysql', 'mariadb', 'mongo', 'web-server'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("postgres", "mysql", "mariadb", "mongo", "web-server"),
				},
			},
			"backup_type": schema.StringAttribute{
				Description: "The type of backup. Valid values: 'database', 'compose'. Defaults to 'database'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("database"),
				Validators: []validator.String{
					stringvalidator.OneOf("database", "compose"),
				},
			},
			"postgres_id": schema.StringAttribute{
				Description: "The ID of the Postgres service to backup. Required when database_type is 'postgres'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mysql_id": schema.StringAttribute{
				Description: "The ID of the MySQL service to backup. Required when database_type is 'mysql'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mariadb_id": schema.StringAttribute{
				Description: "The ID of the MariaDB service to backup. Required when database_type is 'mariadb'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"mongo_id": schema.StringAttribute{
				Description: "The ID of the MongoDB service to backup. Required when database_type is 'mongo'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose_id": schema.StringAttribute{
				Description: "The ID of the compose service to backup. Required when backup_type is 'compose'.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_name": schema.StringAttribute{
				Description: "The service name within a compose stack. Required when compose_id is set.",
				Optional:    true,
			},
		},
	}
}

func (r *BackupResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BackupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BackupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateBackupRequest{
		Schedule:      plan.Schedule.ValueString(),
		Prefix:        plan.Prefix.ValueString(),
		DestinationID: plan.DestinationID.ValueString(),
		Database:      plan.Database.ValueString(),
		DatabaseType:  plan.DatabaseType.ValueString(),
		BackupType:    plan.BackupType.ValueString(),
	}

	if !plan.Enabled.IsNull() && !plan.Enabled.IsUnknown() {
		enabled := plan.Enabled.ValueBool()
		createReq.Enabled = &enabled
	}

	if !plan.KeepLatestCount.IsNull() && !plan.KeepLatestCount.IsUnknown() {
		count := int(plan.KeepLatestCount.ValueInt64())
		createReq.KeepLatestCount = &count
	}

	if !plan.PostgresID.IsNull() && !plan.PostgresID.IsUnknown() {
		id := plan.PostgresID.ValueString()
		createReq.PostgresID = &id
	}

	if !plan.MysqlID.IsNull() && !plan.MysqlID.IsUnknown() {
		id := plan.MysqlID.ValueString()
		createReq.MysqlID = &id
	}

	if !plan.MariadbID.IsNull() && !plan.MariadbID.IsUnknown() {
		id := plan.MariadbID.ValueString()
		createReq.MariadbID = &id
	}

	if !plan.MongoID.IsNull() && !plan.MongoID.IsUnknown() {
		id := plan.MongoID.ValueString()
		createReq.MongoID = &id
	}

	if !plan.ComposeID.IsNull() && !plan.ComposeID.IsUnknown() {
		id := plan.ComposeID.ValueString()
		createReq.ComposeID = &id
	}

	if !plan.ServiceName.IsNull() && !plan.ServiceName.IsUnknown() {
		name := plan.ServiceName.ValueString()
		createReq.ServiceName = &name
	}

	createResp, err := r.client.CreateBackup(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Backup", "Could not create backup: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.BackupID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := r.client.GetBackup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Backup", "Could not read backup ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(backup.BackupID)
	state.Schedule = types.StringValue(backup.Schedule)
	state.Enabled = types.BoolValue(backup.Enabled)
	state.Prefix = types.StringValue(backup.Prefix)
	state.DestinationID = types.StringValue(backup.DestinationID)
	state.Database = types.StringValue(backup.Database)
	state.DatabaseType = types.StringValue(backup.DatabaseType)
	state.BackupType = types.StringValue(backup.BackupType)

	if backup.KeepLatestCount != nil {
		state.KeepLatestCount = types.Int64Value(int64(*backup.KeepLatestCount))
	} else {
		state.KeepLatestCount = types.Int64Null()
	}

	if backup.PostgresID != nil {
		state.PostgresID = types.StringValue(*backup.PostgresID)
	} else {
		state.PostgresID = types.StringNull()
	}

	if backup.MysqlID != nil {
		state.MysqlID = types.StringValue(*backup.MysqlID)
	} else {
		state.MysqlID = types.StringNull()
	}

	if backup.MariadbID != nil {
		state.MariadbID = types.StringValue(*backup.MariadbID)
	} else {
		state.MariadbID = types.StringNull()
	}

	if backup.MongoID != nil {
		state.MongoID = types.StringValue(*backup.MongoID)
	} else {
		state.MongoID = types.StringNull()
	}

	if backup.ComposeID != nil {
		state.ComposeID = types.StringValue(*backup.ComposeID)
	} else {
		state.ComposeID = types.StringNull()
	}

	if backup.ServiceName != nil {
		state.ServiceName = types.StringValue(*backup.ServiceName)
	} else {
		state.ServiceName = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *BackupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BackupResourceModel
	var state BackupResourceModel

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

	updateReq := client.UpdateBackupRequest{
		BackupID: state.ID.ValueString(),
	}

	if !plan.Schedule.Equal(state.Schedule) {
		schedule := plan.Schedule.ValueString()
		updateReq.Schedule = &schedule
	}

	if !plan.Enabled.Equal(state.Enabled) {
		enabled := plan.Enabled.ValueBool()
		updateReq.Enabled = &enabled
	}

	if !plan.Prefix.Equal(state.Prefix) {
		prefix := plan.Prefix.ValueString()
		updateReq.Prefix = &prefix
	}

	if !plan.DestinationID.Equal(state.DestinationID) {
		destID := plan.DestinationID.ValueString()
		updateReq.DestinationID = &destID
	}

	if !plan.KeepLatestCount.Equal(state.KeepLatestCount) {
		if !plan.KeepLatestCount.IsNull() && !plan.KeepLatestCount.IsUnknown() {
			count := int(plan.KeepLatestCount.ValueInt64())
			updateReq.KeepLatestCount = &count
		}
	}

	if !plan.Database.Equal(state.Database) {
		db := plan.Database.ValueString()
		updateReq.Database = &db
	}

	err := r.client.UpdateBackup(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Backup", "Could not update backup: "+err.Error())
		return
	}

	// Preserve immutable fields from state
	plan.ID = state.ID
	plan.PostgresID = state.PostgresID
	plan.MysqlID = state.MysqlID
	plan.MariadbID = state.MariadbID
	plan.MongoID = state.MongoID
	plan.ComposeID = state.ComposeID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BackupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state BackupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteBackup(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Backup", "Could not delete backup: "+err.Error())
		return
	}
}

func (r *BackupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
