package certificates

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/reserve-protocol/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &CertificatesDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &CertificatesDataSource{}
}

// CertificatesDataSource is the data source implementation.
type CertificatesDataSource struct {
	client *client.Client
}

// CertificatesDataSourceModel describes the data source data model.
type CertificatesDataSourceModel struct {
	Certificates []CertificateModel `tfsdk:"certificates"`
}

// CertificateModel describes a single certificate (without private key for security).
type CertificateModel struct {
	ID             types.String `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	AutoRenew      types.Bool   `tfsdk:"auto_renew"`
	OrganizationID types.String `tfsdk:"organization_id"`
	CreatedAt      types.String `tfsdk:"created_at"`
}

func (d *CertificatesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificates"
}

func (d *CertificatesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches all SSL certificates from Dokploy. Use this data source to discover certificate IDs for importing existing resources.",
		Attributes: map[string]schema.Attribute{
			"certificates": schema.ListNestedAttribute{
				Description: "List of all SSL certificates.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the certificate.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "The name of the certificate.",
							Computed:    true,
						},
						"auto_renew": schema.BoolAttribute{
							Description: "Whether the certificate auto-renews.",
							Computed:    true,
						},
						"organization_id": schema.StringAttribute{
							Description: "The organization ID the certificate belongs to.",
							Computed:    true,
						},
						"created_at": schema.StringAttribute{
							Description: "The creation timestamp of the certificate.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *CertificatesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *CertificatesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CertificatesDataSourceModel

	certificates, err := d.client.GetCertificates(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Certificates",
			err.Error(),
		)
		return
	}

	// Initialize to empty slice to avoid null in state
	state.Certificates = []CertificateModel{}

	for _, cert := range certificates {
		state.Certificates = append(state.Certificates, CertificateModel{
			ID:             types.StringValue(cert.CertificateID),
			Name:           types.StringValue(cert.Name),
			AutoRenew:      types.BoolValue(cert.AutoRenew),
			OrganizationID: types.StringValue(cert.OrganizationID),
			CreatedAt:      types.StringValue(cert.CreatedAt.Format("2006-01-02T15:04:05Z")),
		})
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
