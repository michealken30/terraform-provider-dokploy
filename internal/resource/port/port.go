package port

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

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &PortResource{}
	_ resource.ResourceWithConfigure   = &PortResource{}
	_ resource.ResourceWithImportState = &PortResource{}
)

func NewResource() resource.Resource {
	return &PortResource{}
}

type PortResource struct {
	client *client.Client
}

type PortResourceModel struct {
	ID            types.String `tfsdk:"id"`
	PublishedPort types.Int64  `tfsdk:"published_port"`
	TargetPort    types.Int64  `tfsdk:"target_port"`
	Protocol      types.String `tfsdk:"protocol"`
	PublishMode   types.String `tfsdk:"publish_mode"`
	ApplicationID types.String `tfsdk:"application_id"`
}

func (r *PortResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_port"
}

func (r *PortResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy port mapping for an application.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the port mapping.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"published_port": schema.Int64Attribute{
				Description: "The external port to expose on the host.",
				Required:    true,
			},
			"target_port": schema.Int64Attribute{
				Description: "The internal container port to map to.",
				Required:    true,
			},
			"protocol": schema.StringAttribute{
				Description: "The network protocol. Valid values: 'tcp', 'udp'. Defaults to 'tcp'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("tcp"),
				Validators: []validator.String{
					stringvalidator.OneOf("tcp", "udp"),
				},
			},
			"publish_mode": schema.StringAttribute{
				Description: "The publish mode for swarm. Valid values: 'ingress', 'host'. Defaults to 'ingress'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ingress"),
				Validators: []validator.String{
					stringvalidator.OneOf("ingress", "host"),
				},
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this port mapping belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *PortResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PortResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan PortResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreatePortRequest{
		PublishedPort: int(plan.PublishedPort.ValueInt64()),
		TargetPort:    int(plan.TargetPort.ValueInt64()),
		Protocol:      plan.Protocol.ValueString(),
		PublishMode:   plan.PublishMode.ValueString(),
		ApplicationID: plan.ApplicationID.ValueString(),
	}

	createResp, err := r.client.CreatePort(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Port", "Could not create port: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.PortID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PortResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state PortResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	port, err := r.client.GetPort(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Port", "Could not read port ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.ID = types.StringValue(port.PortID)
	state.PublishedPort = types.Int64Value(int64(port.PublishedPort))
	state.TargetPort = types.Int64Value(int64(port.TargetPort))
	state.Protocol = types.StringValue(port.Protocol)
	state.PublishMode = types.StringValue(port.PublishMode)

	if port.ApplicationID != nil {
		state.ApplicationID = types.StringValue(*port.ApplicationID)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *PortResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan PortResourceModel
	var state PortResourceModel

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

	updateReq := client.UpdatePortRequest{
		PortID:        state.ID.ValueString(),
		PublishedPort: int(plan.PublishedPort.ValueInt64()),
		TargetPort:    int(plan.TargetPort.ValueInt64()),
		Protocol:      plan.Protocol.ValueString(),
		PublishMode:   plan.PublishMode.ValueString(),
	}

	err := r.client.UpdatePort(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Port", "Could not update port: "+err.Error())
		return
	}

	// Preserve immutable fields from state
	plan.ID = state.ID
	plan.ApplicationID = state.ApplicationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *PortResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state PortResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePort(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Port", "Could not delete port: "+err.Error())
		return
	}
}

func (r *PortResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
