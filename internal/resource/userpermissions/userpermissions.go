package userpermissions

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &UserPermissionsResource{}
	_ resource.ResourceWithConfigure   = &UserPermissionsResource{}
	_ resource.ResourceWithImportState = &UserPermissionsResource{}
)

func NewResource() resource.Resource {
	return &UserPermissionsResource{}
}

type UserPermissionsResource struct {
	client *client.Client
}

type UserPermissionsResourceModel struct {
	ID                      types.String `tfsdk:"id"`
	UserID                  types.String `tfsdk:"user_id"`
	AccessedProjects        types.List   `tfsdk:"accessed_projects"`
	AccessedEnvironments    types.List   `tfsdk:"accessed_environments"`
	AccessedServices        types.List   `tfsdk:"accessed_services"`
	CanCreateProjects       types.Bool   `tfsdk:"can_create_projects"`
	CanCreateServices       types.Bool   `tfsdk:"can_create_services"`
	CanDeleteProjects       types.Bool   `tfsdk:"can_delete_projects"`
	CanDeleteServices       types.Bool   `tfsdk:"can_delete_services"`
	CanAccessToDocker       types.Bool   `tfsdk:"can_access_to_docker"`
	CanAccessToTraefikFiles types.Bool   `tfsdk:"can_access_to_traefik_files"`
	CanAccessToAPI          types.Bool   `tfsdk:"can_access_to_api"`
	CanAccessToSSHKeys      types.Bool   `tfsdk:"can_access_to_ssh_keys"`
	CanAccessToGitProviders types.Bool   `tfsdk:"can_access_to_git_providers"`
	CanDeleteEnvironments   types.Bool   `tfsdk:"can_delete_environments"`
	CanCreateEnvironments   types.Bool   `tfsdk:"can_create_environments"`
}

func (r *UserPermissionsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user_permissions"
}

