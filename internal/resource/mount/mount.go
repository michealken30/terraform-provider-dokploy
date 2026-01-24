package mount

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &MountResource{}
	_ resource.ResourceWithConfigure   = &MountResource{}
	_ resource.ResourceWithImportState = &MountResource{}
)

func NewResource() resource.Resource {
	return &MountResource{}
}

type MountResource struct {
	client *client.Client
}

type MountResourceModel struct {
	ID          types.String `tfsdk:"id"`
	Type        types.String `tfsdk:"type"`
	HostPath    types.String `tfsdk:"host_path"`
	VolumeName  types.String `tfsdk:"volume_name"`
	Content     types.String `tfsdk:"content"`
	FilePath    types.String `tfsdk:"file_path"`
	MountPath   types.String `tfsdk:"mount_path"`
	ServiceType types.String `tfsdk:"service_type"`
	ServiceID   types.String `tfsdk:"service_id"`
}

func (r *MountResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_mount"
}

func (r *MountResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy mount (volume, bind, or file) for a service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the mount.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Description: "The type of mount. Valid values: 'bind', 'volume', 'file'.",
				Required:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("bind", "volume", "file"),
				},
			},
			"host_path": schema.StringAttribute{
				Description: "The path on the host for bind mounts. Required when type is 'bind'.",
				Optional:    true,
			},
			"volume_name": schema.StringAttribute{
				Description: "The name of the Docker volume. Required when type is 'volume'.",
				Optional:    true,
			},
			"content": schema.StringAttribute{
				Description: "The file content for file mounts. Required when type is 'file'.",
				Optional:    true,
			},
			"file_path": schema.StringAttribute{
				Description: "The path where the file should be created for file mounts.",
				Optional:    true,
			},
			"mount_path": schema.StringAttribute{
				Description: "The path inside the container where the mount will be available.",
				Required:    true,
			},
			"service_type": schema.StringAttribute{
				Description: "The type of service this mount belongs to. Valid values: 'application', 'postgres', 'mysql', 'mariadb', 'mongo', 'redis', 'compose'. Defaults to 'application'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("application"),
				Validators: []validator.String{
					stringvalidator.OneOf("application", "postgres", "mysql", "mariadb", "mongo", "redis", "compose"),
				},
			},
			"service_id": schema.StringAttribute{
				Description: "The ID of the service (application, database, or compose) this mount belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *MountResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *MountResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan MountResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate type-specific requirements
	mountType := plan.Type.ValueString()
	switch mountType {
	case "bind":
		if plan.HostPath.IsNull() || plan.HostPath.IsUnknown() {
			resp.Diagnostics.AddError("Invalid Configuration", "host_path is required when type is 'bind'.")
			return
		}
	case "volume":
		if plan.VolumeName.IsNull() || plan.VolumeName.IsUnknown() {
			resp.Diagnostics.AddError("Invalid Configuration", "volume_name is required when type is 'volume'.")
			return
		}
	case "file":
		if plan.Content.IsNull() || plan.Content.IsUnknown() {
			resp.Diagnostics.AddError("Invalid Configuration", "content is required when type is 'file'.")
			return
		}
	}

	createReq := client.CreateMountRequest{
		Type:        mountType,
		MountPath:   plan.MountPath.ValueString(),
		ServiceType: plan.ServiceType.ValueString(),
		ServiceID:   plan.ServiceID.ValueString(),
	}

	if !plan.HostPath.IsNull() && !plan.HostPath.IsUnknown() {
		hostPath := plan.HostPath.ValueString()
		createReq.HostPath = &hostPath
	}

	if !plan.VolumeName.IsNull() && !plan.VolumeName.IsUnknown() {
		volumeName := plan.VolumeName.ValueString()
		createReq.VolumeName = &volumeName
	}

	if !plan.Content.IsNull() && !plan.Content.IsUnknown() {
		content := plan.Content.ValueString()
		createReq.Content = &content
	}

	if !plan.FilePath.IsNull() && !plan.FilePath.IsUnknown() {
		filePath := plan.FilePath.ValueString()
		createReq.FilePath = &filePath
	}

	createResp, err := r.client.CreateMount(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Mount", "Could not create mount: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.MountID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MountResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state MountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	mount, err := r.client.GetMount(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Mount", "Could not read mount ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(mount.MountID)
	state.Type = types.StringValue(mount.Type)
	state.MountPath = types.StringValue(mount.MountPath)
	state.ServiceType = types.StringValue(mount.ServiceType)

	// Determine the service ID based on which field is set
	if mount.ApplicationID != nil {
		state.ServiceID = types.StringValue(*mount.ApplicationID)
	} else if mount.PostgresID != nil {
		state.ServiceID = types.StringValue(*mount.PostgresID)
	} else if mount.MysqlID != nil {
		state.ServiceID = types.StringValue(*mount.MysqlID)
	} else if mount.MariadbID != nil {
		state.ServiceID = types.StringValue(*mount.MariadbID)
	} else if mount.MongoID != nil {
		state.ServiceID = types.StringValue(*mount.MongoID)
	} else if mount.RedisID != nil {
		state.ServiceID = types.StringValue(*mount.RedisID)
	} else if mount.ComposeID != nil {
		state.ServiceID = types.StringValue(*mount.ComposeID)
	}

	if mount.HostPath != nil {
		state.HostPath = types.StringValue(*mount.HostPath)
	} else {
		state.HostPath = types.StringNull()
	}

	if mount.VolumeName != nil {
		state.VolumeName = types.StringValue(*mount.VolumeName)
	} else {
		state.VolumeName = types.StringNull()
	}

	if mount.Content != nil {
		state.Content = types.StringValue(*mount.Content)
	} else {
		state.Content = types.StringNull()
	}

	if mount.FilePath != nil {
		state.FilePath = types.StringValue(*mount.FilePath)
	} else {
		state.FilePath = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *MountResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan MountResourceModel
	var state MountResourceModel

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

	updateReq := client.UpdateMountRequest{
		MountID: state.ID.ValueString(),
	}

	if !plan.Type.Equal(state.Type) {
		t := plan.Type.ValueString()
		updateReq.Type = &t
	}

	if !plan.MountPath.Equal(state.MountPath) {
		mp := plan.MountPath.ValueString()
		updateReq.MountPath = &mp
	}

	if !plan.ServiceType.Equal(state.ServiceType) {
		st := plan.ServiceType.ValueString()
		updateReq.ServiceType = &st
	}

	if !plan.HostPath.Equal(state.HostPath) {
		if !plan.HostPath.IsNull() && !plan.HostPath.IsUnknown() {
			hp := plan.HostPath.ValueString()
			updateReq.HostPath = &hp
		}
	}

	if !plan.VolumeName.Equal(state.VolumeName) {
		if !plan.VolumeName.IsNull() && !plan.VolumeName.IsUnknown() {
			vn := plan.VolumeName.ValueString()
			updateReq.VolumeName = &vn
		}
	}

	if !plan.Content.Equal(state.Content) {
		if !plan.Content.IsNull() && !plan.Content.IsUnknown() {
			c := plan.Content.ValueString()
			updateReq.Content = &c
		}
	}

	if !plan.FilePath.Equal(state.FilePath) {
		if !plan.FilePath.IsNull() && !plan.FilePath.IsUnknown() {
			fp := plan.FilePath.ValueString()
			updateReq.FilePath = &fp
		}
	}

	err := r.client.UpdateMount(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Mount", "Could not update mount: "+err.Error())
		return
	}

	// Preserve immutable fields from state
	plan.ID = state.ID
	plan.ServiceID = state.ServiceID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *MountResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state MountResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteMount(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Mount", "Could not delete mount: "+err.Error())
		return
	}
}

func (r *MountResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
