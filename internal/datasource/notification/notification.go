package notification

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &NotificationDataSource{}

func NewDataSource() datasource.DataSource {
	return &NotificationDataSource{}
}

type NotificationDataSource struct {
	client *client.Client
}

type NotificationDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	NotificationID   types.String `tfsdk:"notification_id"`
	Name             types.String `tfsdk:"name"`
	NotificationType types.String `tfsdk:"notification_type"`
	AppDeploy        types.Bool   `tfsdk:"app_deploy"`
	AppBuildError    types.Bool   `tfsdk:"app_build_error"`
	DatabaseBackup   types.Bool   `tfsdk:"database_backup"`
	DokployRestart   types.Bool   `tfsdk:"dokploy_restart"`
	DockerCleanup    types.Bool   `tfsdk:"docker_cleanup"`
}

func (d *NotificationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

func (d *NotificationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single notification configuration from Dokploy by ID or name.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the notification.",
				Computed:    true,
			},
			"notification_id": schema.StringAttribute{
				Description: "The notification ID to look up. Mutually exclusive with name.",
				Optional:    true,
			},
			"name": schema.StringAttribute{
				Description: "The notification name to look up. Mutually exclusive with notification_id.",
				Optional:    true,
				Computed:    true,
			},
			"notification_type": schema.StringAttribute{
				Description: "The type of notification (slack, discord, telegram, email).",
				Computed:    true,
			},
			"app_deploy": schema.BoolAttribute{
				Description: "Whether to notify on app deployment.",
				Computed:    true,
			},
			"app_build_error": schema.BoolAttribute{
				Description: "Whether to notify on app build errors.",
				Computed:    true,
			},
			"database_backup": schema.BoolAttribute{
				Description: "Whether to notify on database backup.",
				Computed:    true,
			},
			"dokploy_restart": schema.BoolAttribute{
				Description: "Whether to notify on Dokploy restart.",
				Computed:    true,
			},
			"docker_cleanup": schema.BoolAttribute{
				Description: "Whether to notify on Docker cleanup.",
				Computed:    true,
			},
		},
	}
}

func (d *NotificationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	c, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError("Unexpected Data Source Configure Type", fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData))
		return
	}
	d.client = c
}

func (d *NotificationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config NotificationDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.NotificationID.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either notification_id or name must be specified.")
		return
	}

	if !config.NotificationID.IsNull() && !config.Name.IsNull() {
		resp.Diagnostics.AddError("Conflicting Attributes", "Only one of notification_id or name can be specified, not both.")
		return
	}

	var notification *client.Notification
	var err error

	if !config.NotificationID.IsNull() {
		// Fetch all and find by ID
		notifications, fetchErr := d.client.GetNotifications(ctx)
		if fetchErr != nil {
			resp.Diagnostics.AddError("Unable to Read Dokploy Notifications", fetchErr.Error())
			return
		}
		id := config.NotificationID.ValueString()
		for i := range notifications {
			if notifications[i].NotificationID == id {
				notification = &notifications[i]
				break
			}
		}
		if notification == nil {
			err = fmt.Errorf("notification with ID %s not found", id)
		}
	} else {
		// Look up by name
		notifications, fetchErr := d.client.GetNotifications(ctx)
		if fetchErr != nil {
			resp.Diagnostics.AddError("Unable to Read Dokploy Notifications", fetchErr.Error())
			return
		}
		name := config.Name.ValueString()
		for i := range notifications {
			if notifications[i].Name == name {
				notification = &notifications[i]
				break
			}
		}
		if notification == nil {
			err = fmt.Errorf("notification with name %s not found", name)
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy Notification", err.Error())
		return
	}

	state := NotificationDataSourceModel{
		ID:               types.StringValue(notification.NotificationID),
		NotificationID:   types.StringValue(notification.NotificationID),
		Name:             types.StringValue(notification.Name),
		NotificationType: types.StringValue(notification.NotificationType),
		AppDeploy:        types.BoolValue(notification.AppDeploy),
		AppBuildError:    types.BoolValue(notification.AppBuildError),
		DatabaseBackup:   types.BoolValue(notification.DatabaseBackup),
		DokployRestart:   types.BoolValue(notification.DokployRestart),
		DockerCleanup:    types.BoolValue(notification.DockerCleanup),
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
