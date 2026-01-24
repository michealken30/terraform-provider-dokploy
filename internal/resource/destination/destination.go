package destination

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
	_ resource.Resource                = &DestinationResource{}
	_ resource.ResourceWithConfigure   = &DestinationResource{}
	_ resource.ResourceWithImportState = &DestinationResource{}
)

func NewResource() resource.Resource {
	return &DestinationResource{}
}

type DestinationResource struct {
	client *client.Client
}

type DestinationResourceModel struct {
	ID              types.String `tfsdk:"id"`
	Name            types.String `tfsdk:"name"`
	AccessKey       types.String `tfsdk:"access_key"`
	SecretAccessKey types.String `tfsdk:"secret_access_key"`
	Bucket          types.String `tfsdk:"bucket"`
	Region          types.String `tfsdk:"region"`
	Endpoint        types.String `tfsdk:"endpoint"`
	OrganizationID  types.String `tfsdk:"organization_id"`
}

func (r *DestinationResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_destination"
}

func (r *DestinationResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a Dokploy backup destination (S3-compatible storage).",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the destination.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Description: "The name of the destination.",
				Required:    true,
			},
			"access_key": schema.StringAttribute{
				Description: "The access key for the S3-compatible storage.",
				Required:    true,
				Sensitive:   true,
			},
			"secret_access_key": schema.StringAttribute{
				Description: "The secret access key for the S3-compatible storage.",
				Required:    true,
				Sensitive:   true,
			},
			"bucket": schema.StringAttribute{
				Description: "The bucket name.",
				Required:    true,
			},
			"region": schema.StringAttribute{
				Description: "The region of the S3-compatible storage.",
				Required:    true,
			},
			"endpoint": schema.StringAttribute{
				Description: "The endpoint URL of the S3-compatible storage.",
				Required:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The organization ID this destination belongs to.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *DestinationResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *DestinationResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan DestinationResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createReq := client.CreateDestinationRequest{
		Name:            plan.Name.ValueString(),
		AccessKey:       plan.AccessKey.ValueString(),
		SecretAccessKey: plan.SecretAccessKey.ValueString(),
		Bucket:          plan.Bucket.ValueString(),
		Region:          plan.Region.ValueString(),
		Endpoint:        plan.Endpoint.ValueString(),
		OrganizationID:  plan.OrganizationID.ValueString(),
	}

	createResp, err := r.client.CreateDestination(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Creating Dokploy Destination", "Could not create destination: "+err.Error())
		return
	}

	plan.ID = types.StringValue(createResp.DestinationID)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state DestinationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	dest, err := r.client.GetDestination(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Reading Dokploy Destination", "Could not read destination ID "+state.ID.ValueString()+": "+err.Error())
		return
	}

	state.Name = types.StringValue(dest.Name)
	state.AccessKey = types.StringValue(dest.AccessKey)
	state.SecretAccessKey = types.StringValue(dest.SecretAccessKey)
	state.Bucket = types.StringValue(dest.Bucket)
	state.Region = types.StringValue(dest.Region)
	state.Endpoint = types.StringValue(dest.Endpoint)
	state.OrganizationID = types.StringValue(dest.OrganizationID)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan DestinationResourceModel
	var state DestinationResourceModel

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

	updateReq := client.UpdateDestinationRequest{
		DestinationID: state.ID.ValueString(),
	}

	if !plan.Name.Equal(state.Name) {
		name := plan.Name.ValueString()
		updateReq.Name = &name
	}
	if !plan.AccessKey.Equal(state.AccessKey) {
		accessKey := plan.AccessKey.ValueString()
		updateReq.AccessKey = &accessKey
	}
	if !plan.SecretAccessKey.Equal(state.SecretAccessKey) {
		secretKey := plan.SecretAccessKey.ValueString()
		updateReq.SecretAccessKey = &secretKey
	}
	if !plan.Bucket.Equal(state.Bucket) {
		bucket := plan.Bucket.ValueString()
		updateReq.Bucket = &bucket
	}
	if !plan.Region.Equal(state.Region) {
		region := plan.Region.ValueString()
		updateReq.Region = &region
	}
	if !plan.Endpoint.Equal(state.Endpoint) {
		endpoint := plan.Endpoint.ValueString()
		updateReq.Endpoint = &endpoint
	}

	err := r.client.UpdateDestination(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error Updating Dokploy Destination", "Could not update destination: "+err.Error())
		return
	}

	// Preserve computed and immutable fields from state
	plan.ID = state.ID
	plan.OrganizationID = state.OrganizationID

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *DestinationResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state DestinationResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteDestination(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error Deleting Dokploy Destination", "Could not delete destination: "+err.Error())
		return
	}
}

func (r *DestinationResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
