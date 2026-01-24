package certificate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &CertificateResource{}
	_ resource.ResourceWithConfigure   = &CertificateResource{}
	_ resource.ResourceWithImportState = &CertificateResource{}
)

func NewResource() resource.Resource {
	return &CertificateResource{}
}

type CertificateResource struct {
	client *client.Client
}

type CertificateResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	CertificateData types.String `tfsdk:"certificate_data"`
	PrivateKey      types.String `tfsdk:"private_key"`
	CertificatePath types.String `tfsdk:"certificate_path"`
	AutoRenew       types.Bool   `tfsdk:"auto_renew"`
	OrganizationID  types.String `tfsdk:"organization_id"`
}

func (r *CertificateResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

func (r *CertificateResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy SSL certificate.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the certificate.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the certificate.",
				Required:    true,
			},
			"certificate_data": schema.StringAttribute{
				Description: "The certificate data (PEM format).",
				Required:    true,
				Sensitive:   true,
			},
			"private_key": schema.StringAttribute{
				Description: "The private key for the certificate (PEM format).",
				Required:    true,
				Sensitive:   true,
			},
			"certificate_path": schema.StringAttribute{
				Description: "The path where the certificate is stored. Computed by Dokploy.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"auto_renew": schema.BoolAttribute{
				Description: "Whether to auto-renew the certificate.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID this certificate belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *CertificateResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CertificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan CertificateResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateCertificateRequest{
		Name:            plan.Name.ValueString(),
		CertificateData: plan.CertificateData.ValueString(),
		PrivateKey:      plan.PrivateKey.ValueString(),
		OrganizationID:  plan.OrganizationID.ValueString(),
	}

	if !plan.AutoRenew.IsNull() && !plan.AutoRenew.IsUnknown() {
		autoRenew := plan.AutoRenew.ValueBool()
		createReq.AutoRenew = &autoRenew
	}

	createResp, err := r.client.CreateCertificate(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Certificate", "Could not create certificate: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.CertificateID)

	// Read back to get computed values
	cert, err := r.client.GetCertificate(ctx, createResp.CertificateID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Certificate", "Could not read created certificate: "+err.Error())
		return
	}

	plan.CertificatePath = types.StringValue(cert.CertificatePath)
	plan.AutoRenew = types.BoolValue(cert.AutoRenew)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state CertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	cert, err := r.client.GetCertificate(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Certificate", "Could not read certificate ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Name = types.StringValue(cert.Name)
	state.CertificateData = types.StringValue(cert.CertificateData)
	state.PrivateKey = types.StringValue(cert.PrivateKey)
	state.CertificatePath = types.StringValue(cert.CertificatePath)
	state.AutoRenew = types.BoolValue(cert.AutoRenew)
	state.OrganizationID = types.StringValue(cert.OrganizationID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan CertificateResourceModel
	var state CertificateResourceModel

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

	updateReq := client.UpdateCertificateRequest{
		CertificateID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}
	if !plan.CertificateData.Equal(state.CertificateData) {
		certData := plan.CertificateData.ValueString()
		updateReq.CertificateData = &certData
	}
	if !plan.PrivateKey.Equal(state.PrivateKey) {
		privateKey := plan.PrivateKey.ValueString()
		updateReq.PrivateKey = &privateKey
	}
	if !plan.AutoRenew.Equal(state.AutoRenew) {
		autoRenew := plan.AutoRenew.ValueBool()
		updateReq.AutoRenew = &autoRenew
	}

	err := r.client.UpdateCertificate(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Certificate", "Could not update certificate: "+err.Error())
		return
	}

	// Preserve computed and immutable fields from state
	plan.ID = state.ID
	plan.CertificatePath = state.CertificatePath
	plan.OrganizationID = state.OrganizationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *CertificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state CertificateResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCertificate(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Certificate", "Could not delete certificate: "+err.Error())
		return
	}
}

func (r *CertificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
