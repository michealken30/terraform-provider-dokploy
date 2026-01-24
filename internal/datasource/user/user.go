package user

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &UserDataSource{}

func NewDataSource() datasource.DataSource {
	return &UserDataSource{}
}

type UserDataSource struct {
	client *client.Client
}

type UserDataSourceModel struct {
	ID               types.String `tfsdk:"id"`
	UserID           types.String `tfsdk:"user_id"`
	Email            types.String `tfsdk:"email"`
	Name             types.String `tfsdk:"name"`
	Image            types.String `tfsdk:"image"`
	Role             types.String `tfsdk:"role"`
	IsRegistered     types.Bool   `tfsdk:"is_registered"`
	EmailVerified    types.Bool   `tfsdk:"email_verified"`
	TwoFactorEnabled types.Bool   `tfsdk:"two_factor_enabled"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (d *UserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *UserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single user from Dokploy by ID or email.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the user.",
				Computed:    true,
			},
			"user_id": schema.StringAttribute{
				Description: "The user ID to look up. Mutually exclusive with email.",
				Optional:    true,
			},
			"email": schema.StringAttribute{
				Description: "The email to look up. Mutually exclusive with user_id.",
				Optional:    true,
				Computed:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the user.",
				Computed:    true,
			},
			"image": schema.StringAttribute{
				Description: "The profile image URL of the user.",
				Computed:    true,
			},
			"role": schema.StringAttribute{
				Description: "The role of the user.",
				Computed:    true,
			},
			"is_registered": schema.BoolAttribute{
				Description: "Whether the user has completed registration.",
				Computed:    true,
			},
			"email_verified": schema.BoolAttribute{
				Description: "Whether the user's email is verified.",
				Computed:    true,
			},
			"two_factor_enabled": schema.BoolAttribute{
				Description: "Whether two-factor authentication is enabled.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp of the user.",
				Computed:    true,
			},
		},
	}
}

func (d *UserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config UserDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.UserID.IsNull() && config.Email.IsNull() {
		resp.Diagnostics.AddError("Missing Required Attribute", "Either user_id or email must be specified.")
		return
	}

	if !config.UserID.IsNull() && !config.Email.IsNull() {
		resp.Diagnostics.AddError("Conflicting Attributes", "Only one of user_id or email can be specified, not both.")
		return
	}

	var user *client.User
	var err error

	if !config.UserID.IsNull() {
		user, err = d.client.GetUser(ctx, config.UserID.ValueString())
	} else {
		// Look up by email
		users, fetchErr := d.client.GetUsers(ctx)
		if fetchErr != nil {
			resp.Diagnostics.AddError("Unable to Read Dokploy Users", fetchErr.Error())
			return
		}
		email := config.Email.ValueString()
		for i := range users {
			if users[i].Email == email {
				user = &users[i]
				break
			}
		}
		if user == nil {
			err = fmt.Errorf("user with email %s not found", email)
		}
	}

	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy User", err.Error())
		return
	}

	state := UserDataSourceModel{
		ID:            types.StringValue(user.ID),
		UserID:        types.StringValue(user.ID),
		Email:         types.StringValue(user.Email),
		Name:          types.StringValue(user.Name),
		IsRegistered:  types.BoolValue(user.IsRegistered),
		EmailVerified: types.BoolValue(user.EmailVerified),
	}

	if user.Image != nil {
		state.Image = types.StringValue(*user.Image)
	} else {
		state.Image = types.StringNull()
	}

	if user.Role != nil {
		state.Role = types.StringValue(*user.Role)
	} else {
		state.Role = types.StringNull()
	}

	if user.TwoFactorEnabled != nil {
		state.TwoFactorEnabled = types.BoolValue(*user.TwoFactorEnabled)
	} else {
		state.TwoFactorEnabled = types.BoolNull()
	}

	if user.CreatedAt != nil {
		state.CreatedAt = types.StringValue(*user.CreatedAt)
	} else {
		state.CreatedAt = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
