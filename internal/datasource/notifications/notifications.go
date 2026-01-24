package notifications

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &NotificationsDataSource{}

func NewDataSource() datasource.DataSource {
	return &NotificationsDataSource{}
}

type NotificationsDataSource struct {
	client *client.Client
}

type NotificationsDataSourceModel struct {
	Notifications []NotificationModel `tfsdk:"notifications"`
}

type NotificationModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	NotificationType types.String `tfsdk:"notification_type"`
	AppDeploy        types.Bool   `tfsdk:"app_deploy"`
	AppBuildError    types.Bool   `tfsdk:"app_build_error"`
	DatabaseBackup   types.Bool   `tfsdk:"database_backup"`
	DokployRestart   types.Bool   `tfsdk:"dokploy_restart"`
	DockerCleanup    types.Bool   `tfsdk:"docker_cleanup"`
}

func (d *NotificationsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notifications"
}

func (d *NotificationsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all notification configurations from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"notifications": schema.ListNestedAttribute{
				Description: "List of all notifications.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the notification.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the notification.",
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
				},
			},
		},
	}
}

func (d *NotificationsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *NotificationsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state NotificationsDataSourceModel

	notifications, err := d.client.GetNotifications(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy Notifications", err.Error())
		return
	}

	state.Notifications = []NotificationModel{}
	for _, n := range notifications {
		state.Notifications = append(state.Notifications, NotificationModel{
			ID:               types.StringValue(n.NotificationID),
			Name:             types.StringValue(n.Name),
			NotificationType: types.StringValue(n.NotificationType),
			AppDeploy:        types.BoolValue(n.AppDeploy),
			AppBuildError:    types.BoolValue(n.AppBuildError),
			DatabaseBackup:   types.BoolValue(n.DatabaseBackup),
			DokployRestart:   types.BoolValue(n.DokployRestart),
			DockerCleanup:    types.BoolValue(n.DockerCleanup),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
