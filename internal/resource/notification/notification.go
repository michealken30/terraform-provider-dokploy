package notification

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &NotificationResource{}
	_ resource.ResourceWithConfigure   = &NotificationResource{}
	_ resource.ResourceWithImportState = &NotificationResource{}
)

func NewResource() resource.Resource {
	return &NotificationResource{}
}

type NotificationResource struct {
	client *client.Client
}

type NotificationResourceModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	NotificationType types.String `tfsdk:"notification_type"`
	AppDeploy        types.Bool   `tfsdk:"app_deploy"`
	AppBuildError    types.Bool   `tfsdk:"app_build_error"`
	DatabaseBackup   types.Bool   `tfsdk:"database_backup"`
	VolumeBackup     types.Bool   `tfsdk:"volume_backup"`
	DokployRestart   types.Bool   `tfsdk:"dokploy_restart"`
	DockerCleanup    types.Bool   `tfsdk:"docker_cleanup"`
	ServerThreshold  types.Bool   `tfsdk:"server_threshold"`
	// Slack
	SlackWebhookURL types.String `tfsdk:"slack_webhook_url"`
	SlackChannel    types.String `tfsdk:"slack_channel"`
	// Discord
	DiscordWebhookURL types.String `tfsdk:"discord_webhook_url"`
	DiscordDecoration types.Bool   `tfsdk:"discord_decoration"`
	// Telegram
	TelegramBotToken        types.String `tfsdk:"telegram_bot_token"`
	TelegramChatID          types.String `tfsdk:"telegram_chat_id"`
	TelegramMessageThreadID types.String `tfsdk:"telegram_message_thread_id"`
	// Email
	EmailSmtpServer  types.String `tfsdk:"email_smtp_server"`
	EmailSmtpPort    types.Int64  `tfsdk:"email_smtp_port"`
	EmailUsername    types.String `tfsdk:"email_username"`
	EmailPassword    types.String `tfsdk:"email_password"`
	EmailFromAddress types.String `tfsdk:"email_from_address"`
	EmailToAddresses types.List   `tfsdk:"email_to_addresses"`
}

func (r *NotificationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_notification"
}

func (r *NotificationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy notification configuration (Slack, Discord, Telegram, or Email).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the notification.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the notification.",
				Required:    true,
			},
			"notification_type": schema.StringAttribute{
				Description: "The type of notification. Valid values: 'slack', 'discord', 'telegram', 'email'.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("slack", "discord", "telegram", "email"),
				},
			},
			// Event triggers
			"app_deploy": schema.BoolAttribute{
				Description: "Notify on application deployment.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"app_build_error": schema.BoolAttribute{
				Description: "Notify on application build errors.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"database_backup": schema.BoolAttribute{
				Description: "Notify on database backup completion.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"volume_backup": schema.BoolAttribute{
				Description: "Notify on volume backup completion.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"dokploy_restart": schema.BoolAttribute{
				Description: "Notify on Dokploy restart.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"docker_cleanup": schema.BoolAttribute{
				Description: "Notify on Docker cleanup.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"server_threshold": schema.BoolAttribute{
				Description: "Notify when server thresholds are exceeded.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			// Slack-specific
			"slack_webhook_url": schema.StringAttribute{
				Description: "Slack webhook URL. Required when notification_type is 'slack'.",
				Optional:    true,
				Sensitive:   true,
			},
			"slack_channel": schema.StringAttribute{
				Description: "Slack channel name. Required when notification_type is 'slack'.",
				Optional:    true,
			},
			// Discord-specific
			"discord_webhook_url": schema.StringAttribute{
				Description: "Discord webhook URL. Required when notification_type is 'discord'.",
				Optional:    true,
				Sensitive:   true,
			},
			"discord_decoration": schema.BoolAttribute{
				Description: "Enable Discord message decoration. Required when notification_type is 'discord'.",
				Optional:    true,
			},
			// Telegram-specific
			"telegram_bot_token": schema.StringAttribute{
				Description: "Telegram bot token. Required when notification_type is 'telegram'.",
				Optional:    true,
				Sensitive:   true,
			},
			"telegram_chat_id": schema.StringAttribute{
				Description: "Telegram chat ID. Required when notification_type is 'telegram'.",
				Optional:    true,
			},
			"telegram_message_thread_id": schema.StringAttribute{
				Description: "Telegram message thread ID for topics.",
				Optional:    true,
			},
			// Email-specific
			"email_smtp_server": schema.StringAttribute{
				Description: "SMTP server hostname. Required when notification_type is 'email'.",
				Optional:    true,
			},
			"email_smtp_port": schema.Int64Attribute{
				Description: "SMTP server port. Required when notification_type is 'email'.",
				Optional:    true,
			},
			"email_username": schema.StringAttribute{
				Description: "SMTP username. Required when notification_type is 'email'.",
				Optional:    true,
			},
			"email_password": schema.StringAttribute{
				Description: "SMTP password. Required when notification_type is 'email'.",
				Optional:    true,
				Sensitive:   true,
			},
			"email_from_address": schema.StringAttribute{
				Description: "From email address. Required when notification_type is 'email'.",
				Optional:    true,
			},
			"email_to_addresses": schema.ListAttribute{
				Description: "List of recipient email addresses. Required when notification_type is 'email'.",
				Optional:    true,
				ElementType: types.StringType,
			},
		},
	}
}

