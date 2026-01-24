package users

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var _ datasource.DataSource = &UsersDataSource{}

func NewDataSource() datasource.DataSource {
	return &UsersDataSource{}
}

type UsersDataSource struct {
	client *client.Client
}

type UsersDataSourceModel struct {
	Users []UserModel `tfsdk:"users"`
}

type UserModel struct {
	ID               types.String `tfsdk:"id"`
	Name             types.String `tfsdk:"name"`
	Email            types.String `tfsdk:"email"`
	Image            types.String `tfsdk:"image"`
	Role             types.String `tfsdk:"role"`
	IsRegistered     types.Bool   `tfsdk:"is_registered"`
	EmailVerified    types.Bool   `tfsdk:"email_verified"`
	TwoFactorEnabled types.Bool   `tfsdk:"two_factor_enabled"`
	CreatedAt        types.String `tfsdk:"created_at"`
}

func (d *UsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_users"
}

func (d *UsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all users from Dokploy.",
		Attributes: map[string]schema.Attribute{
			"users": schema.ListNestedAttribute{
				Description: "List of all users.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the user.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the user.",
							Computed:    true,
						},
						"email": schema.StringAttribute{
							Description: "The email of the user.",
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
				},
			},
		},
	}
}

func (d *UsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *UsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state UsersDataSourceModel

	users, err := d.client.GetUsers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Unable to Read Dokploy Users", err.Error())
		return
	}

	state.Users = []UserModel{}
	for _, user := range users {
		userModel := UserModel{
			ID:            types.StringValue(user.ID),
			Name:          types.StringValue(user.Name),
			Email:         types.StringValue(user.Email),
			IsRegistered:  types.BoolValue(user.IsRegistered),
			EmailVerified: types.BoolValue(user.EmailVerified),
		}

		if user.Image != nil {
			userModel.Image = types.StringValue(*user.Image)
		} else {
			userModel.Image = types.StringNull()
		}

		if user.Role != nil {
			userModel.Role = types.StringValue(*user.Role)
		} else {
			userModel.Role = types.StringNull()
		}

		if user.TwoFactorEnabled != nil {
			userModel.TwoFactorEnabled = types.BoolValue(*user.TwoFactorEnabled)
		} else {
			userModel.TwoFactorEnabled = types.BoolNull()
		}

		if user.CreatedAt != nil {
			userModel.CreatedAt = types.StringValue(*user.CreatedAt)
		} else {
			userModel.CreatedAt = types.StringNull()
		}

		state.Users = append(state.Users, userModel)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
