package application

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thefrozenfire/terraform-provider-dokploy/internal/client"
)

// Ensure the implementation satisfies the expected interfaces.
var _ datasource.DataSource = &ApplicationDataSource{}

// NewDataSource is a helper function to simplify the provider implementation.
func NewDataSource() datasource.DataSource {
	return &ApplicationDataSource{}
}

// ApplicationDataSource is the data source implementation.
type ApplicationDataSource struct {
	client *client.Client
}

// ApplicationDataSourceModel describes the data source data model.
type ApplicationDataSourceModel struct {
	ID            types.String  `tfsdk:"id"`
	ApplicationID types.String  `tfsdk:"application_id"`
	Name          types.String  `tfsdk:"name"`
	AppName       types.String  `tfsdk:"app_name"`
	Description   types.String  `tfsdk:"description"`
	EnvironmentID types.String  `tfsdk:"environment_id"`
	SourceType    types.String  `tfsdk:"source_type"`
	Repository    types.String  `tfsdk:"repository"`
	Owner         types.String  `tfsdk:"owner"`
	Branch        types.String  `tfsdk:"branch"`
	BuildPath     types.String  `tfsdk:"build_path"`
	AutoDeploy    types.Bool    `tfsdk:"auto_deploy"`
	BuildType     types.String  `tfsdk:"build_type"`
	Dockerfile    types.String  `tfsdk:"dockerfile"`
	Replicas      types.Int64   `tfsdk:"replicas"`
	Status        types.String  `tfsdk:"status"`
	ServerID      types.String  `tfsdk:"server_id"`
	RegistryID    types.String  `tfsdk:"registry_id"`
	CreatedAt     types.String  `tfsdk:"created_at"`
	Domains       []DomainModel `tfsdk:"domains"`
	Ports         []PortModel   `tfsdk:"ports"`
	Mounts        []MountModel  `tfsdk:"mounts"`
}

// DomainModel describes a domain configuration.
type DomainModel struct {
	ID              types.String `tfsdk:"id"`
	Host            types.String `tfsdk:"host"`
	Port            types.Int64  `tfsdk:"port"`
	HTTPS           types.Bool   `tfsdk:"https"`
	CertificateType types.String `tfsdk:"certificate_type"`
	Path            types.String `tfsdk:"path"`
}

// PortModel describes a port mapping.
type PortModel struct {
	ID            types.String `tfsdk:"id"`
	PublishedPort types.Int64  `tfsdk:"published_port"`
	TargetPort    types.Int64  `tfsdk:"target_port"`
	Protocol      types.String `tfsdk:"protocol"`
}

// MountModel describes a volume mount.
type MountModel struct {
	ID         types.String `tfsdk:"id"`
	Type       types.String `tfsdk:"type"`
	HostPath   types.String `tfsdk:"host_path"`
	VolumeName types.String `tfsdk:"volume_name"`
	MountPath  types.String `tfsdk:"mount_path"`
}

func (d *ApplicationDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_application"
}

