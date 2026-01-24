package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &RegistryResource{}
	_ resource.ResourceWithConfigure   = &RegistryResource{}
	_ resource.ResourceWithImportState = &RegistryResource{}
)

func NewResource() resource.Resource {
	return &RegistryResource{}
}

type RegistryResource struct {
	client *client.Client
}

type RegistryResourceModel struct {
	ID             types.String `tfsdk:"id"`
	RegistryName   types.String `tfsdk:"registry_name"`
	Username       types.String `tfsdk:"username"`
	Password       types.String `tfsdk:"password"`
	RegistryURL    types.String `tfsdk:"registry_url"`
	ImagePrefix    types.String `tfsdk:"image_prefix"`
	RegistryType   types.String `tfsdk:"registry_type"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (r *RegistryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *RegistryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy container registry configuration.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the registry.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"registry_name": schema.StringAttribute{
				Description: "The name of the registry.",
				Required:    true,
			},
			"username": schema.StringAttribute{
				Description: "The username for registry authentication.",
				Required:    true,
			},
			"password": schema.StringAttribute{
				Description: "The password for registry authentication.",
				Required:    true,
				Sensitive:   true,
			},
			"registry_url": schema.StringAttribute{
				Description: "The URL of the registry.",
				Required:    true,
			},
			"image_prefix": schema.StringAttribute{
				Description: "Optional prefix for images in this registry.",
				Optional:    true,
			},
			"registry_type": schema.StringAttribute{
				Description: "The type of registry (e.g., 'selfHosted', 'docker', 'github', 'gitlab', 'digitalocean', 'gcp', 'azure', 'ecr').",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID this registry belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *RegistryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *RegistryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan RegistryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateRegistryRequest{
		RegistryName:   plan.RegistryName.ValueString(),
		Username:       plan.Username.ValueString(),
		Password:       plan.Password.ValueString(),
		RegistryURL:    plan.RegistryURL.ValueString(),
		RegistryType:   plan.RegistryType.ValueString(),
		OrganizationID: plan.OrganizationID.ValueString(),
	}

	if !plan.ImagePrefix.IsNull() && !plan.ImagePrefix.IsUnknown() && plan.ImagePrefix.ValueString() != "" {
		prefix := plan.ImagePrefix.ValueString()
		createReq.ImagePrefix = &prefix
	}

	createResp, err := r.client.CreateRegistry(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Registry", "Could not create registry: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.RegistryID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state RegistryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.client.GetRegistry(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Registry", "Could not read registry ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.RegistryName = types.StringValue(registry.RegistryName)
	state.Username = types.StringValue(registry.Username)
	state.Password = types.StringValue(registry.Password)
	state.RegistryURL = types.StringValue(registry.RegistryURL)
	state.RegistryType = types.StringValue(registry.RegistryType)
	state.OrganizationID = types.StringValue(registry.OrganizationID)

	if registry.ImagePrefix != nil && *registry.ImagePrefix != "" {
		state.ImagePrefix = types.StringValue(*registry.ImagePrefix)
	} else {
		state.ImagePrefix = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan RegistryResourceModel
	var state RegistryResourceModel

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

	updateReq := client.UpdateRegistryRequest{
		RegistryID: state.ID.ValueString(),
	}

	if !plan.RegistryName.Equal(state.RegistryName) {
		name := plan.RegistryName.ValueString()
		updateReq.RegistryName = &name
	}
	if !plan.Username.Equal(state.Username) {
		username := plan.Username.ValueString()
		updateReq.Username = &username
	}
	if !plan.Password.Equal(state.Password) {
		password := plan.Password.ValueString()
		updateReq.Password = &password
	}
	if !plan.RegistryURL.Equal(state.RegistryURL) {
		url := plan.RegistryURL.ValueString()
		updateReq.RegistryURL = &url
	}
	if !plan.ImagePrefix.Equal(state.ImagePrefix) {
		prefix := plan.ImagePrefix.ValueString()
		updateReq.ImagePrefix = &prefix
	}

	err := r.client.UpdateRegistry(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Registry", "Could not update registry: "+err.Error())
		return
	}

	// Preserve computed and immutable fields from state
	plan.ID = state.ID
	plan.RegistryType = state.RegistryType
	plan.OrganizationID = state.OrganizationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *RegistryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state RegistryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteRegistry(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Registry", "Could not delete registry: "+err.Error())
		return
	}
}

func (r *RegistryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
