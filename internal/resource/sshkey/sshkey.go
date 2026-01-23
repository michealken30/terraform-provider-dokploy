package sshkey

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &SSHKeyResource{}
	_ resource.ResourceWithConfigure   = &SSHKeyResource{}
	_ resource.ResourceWithImportState = &SSHKeyResource{}
)

func NewResource() resource.Resource {
	return &SSHKeyResource{}
}

type SSHKeyResource struct {
	client *client.Client
}

type SSHKeyResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	PrivateKey     types.String `tfsdk:"private_key"`
	PublicKey      types.String `tfsdk:"public_key"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (r *SSHKeyResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (r *SSHKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy SSH key for server connections.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the SSH key.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the SSH key.",
				Required:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the SSH key.",
				Optional:    true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key content (PEM format).",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"public_key": schema.StringAttribute{
				Description: "The public key content.",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID this SSH key belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SSHKeyResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SSHKeyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan SSHKeyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateSSHKeyRequest{
		Name:           plan.Name.ValueString(),
		PrivateKey:     plan.PrivateKey.ValueString(),
		PublicKey:      plan.PublicKey.ValueString(),
		OrganizationID: plan.OrganizationID.ValueString(),
	}

	if !plan.Description.IsNull() && !plan.Description.IsUnknown() && plan.Description.ValueString() != "" {
		desc := plan.Description.ValueString()
		createReq.Description = &desc
	}

	createResp, err := r.client.CreateSSHKey(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy SSH Key", "Could not create SSH key: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.SSHKeyID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SSHKeyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state SSHKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshKey, err := r.client.GetSSHKey(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy SSH Key", "Could not read SSH key ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Name = types.StringValue(sshKey.Name)
	state.PrivateKey = types.StringValue(sshKey.PrivateKey)
	state.PublicKey = types.StringValue(sshKey.PublicKey)
	state.OrganizationID = types.StringValue(sshKey.OrganizationID)

	if sshKey.Description != "" {
		state.Description = types.StringValue(sshKey.Description)
	} else {
		state.Description = types.StringNull()
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *SSHKeyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan SSHKeyResourceModel
	var state SSHKeyResourceModel

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

	updateReq := client.UpdateSSHKeyRequest{
		SSHKeyID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}
	if !plan.Description.Equal(state.Description) {
		desc := plan.Description.ValueString()
		updateReq.Description = &desc
	}

	err := r.client.UpdateSSHKey(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy SSH Key", "Could not update SSH key: "+err.Error())
		return
	}

	// Preserve computed and immutable fields from state
	plan.ID = state.ID
	plan.PrivateKey = state.PrivateKey
	plan.PublicKey = state.PublicKey
	plan.OrganizationID = state.OrganizationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *SSHKeyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state SSHKeyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteSSHKey(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy SSH Key", "Could not delete SSH key: "+err.Error())
		return
	}
}

func (r *SSHKeyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