func (d *ApplicationDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches a single application from Dokploy by ID.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "The unique identifier of the application (same as application_id).",
				Computed:    true,
			},
			"application_id": schema.StringAttribute{
				Description: "The application ID to look up.",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "The name of the application.",
				Computed:    true,
			},
			"app_name": schema.StringAttribute{
				Description: "The internal app name (used for Docker naming).",
				Computed:    true,
			},
			"description": schema.StringAttribute{
				Description: "The description of the application.",
				Computed:    true,
			},
			"environment_id": schema.StringAttribute{
				Description: "The environment ID the application belongs to.",
				Computed:    true,
			},
			"source_type": schema.StringAttribute{
				Description: "The source type (github, docker, etc.).",
				Computed:    true,
			},
			"repository": schema.StringAttribute{
				Description: "The repository name.",
				Computed:    true,
			},
			"owner": schema.StringAttribute{
				Description: "The repository owner.",
				Computed:    true,
			},
			"branch": schema.StringAttribute{
				Description: "The branch to deploy from.",
				Computed:    true,
			},
			"build_path": schema.StringAttribute{
				Description: "The build path within the repository.",
				Computed:    true,
			},
			"auto_deploy": schema.BoolAttribute{
				Description: "Whether auto-deploy is enabled.",
				Computed:    true,
			},
			"build_type": schema.StringAttribute{
				Description: "The build type (nixpacks, dockerfile, etc.).",
				Computed:    true,
			},
			"dockerfile": schema.StringAttribute{
				Description: "The Dockerfile path.",
				Computed:    true,
			},
			"replicas": schema.Int64Attribute{
				Description: "The number of replicas.",
				Computed:    true,
			},
			"status": schema.StringAttribute{
				Description: "The current status of the application.",
				Computed:    true,
			},
			"server_id": schema.StringAttribute{
				Description: "The server ID where deployed.",
				Computed:    true,
			},
			"registry_id": schema.StringAttribute{
				Description: "The registry ID for the Docker image.",
				Computed:    true,
			},
			"created_at": schema.StringAttribute{
				Description: "The creation timestamp of the application.",
				Computed:    true,
			},
			"domains": schema.ListNestedAttribute{
				Description: "List of domains attached to the application.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the domain.",
							Computed:    true,
						},
						"host": schema.StringAttribute{
							Description: "The domain host.",
							Computed:    true,
						},
						"port": schema.Int64Attribute{
							Description: "The port to route traffic to.",
							Computed:    true,
						},
						"https": schema.BoolAttribute{
							Description: "Whether HTTPS is enabled.",
							Computed:    true,
						},
						"certificate_type": schema.StringAttribute{
							Description: "The certificate type.",
							Computed:    true,
						},
						"path": schema.StringAttribute{
							Description: "The path prefix.",
							Computed:    true,
						},
					},
				},
			},
			"ports": schema.ListNestedAttribute{
				Description: "List of port mappings.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the port mapping.",
							Computed:    true,
						},
						"published_port": schema.Int64Attribute{
							Description: "The published (external) port.",
							Computed:    true,
						},
						"target_port": schema.Int64Attribute{
							Description: "The target (internal) port.",
							Computed:    true,
						},
						"protocol": schema.StringAttribute{
							Description: "The protocol (tcp/udp).",
							Computed:    true,
						},
					},
				},
			},
			"mounts": schema.ListNestedAttribute{
				Description: "List of volume mounts.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "The unique identifier of the mount.",
							Computed:    true,
						},
						"type": schema.StringAttribute{
							Description: "The mount type.",
							Computed:    true,
						},
						"host_path": schema.StringAttribute{
							Description: "The host path.",
							Computed:    true,
						},
						"volume_name": schema.StringAttribute{
							Description: "The volume name.",
							Computed:    true,
						},
						"mount_path": schema.StringAttribute{
							Description: "The mount path inside the container.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ApplicationDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ApplicationDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config ApplicationDataSourceModel

	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	app, err := d.client.GetApplication(ctx, config.ApplicationID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read Dokploy Application",
			err.Error(),
		)
		return
	}

	serverID := ""
	if app.ServerID != nil {
		serverID = *app.ServerID
	}

	registryID := ""
	if app.RegistryID != nil {
		registryID = *app.RegistryID
	}

	state := ApplicationDataSourceModel{
		ID:            types.StringValue(app.ApplicationID),
		ApplicationID: types.StringValue(app.ApplicationID),
		Name:          types.StringValue(app.Name),
		AppName:       types.StringValue(app.AppName),
		Description:   types.StringValue(app.Description),
		EnvironmentID: types.StringValue(app.EnvironmentID),
		SourceType:    types.StringValue(app.SourceType),
		Repository:    types.StringValue(app.Repository),
		Owner:         types.StringValue(app.Owner),
		Branch:        types.StringValue(app.Branch),
		BuildPath:     types.StringValue(app.BuildPath),
		AutoDeploy:    types.BoolValue(app.AutoDeploy),
		BuildType:     types.StringValue(app.BuildType),
		Dockerfile:    types.StringValue(app.Dockerfile),
		Replicas:      types.Int64Value(int64(app.Replicas)),
		Status:        types.StringValue(app.ApplicationStatus),
		ServerID:      types.StringValue(serverID),
		RegistryID:    types.StringValue(registryID),
		CreatedAt:     types.StringValue(app.CreatedAt.Format("2006-01-02T15:04:05Z")),
		Domains:       []DomainModel{},
		Ports:         []PortModel{},
		Mounts:        []MountModel{},
	}

	for _, domain := range app.Domains {
		port := int64(0)
		if domain.Port != nil {
			port = int64(*domain.Port)
		}
		state.Domains = append(state.Domains, DomainModel{
			ID:              types.StringValue(domain.DomainID),
			Host:            types.StringValue(domain.Host),
			Port:            types.Int64Value(port),
			HTTPS:           types.BoolValue(domain.HTTPS),
			CertificateType: types.StringValue(domain.CertificateType),
			Path:            types.StringValue(domain.Path),
		})
	}

	for _, port := range app.Ports {
		state.Ports = append(state.Ports, PortModel{
			ID:            types.StringValue(port.PortID),
			PublishedPort: types.Int64Value(int64(port.PublishedPort)),
			TargetPort:    types.Int64Value(int64(port.TargetPort)),
			Protocol:      types.StringValue(port.Protocol),
		})
	}

	for _, mount := range app.Mounts {
		hostPath := ""
		if mount.HostPath != nil {
			hostPath = *mount.HostPath
		}
		volumeName := ""
		if mount.VolumeName != nil {
			volumeName = *mount.VolumeName
		}
		state.Mounts = append(state.Mounts, MountModel{
			ID:         types.StringValue(mount.MountID),
			Type:       types.StringValue(mount.Type),
			HostPath:   types.StringValue(hostPath),
			VolumeName: types.StringValue(volumeName),
			MountPath:  types.StringValue(mount.MountPath),
		})
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