func (r *UserPermissionsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages permissions for a Dokploy user. Users must exist before permissions can be assigned.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier (same as user_id).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the user to assign permissions to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"accessed_projects": schema.ListAttribute{
				Description: "List of project IDs the user can access.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"accessed_environments": schema.ListAttribute{
				Description: "List of environment IDs the user can access.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"accessed_services": schema.ListAttribute{
				Description: "List of service IDs the user can access.",
				Optional:    true,
				ElementType: types.StringType,
			},
			"can_create_projects": schema.BoolAttribute{
				Description: "Whether the user can create projects.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_create_services": schema.BoolAttribute{
				Description: "Whether the user can create services.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_delete_projects": schema.BoolAttribute{
				Description: "Whether the user can delete projects.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_delete_services": schema.BoolAttribute{
				Description: "Whether the user can delete services.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_access_to_docker": schema.BoolAttribute{
				Description: "Whether the user can access Docker.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_access_to_traefik_files": schema.BoolAttribute{
				Description: "Whether the user can access Traefik configuration files.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_access_to_api": schema.BoolAttribute{
				Description: "Whether the user can access the API.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_access_to_ssh_keys": schema.BoolAttribute{
				Description: "Whether the user can access SSH keys.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_access_to_git_providers": schema.BoolAttribute{
				Description: "Whether the user can access Git providers.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_delete_environments": schema.BoolAttribute{
				Description: "Whether the user can delete environments.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"can_create_environments": schema.BoolAttribute{
				Description: "Whether the user can create environments.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
		},
	}
}

func (r *UserPermissionsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *UserPermissionsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan UserPermissionsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify user exists
	userID := plan.UserID.ValueString()
	_, err := r.client.GetUser(ctx, userID)
	if err != nil {
		resp.Diagnostics.AddError("User Not Found", fmt.Sprintf("Cannot assign permissions to non-existent user: %s", err.Error()))
		return
	}

	// Build permissions request
	permReq, diags := r.buildPermissionsRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AssignUserPermissions(ctx, permReq); err != nil {
		resp.Diagnostics.AddError("Error Assigning User Permissions", err.Error())
		return
	}

	plan.ID = plan.UserID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserPermissionsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state UserPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := r.client.GetUser(ctx, state.UserID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading User Permissions", err.Error())
		return
	}

	state.ID = types.StringValue(user.ID)
	state.UserID = types.StringValue(user.ID)

	// Map permissions from user
	if user.AccessedProjects != nil {
		projects, diags := types.ListValueFrom(ctx, types.StringType, user.AccessedProjects)
		resp.Diagnostics.Append(diags...)
		state.AccessedProjects = projects
	}
	if user.AccessedEnvironments != nil {
		envs, diags := types.ListValueFrom(ctx, types.StringType, user.AccessedEnvironments)
		resp.Diagnostics.Append(diags...)
		state.AccessedEnvironments = envs
	}
	if user.AccessedServices != nil {
		services, diags := types.ListValueFrom(ctx, types.StringType, user.AccessedServices)
		resp.Diagnostics.Append(diags...)
		state.AccessedServices = services
	}

	if user.CanCreateProjects != nil {
		state.CanCreateProjects = types.BoolValue(*user.CanCreateProjects)
	}
	if user.CanCreateServices != nil {
		state.CanCreateServices = types.BoolValue(*user.CanCreateServices)
	}
	if user.CanDeleteProjects != nil {
		state.CanDeleteProjects = types.BoolValue(*user.CanDeleteProjects)
	}
	if user.CanDeleteServices != nil {
		state.CanDeleteServices = types.BoolValue(*user.CanDeleteServices)
	}
	if user.CanAccessToDocker != nil {
		state.CanAccessToDocker = types.BoolValue(*user.CanAccessToDocker)
	}
	if user.CanAccessToTraefikFiles != nil {
		state.CanAccessToTraefikFiles = types.BoolValue(*user.CanAccessToTraefikFiles)
	}
	if user.CanAccessToAPI != nil {
		state.CanAccessToAPI = types.BoolValue(*user.CanAccessToAPI)
	}
	if user.CanAccessToSSHKeys != nil {
		state.CanAccessToSSHKeys = types.BoolValue(*user.CanAccessToSSHKeys)
	}
	if user.CanAccessToGitProviders != nil {
		state.CanAccessToGitProviders = types.BoolValue(*user.CanAccessToGitProviders)
	}
	if user.CanDeleteEnvironments != nil {
		state.CanDeleteEnvironments = types.BoolValue(*user.CanDeleteEnvironments)
	}
	if user.CanCreateEnvironments != nil {
		state.CanCreateEnvironments = types.BoolValue(*user.CanCreateEnvironments)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *UserPermissionsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan UserPermissionsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	permReq, diags := r.buildPermissionsRequest(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.client.AssignUserPermissions(ctx, permReq); err != nil {
		resp.Diagnostics.AddError("Error Updating User Permissions", err.Error())
		return
	}

	plan.ID = plan.UserID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *UserPermissionsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state UserPermissionsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Reset permissions to empty/false on delete
	permReq := client.UserPermissionsRequest{
		ID:                      state.UserID.ValueString(),
		AccessedProjects:        []string{},
		AccessedEnvironments:    []string{},
		AccessedServices:        []string{},
		CanCreateProjects:       false,
		CanCreateServices:       false,
		CanDeleteProjects:       false,
		CanDeleteServices:       false,
		CanAccessToDocker:       false,
		CanAccessToTraefikFiles: false,
		CanAccessToAPI:          false,
		CanAccessToSSHKeys:      false,
		CanAccessToGitProviders: false,
		CanDeleteEnvironments:   false,
		CanCreateEnvironments:   false,
	}

	if err := r.client.AssignUserPermissions(ctx, permReq); err != nil {
		resp.Diagnostics.AddError("Error Resetting User Permissions", err.Error())
		return
	}
}

func (r *UserPermissionsResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import by user ID
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("user_id"), req.ID)...)
}

func (r *UserPermissionsResource) buildPermissionsRequest(ctx context.Context, plan UserPermissionsResourceModel) (client.UserPermissionsRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	var accessedProjects []string
	if !plan.AccessedProjects.IsNull() && !plan.AccessedProjects.IsUnknown() {
		d := plan.AccessedProjects.ElementsAs(ctx, &accessedProjects, false)
		diags.Append(d...)
	}
	if accessedProjects == nil {
		accessedProjects = []string{}
	}

	var accessedEnvironments []string
	if !plan.AccessedEnvironments.IsNull() && !plan.AccessedEnvironments.IsUnknown() {
		d := plan.AccessedEnvironments.ElementsAs(ctx, &accessedEnvironments, false)
		diags.Append(d...)
	}
	if accessedEnvironments == nil {
		accessedEnvironments = []string{}
	}

	var accessedServices []string
	if !plan.AccessedServices.IsNull() && !plan.AccessedServices.IsUnknown() {
		d := plan.AccessedServices.ElementsAs(ctx, &accessedServices, false)
		diags.Append(d...)
	}
	if accessedServices == nil {
		accessedServices = []string{}
	}

	return client.UserPermissionsRequest{
		ID:                      plan.UserID.ValueString(),
		AccessedProjects:        accessedProjects,
		AccessedEnvironments:    accessedEnvironments,
		AccessedServices:        accessedServices,
		CanCreateProjects:       plan.CanCreateProjects.ValueBool(),
		CanCreateServices:       plan.CanCreateServices.ValueBool(),
		CanDeleteProjects:       plan.CanDeleteProjects.ValueBool(),
		CanDeleteServices:       plan.CanDeleteServices.ValueBool(),
		CanAccessToDocker:       plan.CanAccessToDocker.ValueBool(),
		CanAccessToTraefikFiles: plan.CanAccessToTraefikFiles.ValueBool(),
		CanAccessToAPI:          plan.CanAccessToAPI.ValueBool(),
		CanAccessToSSHKeys:      plan.CanAccessToSSHKeys.ValueBool(),
		CanAccessToGitProviders: plan.CanAccessToGitProviders.ValueBool(),
		CanDeleteEnvironments:   plan.CanDeleteEnvironments.ValueBool(),
		CanCreateEnvironments:   plan.CanCreateEnvironments.ValueBool(),
	}, diags
}
