package domain

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &DomainResource{}
	_ resource.ResourceWithConfigure   = &DomainResource{}
	_ resource.ResourceWithImportState = &DomainResource{}
)

func NewResource() resource.Resource {
	return &DomainResource{}
}

type DomainResource struct {
	client *client.Client
}

type DomainResourceModel struct {
	ID                 types.String `tfsdk:"id"`
	Host               types.String `tfsdk:"host"`
	Path               types.String `tfsdk:"path"`
	Port               types.Int64  `tfsdk:"port"`
	HTTPS              types.Bool   `tfsdk:"https"`
	CertificateType    types.String `tfsdk:"certificate_type"`
	CustomCertResolver types.String `tfsdk:"custom_cert_resolver"`
	ApplicationID      types.String `tfsdk:"application_id"`
	ComposeID          types.String `tfsdk:"compose_id"`
	ServiceName        types.String `tfsdk:"service_name"`
	DomainType         types.String `tfsdk:"domain_type"`
	InternalPath       types.String `tfsdk:"internal_path"`
	StripPath          types.Bool   `tfsdk:"strip_path"`
	UniqueConfigKey    types.Int64  `tfsdk:"unique_config_key"`
}

func (r *DomainResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_domain"
}

func (r *DomainResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy domain for an application or compose service.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the domain.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				Description: "The domain hostname (e.g., 'example.com' or 'app.example.com').",
				Required:    true,
			},
			"path": schema.StringAttribute{
				Description: "The URL path prefix for the domain (e.g., '/api'). Defaults to '/'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("/"),
			},
			"port": schema.Int64Attribute{
				Description: "The target port to route traffic to. Defaults to the application's exposed port.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"https": schema.BoolAttribute{
				Description: "Whether to enable HTTPS for this domain.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"certificate_type": schema.StringAttribute{
				Description: "The type of SSL certificate to use. Valid values: 'letsencrypt', 'none', 'custom'.",
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("none"),
				Validators: []validator.String{
					stringvalidator.OneOf("letsencrypt", "none", "custom"),
				},
			},
			"custom_cert_resolver": schema.StringAttribute{
				Description: "Custom certificate resolver name when certificate_type is 'custom'.",
				Optional:    true,
			},
			"application_id": schema.StringAttribute{
				Description: "The ID of the application this domain belongs to. Either application_id or compose_id must be set.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"compose_id": schema.StringAttribute{
				Description: "The ID of the compose service this domain belongs to. Either application_id or compose_id must be set.",
				Optional:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"service_name": schema.StringAttribute{
				Description: "The service name within a compose stack. Required when compose_id is set.",
				Optional:    true,
			},
			"domain_type": schema.StringAttribute{
				Description: "The type of domain. Valid values: 'compose', 'application', 'preview'. Auto-detected if not set.",
				Optional:    true,
				Computed:    true,
				Validators: []validator.String{
					stringvalidator.OneOf("compose", "application", "preview"),
				},
			},
			"internal_path": schema.StringAttribute{
				Description: "Internal path for routing within the application.",
				Optional:    true,
				Computed:    true,
			},
			"strip_path": schema.BoolAttribute{
				Description: "Whether to strip the path prefix before forwarding to the application.",
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
			},
			"unique_config_key": schema.Int64Attribute{
				Description: "Unique configuration key assigned by Dokploy.",
				Computed:    true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *DomainResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DomainResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DomainResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either application_id or compose_id is set
	if plan.ApplicationID.IsNull() && plan.ComposeID.IsNull() {
		resp.Diagnostics.AddError("Invalid Configuration", "Either application_id or compose_id must be set.")
		return
	}

	createReq := client.CreateDomainRequest{
		Host:      plan.Host.ValueString(),
		HTTPS:     plan.HTTPS.ValueBool(),
		StripPath: plan.StripPath.ValueBool(),
	}

	if !plan.Path.IsNull() && !plan.Path.IsUnknown() {
		path := plan.Path.ValueString()
		createReq.Path = &path
	}

	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port := int(plan.Port.ValueInt64())
		createReq.Port = &port
	}

	if !plan.CertificateType.IsNull() && !plan.CertificateType.IsUnknown() {
		createReq.CertificateType = plan.CertificateType.ValueString()
	}

	if !plan.CustomCertResolver.IsNull() && !plan.CustomCertResolver.IsUnknown() {
		resolver := plan.CustomCertResolver.ValueString()
		createReq.CustomCertResolver = &resolver
	}

	if !plan.ApplicationID.IsNull() {
		appID := plan.ApplicationID.ValueString()
		createReq.ApplicationID = &appID
	}

	if !plan.ComposeID.IsNull() {
		composeID := plan.ComposeID.ValueString()
		createReq.ComposeID = &composeID
	}

	if !plan.ServiceName.IsNull() && !plan.ServiceName.IsUnknown() {
		serviceName := plan.ServiceName.ValueString()
		createReq.ServiceName = &serviceName
	}

	if !plan.DomainType.IsNull() && !plan.DomainType.IsUnknown() {
		domainType := plan.DomainType.ValueString()
		createReq.DomainType = &domainType
	}

	if !plan.InternalPath.IsNull() && !plan.InternalPath.IsUnknown() {
		internalPath := plan.InternalPath.ValueString()
		createReq.InternalPath = &internalPath
	}

	createResp, err := r.client.CreateDomain(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Domain", "Could not create domain: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.DomainID)

	// Read back to get computed values
	domain, err := r.client.GetDomain(ctx, createResp.DomainID)
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Domain", "Could not read created domain: "+err.Error())
		return
	}

	r.mapDomainToState(domain, &plan)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	domain, err := r.client.GetDomain(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Domain", "Could not read domain ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	r.mapDomainToState(domain, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DomainResourceModel
	var state DomainResourceModel

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

	updateReq := client.UpdateDomainRequest{
		DomainID:        state.ID.ValueString(),
		Host:            plan.Host.ValueString(),
		HTTPS:           plan.HTTPS.ValueBool(),
		CertificateType: plan.CertificateType.ValueString(),
		StripPath:       plan.StripPath.ValueBool(),
	}

	if !plan.Path.IsNull() && !plan.Path.IsUnknown() {
		path := plan.Path.ValueString()
		updateReq.Path = &path
	}

	if !plan.Port.IsNull() && !plan.Port.IsUnknown() {
		port := int(plan.Port.ValueInt64())
		updateReq.Port = &port
	}

	if !plan.CustomCertResolver.IsNull() && !plan.CustomCertResolver.IsUnknown() {
		resolver := plan.CustomCertResolver.ValueString()
		updateReq.CustomCertResolver = &resolver
	}

	if !plan.ServiceName.IsNull() && !plan.ServiceName.IsUnknown() {
		serviceName := plan.ServiceName.ValueString()
		updateReq.ServiceName = &serviceName
	}

	if !plan.DomainType.IsNull() && !plan.DomainType.IsUnknown() {
		domainType := plan.DomainType.ValueString()
		updateReq.DomainType = &domainType
	}

	if !plan.InternalPath.IsNull() && !plan.InternalPath.IsUnknown() {
		internalPath := plan.InternalPath.ValueString()
		updateReq.InternalPath = &internalPath
	}

	err := r.client.UpdateDomain(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Domain", "Could not update domain: "+err.Error())
		return
	}

	// Preserve computed and immutable fields from state
	plan.ID = state.ID
	plan.ApplicationID = state.ApplicationID
	plan.ComposeID = state.ComposeID
	plan.UniqueConfigKey = state.UniqueConfigKey

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DomainResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DomainResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDomain(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Domain", "Could not delete domain: "+err.Error())
		return
	}
}

func (r *DomainResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *DomainResource) mapDomainToState(domain *client.Domain, state *DomainResourceModel) {
	state.ID = types.StringValue(domain.DomainID)
	state.Host = types.StringValue(domain.Host)
	state.Path = types.StringValue(domain.Path)
	state.HTTPS = types.BoolValue(domain.HTTPS)
	state.CertificateType = types.StringValue(domain.CertificateType)
	state.StripPath = types.BoolValue(domain.StripPath)
	state.UniqueConfigKey = types.Int64Value(int64(domain.UniqueConfigKey))

	if domain.Port != nil {
		state.Port = types.Int64Value(int64(*domain.Port))
	} else {
		state.Port = types.Int64Null()
	}

	if domain.CustomCertResolver != nil {
		state.CustomCertResolver = types.StringValue(*domain.CustomCertResolver)
	} else {
		state.CustomCertResolver = types.StringNull()
	}

	if domain.ApplicationID != nil {
		state.ApplicationID = types.StringValue(*domain.ApplicationID)
	} else {
		state.ApplicationID = types.StringNull()
	}

	if domain.ComposeID != nil {
		state.ComposeID = types.StringValue(*domain.ComposeID)
	} else {
		state.ComposeID = types.StringNull()
	}

	if domain.ServiceName != nil {
		state.ServiceName = types.StringValue(*domain.ServiceName)
	} else {
		state.ServiceName = types.StringNull()
	}

	if domain.DomainType != nil {
		state.DomainType = types.StringValue(*domain.DomainType)
	} else {
		state.DomainType = types.StringNull()
	}

	if domain.InternalPath != nil {
		state.InternalPath = types.StringValue(*domain.InternalPath)
	} else {
		state.InternalPath = types.StringNull()
	}
}