func (r *NotificationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *NotificationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan NotificationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notificationType := plan.NotificationType.ValueString()
	var notificationID string

	base := client.NotificationBase{
		Name:            plan.Name.ValueString(),
		AppDeploy:       plan.AppDeploy.ValueBool(),
		AppBuildError:   plan.AppBuildError.ValueBool(),
		DatabaseBackup:  plan.DatabaseBackup.ValueBool(),
		VolumeBackup:    plan.VolumeBackup.ValueBool(),
		DokployRestart:  plan.DokployRestart.ValueBool(),
		DockerCleanup:   plan.DockerCleanup.ValueBool(),
		ServerThreshold: plan.ServerThreshold.ValueBool(),
	}

	switch notificationType {
	case "slack":
		createReq := client.CreateSlackNotificationRequest{
			NotificationBase: base,
			WebhookURL:       plan.SlackWebhookURL.ValueString(),
			Channel:          plan.SlackChannel.ValueString(),
		}
		createResp, err := r.client.CreateSlackNotification(ctx, createReq)
		if err != nil {
			resp.Diagnostics.AddError("Error Creating Slack Notification", err.Error())
			return
		}
		notificationID = createResp.NotificationID

	case "discord":
		createReq := client.CreateDiscordNotificationRequest{
			NotificationBase: base,
			WebhookURL:       plan.DiscordWebhookURL.ValueString(),
			Decoration:       plan.DiscordDecoration.ValueBool(),
		}
		createResp, err := r.client.CreateDiscordNotification(ctx, createReq)
		if err != nil {
			resp.Diagnostics.AddError("Error Creating Discord Notification", err.Error())
			return
		}
		notificationID = createResp.NotificationID

	case "telegram":
		createReq := client.CreateTelegramNotificationRequest{
			NotificationBase: base,
			BotToken:         plan.TelegramBotToken.ValueString(),
			ChatID:           plan.TelegramChatID.ValueString(),
			MessageThreadID:  plan.TelegramMessageThreadID.ValueString(),
		}
		createResp, err := r.client.CreateTelegramNotification(ctx, createReq)
		if err != nil {
			resp.Diagnostics.AddError("Error Creating Telegram Notification", err.Error())
			return
		}
		notificationID = createResp.NotificationID

	case "email":
		var toAddresses []string
		diags := plan.EmailToAddresses.ElementsAs(ctx, &toAddresses, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		createReq := client.CreateEmailNotificationRequest{
			NotificationBase: base,
			SmtpServer:       plan.EmailSmtpServer.ValueString(),
			SmtpPort:         int(plan.EmailSmtpPort.ValueInt64()),
			Username:         plan.EmailUsername.ValueString(),
			Password:         plan.EmailPassword.ValueString(),
			FromAddress:      plan.EmailFromAddress.ValueString(),
			ToAddresses:      toAddresses,
		}
		createResp, err := r.client.CreateEmailNotification(ctx, createReq)
		if err != nil {
			resp.Diagnostics.AddError("Error Creating Email Notification", err.Error())
			return
		}
		notificationID = createResp.NotificationID
	}

	plan.ID = types.StringValue(notificationID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state NotificationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	notification, err := r.client.GetNotification(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Notification", err.Error())
		return
	}

	state.ID = types.StringValue(notification.NotificationID)
	state.Name = types.StringValue(notification.Name)
	state.NotificationType = types.StringValue(notification.NotificationType)
	state.AppDeploy = types.BoolValue(notification.AppDeploy)
	state.AppBuildError = types.BoolValue(notification.AppBuildError)
	state.DatabaseBackup = types.BoolValue(notification.DatabaseBackup)
	state.VolumeBackup = types.BoolValue(notification.VolumeBackup)
	state.DokployRestart = types.BoolValue(notification.DokployRestart)
	state.DockerCleanup = types.BoolValue(notification.DockerCleanup)
	state.ServerThreshold = types.BoolValue(notification.ServerThreshold)

	// Type-specific fields
	switch notification.NotificationType {
	case "slack":
		if notification.WebhookURL != nil {
			state.SlackWebhookURL = types.StringValue(*notification.WebhookURL)
		}
		if notification.Channel != nil {
			state.SlackChannel = types.StringValue(*notification.Channel)
		}
	case "discord":
		if notification.WebhookURL != nil {
			state.DiscordWebhookURL = types.StringValue(*notification.WebhookURL)
		}
		if notification.Decoration != nil {
			state.DiscordDecoration = types.BoolValue(*notification.Decoration)
		}
	case "telegram":
		if notification.BotToken != nil {
			state.TelegramBotToken = types.StringValue(*notification.BotToken)
		}
		if notification.ChatID != nil {
			state.TelegramChatID = types.StringValue(*notification.ChatID)
		}
		if notification.MessageThreadID != nil {
			state.TelegramMessageThreadID = types.StringValue(*notification.MessageThreadID)
		}
	case "email":
		if notification.SmtpServer != nil {
			state.EmailSmtpServer = types.StringValue(*notification.SmtpServer)
		}
		if notification.SmtpPort != nil {
			state.EmailSmtpPort = types.Int64Value(int64(*notification.SmtpPort))
		}
		if notification.Username != nil {
			state.EmailUsername = types.StringValue(*notification.Username)
		}
		if notification.Password != nil {
			state.EmailPassword = types.StringValue(*notification.Password)
		}
		if notification.FromAddress != nil {
			state.EmailFromAddress = types.StringValue(*notification.FromAddress)
		}
		if len(notification.ToAddresses) > 0 {
			elements := make([]attr.Value, len(notification.ToAddresses))
			for i, addr := range notification.ToAddresses {
				elements[i] = types.StringValue(addr)
			}
			state.EmailToAddresses, _ = types.ListValue(types.StringType, elements)
		}
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan NotificationResourceModel
	var state NotificationResourceModel

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

	notificationType := plan.NotificationType.ValueString()
	notificationID := state.ID.ValueString()

	base := client.NotificationBase{
		Name:            plan.Name.ValueString(),
		AppDeploy:       plan.AppDeploy.ValueBool(),
		AppBuildError:   plan.AppBuildError.ValueBool(),
		DatabaseBackup:  plan.DatabaseBackup.ValueBool(),
		VolumeBackup:    plan.VolumeBackup.ValueBool(),
		DokployRestart:  plan.DokployRestart.ValueBool(),
		DockerCleanup:   plan.DockerCleanup.ValueBool(),
		ServerThreshold: plan.ServerThreshold.ValueBool(),
	}

	switch notificationType {
	case "slack":
		updateReq := client.UpdateSlackNotificationRequest{
			NotificationID: notificationID,
			CreateSlackNotificationRequest: client.CreateSlackNotificationRequest{
				NotificationBase: base,
				WebhookURL:       plan.SlackWebhookURL.ValueString(),
				Channel:          plan.SlackChannel.ValueString(),
			},
		}
		if err := r.client.UpdateSlackNotification(ctx, updateReq); err != nil {
			resp.Diagnostics.AddError("Error Updating Slack Notification", err.Error())
			return
		}

	case "discord":
		updateReq := client.UpdateDiscordNotificationRequest{
			NotificationID: notificationID,
			CreateDiscordNotificationRequest: client.CreateDiscordNotificationRequest{
				NotificationBase: base,
				WebhookURL:       plan.DiscordWebhookURL.ValueString(),
				Decoration:       plan.DiscordDecoration.ValueBool(),
			},
		}
		if err := r.client.UpdateDiscordNotification(ctx, updateReq); err != nil {
			resp.Diagnostics.AddError("Error Updating Discord Notification", err.Error())
			return
		}

	case "telegram":
		updateReq := client.UpdateTelegramNotificationRequest{
			NotificationID: notificationID,
			CreateTelegramNotificationRequest: client.CreateTelegramNotificationRequest{
				NotificationBase: base,
				BotToken:         plan.TelegramBotToken.ValueString(),
				ChatID:           plan.TelegramChatID.ValueString(),
				MessageThreadID:  plan.TelegramMessageThreadID.ValueString(),
			},
		}
		if err := r.client.UpdateTelegramNotification(ctx, updateReq); err != nil {
			resp.Diagnostics.AddError("Error Updating Telegram Notification", err.Error())
			return
		}

	case "email":
		var toAddresses []string
		diags := plan.EmailToAddresses.ElementsAs(ctx, &toAddresses, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}

		updateReq := client.UpdateEmailNotificationRequest{
			NotificationID: notificationID,
			CreateEmailNotificationRequest: client.CreateEmailNotificationRequest{
				NotificationBase: base,
				SmtpServer:       plan.EmailSmtpServer.ValueString(),
				SmtpPort:         int(plan.EmailSmtpPort.ValueInt64()),
				Username:         plan.EmailUsername.ValueString(),
				Password:         plan.EmailPassword.ValueString(),
				FromAddress:      plan.EmailFromAddress.ValueString(),
				ToAddresses:      toAddresses,
			},
		}
		if err := r.client.UpdateEmailNotification(ctx, updateReq); err != nil {
			resp.Diagnostics.AddError("Error Updating Email Notification", err.Error())
			return
		}
	}

	plan.ID = state.ID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *NotificationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state NotificationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.DeleteNotification(ctx, state.ID.ValueString()); err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Notification", err.Error())
		return
	}
}

func (r *NotificationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
