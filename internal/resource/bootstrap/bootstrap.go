package bootstrap

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

var (
	_ resource.Resource                = &BootstrapResource{}
	_ resource.ResourceWithConfigure   = &BootstrapResource{}
	_ resource.ResourceWithImportState = &BootstrapResource{}
)

func NewResource() resource.Resource {
	return &BootstrapResource{}
}

type BootstrapResource struct {
	client *client.Client
}

type BootstrapResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Host           types.String `tfsdk:"host"`
	Email          types.String `tfsdk:"email"`
	Password       types.String `tfsdk:"password"`
	FirstName      types.String `tfsdk:"first_name"`
	LastName       types.String `tfsdk:"last_name"`
	APIKeyName     types.String `tfsdk:"api_key_name"`
	APIKey         types.String `tfsdk:"api_key"`
	UserID         types.String `tfsdk:"user_id"`
	OrganizationID types.String `tfsdk:"organization_id"`
}

func (r *BootstrapResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bootstrap"
}

func (r *BootstrapResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Bootstraps a fresh Dokploy instance by creating an admin user and API key. " +
			"This resource should only be used once per Dokploy instance to perform initial setup. " +
			"The generated api_key can then be used to configure the provider for other resources.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier (user ID).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"host": schema.StringAttribute{
				Description: "The Dokploy server URL (e.g., http://localhost:3000). Required for bootstrap since provider may not be configured yet.",
				Required:    true,
			},
			"email": schema.StringAttribute{
				Description: "Email address for the admin user.",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"password": schema.StringAttribute{
				Description: "Password for the admin user (min 8 characters, no special characters that need escaping).",
				Required:    true,
				Sensitive:   true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"first_name": schema.StringAttribute{
				Description: "First name of the admin user.",
				Required:    true,
			},
			"last_name": schema.StringAttribute{
				Description: "Last name of the admin user.",
				Required:    true,
			},
			"api_key_name": schema.StringAttribute{
				Description: "Name for the generated API key.",
				Optional:    true,
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"api_key": schema.StringAttribute{
				Description: "The generated API key. Use this to configure the Dokploy provider.",
				Computed:    true,
				Sensitive:   true,
			},
			"user_id": schema.StringAttribute{
				Description: "The ID of the created admin user.",
				Computed:    true,
			},
			"organization_id": schema.StringAttribute{
				Description: "The ID of the organization.",
				Computed:    true,
			},
		},
	}
}

func (r *BootstrapResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *BootstrapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan BootstrapResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	host := plan.Host.ValueString()
	httpClient := &http.Client{Timeout: 30 * time.Second}

	// Step 1: Register admin user
	registerReq := map[string]string{
		"email":    plan.Email.ValueString(),
		"password": plan.Password.ValueString(),
		"name":     plan.FirstName.ValueString(),
		"lastName": plan.LastName.ValueString(),
	}

	registerBody, _ := json.Marshal(registerReq)
	registerResp, err := httpClient.Post(
		host+"/api/auth/sign-up/email",
		"application/json",
		bytes.NewBuffer(registerBody),
	)
	if err != nil {
		resp.Diagnostics.AddError("Registration Failed", "Could not connect to Dokploy: "+err.Error())
		return
	}
	defer registerResp.Body.Close()

	// Extract session cookie
	var sessionCookie string
	for _, cookie := range registerResp.Cookies() {
		if cookie.Name == "better-auth.session_token" {
			sessionCookie = cookie.Value
			break
		}
	}

	if sessionCookie == "" {
		body, _ := io.ReadAll(registerResp.Body)
		resp.Diagnostics.AddError("Registration Failed", "No session cookie returned: "+string(body))
		return
	}

	// Parse registration response
	var registerResult struct {
		Token string `json:"token"`
		User  struct {
			ID    string `json:"id"`
			Email string `json:"email"`
		} `json:"user"`
	}
	registerResp.Body.Close()

	// Re-read with fresh request for user info
	// Step 2: Get organization ID
	userReq, _ := http.NewRequest("GET", host+"/api/user.get", nil)
	userReq.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: sessionCookie})

	userResp, err := httpClient.Do(userReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Get User", err.Error())
		return
	}
	defer userResp.Body.Close()

	var userResult struct {
		OrganizationID string `json:"organizationId"`
		UserID         string `json:"userId"`
	}
	if err := json.NewDecoder(userResp.Body).Decode(&userResult); err != nil {
		resp.Diagnostics.AddError("Failed to Parse User Response", err.Error())
		return
	}

	// Step 3: Create API key
	apiKeyName := plan.APIKeyName.ValueString()
	if apiKeyName == "" {
		apiKeyName = "terraform"
	}

	apiKeyReq := map[string]interface{}{
		"name":             apiKeyName,
		"rateLimitEnabled": false,
		"metadata": map[string]string{
			"organizationId": userResult.OrganizationID,
		},
	}
	apiKeyBody, _ := json.Marshal(apiKeyReq)

	createKeyReq, _ := http.NewRequest("POST", host+"/api/user.createApiKey", bytes.NewBuffer(apiKeyBody))
	createKeyReq.Header.Set("Content-Type", "application/json")
	createKeyReq.AddCookie(&http.Cookie{Name: "better-auth.session_token", Value: sessionCookie})

	createKeyResp, err := httpClient.Do(createKeyReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to Create API Key", err.Error())
		return
	}
	defer createKeyResp.Body.Close()

	var apiKeyResult struct {
		Key string `json:"key"`
	}
	if err := json.NewDecoder(createKeyResp.Body).Decode(&apiKeyResult); err != nil {
		body, _ := io.ReadAll(createKeyResp.Body)
		resp.Diagnostics.AddError("Failed to Parse API Key Response", string(body))
		return
	}

	if apiKeyResult.Key == "" {
		resp.Diagnostics.AddError("Failed to Create API Key", "No key returned")
		return
	}

	// Set state
	plan.ID = types.StringValue(userResult.UserID)
	plan.UserID = types.StringValue(userResult.UserID)
	plan.OrganizationID = types.StringValue(userResult.OrganizationID)
	plan.APIKey = types.StringValue(apiKeyResult.Key)
	plan.APIKeyName = types.StringValue(apiKeyName)
	_ = registerResult // silence unused warning

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BootstrapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state BootstrapResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Bootstrap is a one-time operation - just preserve state
	// We can't easily verify the user still exists without the session
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *BootstrapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan BootstrapResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Bootstrap doesn't support updates - all important fields require replace
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *BootstrapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Bootstrap deletion is a no-op - we don't delete the admin user
	// The user should manually clean up if needed
}

func (r *BootstrapResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
