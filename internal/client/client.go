package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client is the Dokploy API client
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// New creates a new Dokploy API client
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// doRequest performs an authenticated GET request
func (c *Client) doRequest(ctx context.Context, endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/api%s", c.baseURL, endpoint)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// doPostRequest performs an authenticated POST request with JSON body
func (c *Client) doPostRequest(ctx context.Context, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/api%s", c.baseURL, endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	// Debug: log the request body
	fmt.Printf("[DEBUG] POST %s: %s\n", endpoint, string(jsonBody))

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(respBody))
	}

	return respBody, nil
}

// GetProjects fetches all projects with nested environments and services
func (c *Client) GetProjects(ctx context.Context) ([]Project, error) {
	data, err := c.doRequest(ctx, "/project.all")
	if err != nil {
		return nil, fmt.Errorf("fetching projects: %w", err)
	}

	var projects []Project
	if err := json.Unmarshal(data, &projects); err != nil {
		return nil, fmt.Errorf("parsing projects: %w", err)
	}

	return projects, nil
}

// GetProject fetches a single project by ID
func (c *Client) GetProject(ctx context.Context, projectID string) (*Project, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for i := range projects {
		if projects[i].ProjectID == projectID {
			return &projects[i], nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", projectID)
}

// GetProjectByName fetches a single project by name
func (c *Client) GetProjectByName(ctx context.Context, name string) (*Project, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for i := range projects {
		if projects[i].Name == name {
			return &projects[i], nil
		}
	}

	return nil, fmt.Errorf("project not found: %s", name)
}

// GetServers fetches all server configurations
func (c *Client) GetServers(ctx context.Context) ([]Server, error) {
	data, err := c.doRequest(ctx, "/server.all")
	if err != nil {
		return nil, fmt.Errorf("fetching servers: %w", err)
	}

	var servers []Server
	if err := json.Unmarshal(data, &servers); err != nil {
		return nil, fmt.Errorf("parsing servers: %w", err)
	}

	return servers, nil
}

// GetServer fetches a single server by ID
func (c *Client) GetServer(ctx context.Context, serverID string) (*Server, error) {
	servers, err := c.GetServers(ctx)
	if err != nil {
		return nil, err
	}

	for i := range servers {
		if servers[i].ServerID == serverID {
			return &servers[i], nil
		}
	}

	return nil, fmt.Errorf("server not found: %s", serverID)
}

// GetServerByName fetches a single server by name
func (c *Client) GetServerByName(ctx context.Context, name string) (*Server, error) {
	servers, err := c.GetServers(ctx)
	if err != nil {
		return nil, err
	}

	for i := range servers {
		if servers[i].Name == name {
			return &servers[i], nil
		}
	}

	return nil, fmt.Errorf("server not found: %s", name)
}

// GetSSHKeys fetches all SSH keys
func (c *Client) GetSSHKeys(ctx context.Context) ([]SSHKey, error) {
	data, err := c.doRequest(ctx, "/sshKey.all")
	if err != nil {
		return nil, fmt.Errorf("fetching SSH keys: %w", err)
	}

	var keys []SSHKey
	if err := json.Unmarshal(data, &keys); err != nil {
		return nil, fmt.Errorf("parsing SSH keys: %w", err)
	}

	return keys, nil
}

// GetSSHKey fetches a single SSH key by ID
func (c *Client) GetSSHKey(ctx context.Context, sshKeyID string) (*SSHKey, error) {
	keys, err := c.GetSSHKeys(ctx)
	if err != nil {
		return nil, err
	}

	for i := range keys {
		if keys[i].SSHKeyID == sshKeyID {
			return &keys[i], nil
		}
	}

	return nil, fmt.Errorf("SSH key not found: %s", sshKeyID)
}

// GetSSHKeyByName fetches a single SSH key by name
func (c *Client) GetSSHKeyByName(ctx context.Context, name string) (*SSHKey, error) {
	keys, err := c.GetSSHKeys(ctx)
	if err != nil {
		return nil, err
	}

	for i := range keys {
		if keys[i].Name == name {
			return &keys[i], nil
		}
	}

	return nil, fmt.Errorf("SSH key not found: %s", name)
}

// GetRegistries fetches all container registries
func (c *Client) GetRegistries(ctx context.Context) ([]Registry, error) {
	data, err := c.doRequest(ctx, "/registry.all")
	if err != nil {
		return nil, fmt.Errorf("fetching registries: %w", err)
	}

	var registries []Registry
	if err := json.Unmarshal(data, &registries); err != nil {
		return nil, fmt.Errorf("parsing registries: %w", err)
	}

	return registries, nil
}

// GetRegistry fetches a single registry by ID
func (c *Client) GetRegistry(ctx context.Context, registryID string) (*Registry, error) {
	registries, err := c.GetRegistries(ctx)
	if err != nil {
		return nil, err
	}
	for i := range registries {
		if registries[i].RegistryID == registryID {
			return &registries[i], nil
		}
	}
	return nil, fmt.Errorf("registry not found: %s", registryID)
}

// GetRegistryByName fetches a single registry by name
func (c *Client) GetRegistryByName(ctx context.Context, name string) (*Registry, error) {
	registries, err := c.GetRegistries(ctx)
	if err != nil {
		return nil, err
	}
	for i := range registries {
		if registries[i].RegistryName == name {
			return &registries[i], nil
		}
	}
	return nil, fmt.Errorf("registry not found: %s", name)
}

// CreateRegistryRequest represents the request body for creating a registry
type CreateRegistryRequest struct {
	RegistryName   string  `json:"registryName"`
	Username       string  `json:"username"`
	Password       string  `json:"password"`
	RegistryURL    string  `json:"registryUrl"`
	ImagePrefix    *string `json:"imagePrefix,omitempty"`
	RegistryType   string  `json:"registryType"` // "selfHosted", "docker", "github", etc.
	OrganizationID string  `json:"organizationId"`
}

// CreateRegistryResponse represents the response from creating a registry
type CreateRegistryResponse struct {
	RegistryID string `json:"registryId"`
}

// UpdateRegistryRequest represents the request body for updating a registry
type UpdateRegistryRequest struct {
	RegistryID   string  `json:"registryId"`
	RegistryName *string `json:"registryName,omitempty"`
	Username     *string `json:"username,omitempty"`
	Password     *string `json:"password,omitempty"`
	RegistryURL  *string `json:"registryUrl,omitempty"`
	ImagePrefix  *string `json:"imagePrefix,omitempty"`
}

// DeleteRegistryRequest represents the request body for deleting a registry
type DeleteRegistryRequest struct {
	RegistryID string `json:"registryId"`
}

// CreateRegistry creates a new container registry
func (c *Client) CreateRegistry(ctx context.Context, req CreateRegistryRequest) (*CreateRegistryResponse, error) {
	data, err := c.doPostRequest(ctx, "/registry.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating registry: %w", err)
	}
	var resp CreateRegistryResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create registry response: %w", err)
	}
	return &resp, nil
}

// UpdateRegistry updates an existing registry
func (c *Client) UpdateRegistry(ctx context.Context, req UpdateRegistryRequest) error {
	_, err := c.doPostRequest(ctx, "/registry.update", req)
	if err != nil {
		return fmt.Errorf("updating registry: %w", err)
	}
	return nil
}

// DeleteRegistry deletes a registry
func (c *Client) DeleteRegistry(ctx context.Context, registryID string) error {
	req := DeleteRegistryRequest{RegistryID: registryID}
	_, err := c.doPostRequest(ctx, "/registry.remove", req)
	if err != nil {
		return fmt.Errorf("deleting registry: %w", err)
	}
	return nil
}

// GetCertificates fetches all SSL certificates
func (c *Client) GetCertificates(ctx context.Context) ([]Certificate, error) {
	data, err := c.doRequest(ctx, "/certificates.all")
	if err != nil {
		return nil, fmt.Errorf("fetching certificates: %w", err)
	}

	var certificates []Certificate
	if err := json.Unmarshal(data, &certificates); err != nil {
		return nil, fmt.Errorf("parsing certificates: %w", err)
	}

	return certificates, nil
}

// GetCertificate fetches a single certificate by ID
func (c *Client) GetCertificate(ctx context.Context, certificateID string) (*Certificate, error) {
	certificates, err := c.GetCertificates(ctx)
	if err != nil {
		return nil, err
	}
	for i := range certificates {
		if certificates[i].CertificateID == certificateID {
			return &certificates[i], nil
		}
	}
	return nil, fmt.Errorf("certificate not found: %s", certificateID)
}

// GetCertificateByName fetches a single certificate by name
func (c *Client) GetCertificateByName(ctx context.Context, name string) (*Certificate, error) {
	certificates, err := c.GetCertificates(ctx)
	if err != nil {
		return nil, err
	}
	for i := range certificates {
		if certificates[i].Name == name {
			return &certificates[i], nil
		}
	}
	return nil, fmt.Errorf("certificate not found: %s", name)
}

// CreateCertificateRequest represents the request body for creating a certificate
type CreateCertificateRequest struct {
	Name            string `json:"name"`
	CertificateData string `json:"certificateData"`
	PrivateKey      string `json:"privateKey"`
	AutoRenew       *bool  `json:"autoRenew,omitempty"`
	OrganizationID  string `json:"organizationId"`
}

// CreateCertificateResponse represents the response from creating a certificate
type CreateCertificateResponse struct {
	CertificateID string `json:"certificateId"`
}

// UpdateCertificateRequest represents the request body for updating a certificate
type UpdateCertificateRequest struct {
	CertificateID   string  `json:"certificateId"`
	Name            *string `json:"name,omitempty"`
	CertificateData *string `json:"certificateData,omitempty"`
	PrivateKey      *string `json:"privateKey,omitempty"`
	AutoRenew       *bool   `json:"autoRenew,omitempty"`
}

// DeleteCertificateRequest represents the request body for deleting a certificate
type DeleteCertificateRequest struct {
	CertificateID string `json:"certificateId"`
}

// CreateCertificate creates a new SSL certificate
func (c *Client) CreateCertificate(ctx context.Context, req CreateCertificateRequest) (*CreateCertificateResponse, error) {
	data, err := c.doPostRequest(ctx, "/certificates.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating certificate: %w", err)
	}
	var resp CreateCertificateResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create certificate response: %w", err)
	}
	return &resp, nil
}

// UpdateCertificate updates an existing certificate
func (c *Client) UpdateCertificate(ctx context.Context, req UpdateCertificateRequest) error {
	_, err := c.doPostRequest(ctx, "/certificates.update", req)
	if err != nil {
		return fmt.Errorf("updating certificate: %w", err)
	}
	return nil
}

// DeleteCertificate deletes a certificate
func (c *Client) DeleteCertificate(ctx context.Context, certificateID string) error {
	req := DeleteCertificateRequest{CertificateID: certificateID}
	_, err := c.doPostRequest(ctx, "/certificates.remove", req)
	if err != nil {
		return fmt.Errorf("deleting certificate: %w", err)
	}
	return nil
}

// GetDestinations fetches all backup destinations
func (c *Client) GetDestinations(ctx context.Context) ([]Destination, error) {
	data, err := c.doRequest(ctx, "/destination.all")
	if err != nil {
		return nil, fmt.Errorf("fetching destinations: %w", err)
	}

	var destinations []Destination
	if err := json.Unmarshal(data, &destinations); err != nil {
		return nil, fmt.Errorf("parsing destinations: %w", err)
	}

	return destinations, nil
}

// GetDestination fetches a single destination by ID
func (c *Client) GetDestination(ctx context.Context, destinationID string) (*Destination, error) {
	destinations, err := c.GetDestinations(ctx)
	if err != nil {
		return nil, err
	}
	for i := range destinations {
		if destinations[i].DestinationID == destinationID {
			return &destinations[i], nil
		}
	}
	return nil, fmt.Errorf("destination not found: %s", destinationID)
}

// GetDestinationByName fetches a single destination by name
func (c *Client) GetDestinationByName(ctx context.Context, name string) (*Destination, error) {
	destinations, err := c.GetDestinations(ctx)
	if err != nil {
		return nil, err
	}
	for i := range destinations {
		if destinations[i].Name == name {
			return &destinations[i], nil
		}
	}
	return nil, fmt.Errorf("destination not found: %s", name)
}

// CreateDestinationRequest represents the request body for creating a destination
type CreateDestinationRequest struct {
	Name            string `json:"name"`
	AccessKey       string `json:"accessKey"`
	SecretAccessKey string `json:"secretAccessKey"`
	Bucket          string `json:"bucket"`
	Region          string `json:"region"`
	Endpoint        string `json:"endpoint"`
	OrganizationID  string `json:"organizationId"`
}

// CreateDestinationResponse represents the response from creating a destination
type CreateDestinationResponse struct {
	DestinationID string `json:"destinationId"`
}

// UpdateDestinationRequest represents the request body for updating a destination
type UpdateDestinationRequest struct {
	DestinationID   string  `json:"destinationId"`
	Name            *string `json:"name,omitempty"`
	AccessKey       *string `json:"accessKey,omitempty"`
	SecretAccessKey *string `json:"secretAccessKey,omitempty"`
	Bucket          *string `json:"bucket,omitempty"`
	Region          *string `json:"region,omitempty"`
	Endpoint        *string `json:"endpoint,omitempty"`
}

// DeleteDestinationRequest represents the request body for deleting a destination
type DeleteDestinationRequest struct {
	DestinationID string `json:"destinationId"`
}

// CreateDestination creates a new backup destination
func (c *Client) CreateDestination(ctx context.Context, req CreateDestinationRequest) (*CreateDestinationResponse, error) {
	data, err := c.doPostRequest(ctx, "/destination.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating destination: %w", err)
	}
	var resp CreateDestinationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create destination response: %w", err)
	}
	return &resp, nil
}

// UpdateDestination updates an existing destination
func (c *Client) UpdateDestination(ctx context.Context, req UpdateDestinationRequest) error {
	_, err := c.doPostRequest(ctx, "/destination.update", req)
	if err != nil {
		return fmt.Errorf("updating destination: %w", err)
	}
	return nil
}

// DeleteDestination deletes a destination
func (c *Client) DeleteDestination(ctx context.Context, destinationID string) error {
	req := DeleteDestinationRequest{DestinationID: destinationID}
	_, err := c.doPostRequest(ctx, "/destination.remove", req)
	if err != nil {
		return fmt.Errorf("deleting destination: %w", err)
	}
	return nil
}

// GetNotifications fetches all notification configurations
func (c *Client) GetNotifications(ctx context.Context) ([]Notification, error) {
	data, err := c.doRequest(ctx, "/notification.all")
	if err != nil {
		return nil, fmt.Errorf("fetching notifications: %w", err)
	}

	var notifications []Notification
	if err := json.Unmarshal(data, &notifications); err != nil {
		return nil, fmt.Errorf("parsing notifications: %w", err)
	}

	return notifications, nil
}

// GetApplication finds an application by ID across all projects
func (c *Client) GetApplication(ctx context.Context, applicationID string) (*Application, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for _, env := range proj.Environments {
			for i := range env.Applications {
				if env.Applications[i].ApplicationID == applicationID {
					return &env.Applications[i], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("application not found: %s", applicationID)
}

// GetEnvironment finds an environment by ID across all projects
func (c *Client) GetEnvironment(ctx context.Context, environmentID string) (*Environment, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for i := range proj.Environments {
			if proj.Environments[i].EnvironmentID == environmentID {
				return &proj.Environments[i], nil
			}
		}
	}

	return nil, fmt.Errorf("environment not found: %s", environmentID)
}

// GetEnvironmentsByProjectID fetches all environments for a project
func (c *Client) GetEnvironmentsByProjectID(ctx context.Context, projectID string) ([]Environment, error) {
	project, err := c.GetProject(ctx, projectID)
	if err != nil {
		return nil, err
	}

	return project.Environments, nil
}

// =============================================================================
// Project CRUD Operations
// =============================================================================

// CreateProjectRequest represents the request body for creating a project
type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
}

// CreateProjectResponse represents the response from creating a project
type CreateProjectResponse struct {
	Project struct {
		ProjectID string `json:"projectId"`
	} `json:"project"`
}

// UpdateProjectRequest represents the request body for updating a project
type UpdateProjectRequest struct {
	ProjectID   string  `json:"projectId"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// DeleteProjectRequest represents the request body for deleting a project
type DeleteProjectRequest struct {
	ProjectID string `json:"projectId"`
}

// CreateProject creates a new project
func (c *Client) CreateProject(ctx context.Context, req CreateProjectRequest) (*CreateProjectResponse, error) {
	data, err := c.doPostRequest(ctx, "/project.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating project: %w", err)
	}

	var resp CreateProjectResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create project response: %w", err)
	}

	return &resp, nil
}

// UpdateProject updates an existing project
func (c *Client) UpdateProject(ctx context.Context, req UpdateProjectRequest) error {
	_, err := c.doPostRequest(ctx, "/project.update", req)
	if err != nil {
		return fmt.Errorf("updating project: %w", err)
	}
	return nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(ctx context.Context, projectID string) error {
	req := DeleteProjectRequest{ProjectID: projectID}
	_, err := c.doPostRequest(ctx, "/project.remove", req)
	if err != nil {
		return fmt.Errorf("deleting project: %w", err)
	}
	return nil
}

// =============================================================================
// SSH Key CRUD Operations
// =============================================================================

// CreateSSHKeyRequest represents the request body for creating an SSH key
type CreateSSHKeyRequest struct {
	Name           string  `json:"name"`
	Description    *string `json:"description,omitempty"`
	PrivateKey     string  `json:"privateKey"`
	PublicKey      string  `json:"publicKey"`
	OrganizationID string  `json:"organizationId"`
}

// CreateSSHKeyResponse represents the response from creating an SSH key
type CreateSSHKeyResponse struct {
	SSHKeyID string `json:"sshKeyId"`
}

// UpdateSSHKeyRequest represents the request body for updating an SSH key
type UpdateSSHKeyRequest struct {
	SSHKeyID    string  `json:"sshKeyId"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

// DeleteSSHKeyRequest represents the request body for deleting an SSH key
type DeleteSSHKeyRequest struct {
	SSHKeyID string `json:"sshKeyId"`
}

// CreateSSHKey creates a new SSH key
func (c *Client) CreateSSHKey(ctx context.Context, req CreateSSHKeyRequest) (*CreateSSHKeyResponse, error) {
	data, err := c.doPostRequest(ctx, "/sshKey.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating SSH key: %w", err)
	}

	var resp CreateSSHKeyResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create SSH key response: %w", err)
	}

	return &resp, nil
}

// UpdateSSHKey updates an existing SSH key
func (c *Client) UpdateSSHKey(ctx context.Context, req UpdateSSHKeyRequest) error {
	_, err := c.doPostRequest(ctx, "/sshKey.update", req)
	if err != nil {
		return fmt.Errorf("updating SSH key: %w", err)
	}
	return nil
}

// DeleteSSHKey deletes an SSH key
func (c *Client) DeleteSSHKey(ctx context.Context, sshKeyID string) error {
	req := DeleteSSHKeyRequest{SSHKeyID: sshKeyID}
	_, err := c.doPostRequest(ctx, "/sshKey.remove", req)
	if err != nil {
		return fmt.Errorf("deleting SSH key: %w", err)
	}
	return nil
}

// =============================================================================
// Server CRUD Operations
// =============================================================================

// CreateServerRequest represents the request body for creating a server
type CreateServerRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	IPAddress   string  `json:"ipAddress"`
	Port        int     `json:"port"`
	Username    string  `json:"username"`
	SSHKeyID    *string `json:"sshKeyId"`
	ServerType  string  `json:"serverType"` // "deploy" or "build"
}

// CreateServerResponse represents the response from creating a server
type CreateServerResponse struct {
	ServerID string `json:"serverId"`
}

// UpdateServerRequest represents the request body for updating a server
type UpdateServerRequest struct {
	ServerID    string  `json:"serverId"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	IPAddress   *string `json:"ipAddress,omitempty"`
	Port        *int    `json:"port,omitempty"`
	Username    *string `json:"username,omitempty"`
	SSHKeyID    *string `json:"sshKeyId,omitempty"`
}

// DeleteServerRequest represents the request body for deleting a server
type DeleteServerRequest struct {
	ServerID string `json:"serverId"`
}

// CreateServer creates a new server
func (c *Client) CreateServer(ctx context.Context, req CreateServerRequest) (*CreateServerResponse, error) {
	data, err := c.doPostRequest(ctx, "/server.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating server: %w", err)
	}

	var resp CreateServerResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create server response: %w", err)
	}

	return &resp, nil
}

// UpdateServer updates an existing server
func (c *Client) UpdateServer(ctx context.Context, req UpdateServerRequest) error {
	_, err := c.doPostRequest(ctx, "/server.update", req)
	if err != nil {
		return fmt.Errorf("updating server: %w", err)
	}
	return nil
}

// DeleteServer deletes a server
func (c *Client) DeleteServer(ctx context.Context, serverID string) error {
	req := DeleteServerRequest{ServerID: serverID}
	_, err := c.doPostRequest(ctx, "/server.remove", req)
	if err != nil {
		return fmt.Errorf("deleting server: %w", err)
	}
	return nil
}

// =============================================================================
// Environment CRUD Operations
// =============================================================================

// CreateEnvironmentRequest represents the request body for creating an environment
type CreateEnvironmentRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	ProjectID   string  `json:"projectId"`
}

// CreateEnvironmentResponse represents the response from creating an environment
type CreateEnvironmentResponse struct {
	EnvironmentID string `json:"environmentId"`
}

// UpdateEnvironmentRequest represents the request body for updating an environment
type UpdateEnvironmentRequest struct {
	EnvironmentID string  `json:"environmentId"`
	Name          *string `json:"name,omitempty"`
	Description   *string `json:"description,omitempty"`
}

// DeleteEnvironmentRequest represents the request body for deleting an environment
type DeleteEnvironmentRequest struct {
	EnvironmentID string `json:"environmentId"`
}

// CreateEnvironment creates a new environment
func (c *Client) CreateEnvironment(ctx context.Context, req CreateEnvironmentRequest) (*CreateEnvironmentResponse, error) {
	data, err := c.doPostRequest(ctx, "/environment.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating environment: %w", err)
	}

	var resp CreateEnvironmentResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create environment response: %w", err)
	}

	return &resp, nil
}

// UpdateEnvironment updates an existing environment
func (c *Client) UpdateEnvironment(ctx context.Context, req UpdateEnvironmentRequest) error {
	_, err := c.doPostRequest(ctx, "/environment.update", req)
	if err != nil {
		return fmt.Errorf("updating environment: %w", err)
	}
	return nil
}

// DeleteEnvironment deletes an environment
func (c *Client) DeleteEnvironment(ctx context.Context, environmentID string) error {
	req := DeleteEnvironmentRequest{EnvironmentID: environmentID}
	_, err := c.doPostRequest(ctx, "/environment.remove", req)
	if err != nil {
		return fmt.Errorf("deleting environment: %w", err)
	}
	return nil
}

// =============================================================================
// Application CRUD Operations
// =============================================================================

// CreateApplicationRequest represents the request body for creating an application
type CreateApplicationRequest struct {
	Name          string  `json:"name"`
	AppName       *string `json:"appName,omitempty"`
	Description   *string `json:"description,omitempty"`
	EnvironmentID string  `json:"environmentId"`
	ServerID      *string `json:"serverId,omitempty"`
}

// CreateApplicationResponse represents the response from creating an application
type CreateApplicationResponse struct {
	ApplicationID string `json:"applicationId"`
}

// UpdateApplicationRequest represents the request body for updating an application
type UpdateApplicationRequest struct {
	ApplicationID string  `json:"applicationId"`
	Name          *string `json:"name,omitempty"`
	AppName       *string `json:"appName,omitempty"`
	Description   *string `json:"description,omitempty"`
}

// DeleteApplicationRequest represents the request body for deleting an application
type DeleteApplicationRequest struct {
	ApplicationID string `json:"applicationId"`
}

// CreateApplication creates a new application
func (c *Client) CreateApplication(ctx context.Context, req CreateApplicationRequest) (*CreateApplicationResponse, error) {
	data, err := c.doPostRequest(ctx, "/application.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating application: %w", err)
	}

	var resp CreateApplicationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create application response: %w", err)
	}

	return &resp, nil
}

// UpdateApplication updates an existing application
func (c *Client) UpdateApplication(ctx context.Context, req UpdateApplicationRequest) error {
	_, err := c.doPostRequest(ctx, "/application.update", req)
	if err != nil {
		return fmt.Errorf("updating application: %w", err)
	}
	return nil
}

// DeleteApplication deletes an application
func (c *Client) DeleteApplication(ctx context.Context, applicationID string) error {
	req := DeleteApplicationRequest{ApplicationID: applicationID}
	_, err := c.doPostRequest(ctx, "/application.delete", req)
	if err != nil {
		return fmt.Errorf("deleting application: %w", err)
	}
	return nil
}

// =============================================================================
// Compose CRUD Operations
// =============================================================================

// CreateComposeRequest represents the request body for creating a compose service
type CreateComposeRequest struct {
	Name          string  `json:"name"`
	Description   *string `json:"description,omitempty"`
	EnvironmentID string  `json:"environmentId"`
	ServerID      *string `json:"serverId,omitempty"`
	ComposeType   string  `json:"composeType"`
}

// CreateComposeResponse represents the response from creating a compose service
type CreateComposeResponse struct {
	ComposeID string `json:"composeId"`
}

// UpdateComposeRequest represents the request body for updating a compose service
type UpdateComposeRequest struct {
	ComposeID   string  `json:"composeId"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
	ComposeFile *string `json:"composeFile,omitempty"`
}

// DeleteComposeRequest represents the request body for deleting a compose service
type DeleteComposeRequest struct {
	ComposeID string `json:"composeId"`
}

// CreateCompose creates a new compose service
func (c *Client) CreateCompose(ctx context.Context, req CreateComposeRequest) (*CreateComposeResponse, error) {
	data, err := c.doPostRequest(ctx, "/compose.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating compose: %w", err)
	}

	var resp CreateComposeResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create compose response: %w", err)
	}

	return &resp, nil
}

// GetCompose finds a compose by ID across all projects
func (c *Client) GetCompose(ctx context.Context, composeID string) (*Compose, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for _, env := range proj.Environments {
			for i := range env.Compose {
				if env.Compose[i].ComposeID == composeID {
					return &env.Compose[i], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("compose not found: %s", composeID)
}

// UpdateCompose updates an existing compose service
func (c *Client) UpdateCompose(ctx context.Context, req UpdateComposeRequest) error {
	_, err := c.doPostRequest(ctx, "/compose.update", req)
	if err != nil {
		return fmt.Errorf("updating compose: %w", err)
	}
	return nil
}

// DeleteCompose deletes a compose service
func (c *Client) DeleteCompose(ctx context.Context, composeID string) error {
	req := DeleteComposeRequest{ComposeID: composeID}
	_, err := c.doPostRequest(ctx, "/compose.delete", req)
	if err != nil {
		return fmt.Errorf("deleting compose: %w", err)
	}
	return nil
}

// =============================================================================
// Postgres CRUD Operations
// =============================================================================

// CreatePostgresRequest represents the request body for creating a postgres service
type CreatePostgresRequest struct {
	Name             string  `json:"name"`
	AppName          string  `json:"appName"`
	Description      *string `json:"description,omitempty"`
	EnvironmentID    string  `json:"environmentId"`
	ServerID         *string `json:"serverId,omitempty"`
	DatabaseName     string  `json:"databaseName"`
	DatabaseUser     string  `json:"databaseUser"`
	DatabasePassword string  `json:"databasePassword"`
	DockerImage      *string `json:"dockerImage,omitempty"`
}

// CreatePostgresResponse represents the response from creating a postgres service
type CreatePostgresResponse struct {
	PostgresID string `json:"postgresId"`
}

// UpdatePostgresRequest represents the request body for updating a postgres service
type UpdatePostgresRequest struct {
	PostgresID       string  `json:"postgresId"`
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	DatabaseName     *string `json:"databaseName,omitempty"`
	DatabaseUser     *string `json:"databaseUser,omitempty"`
	DatabasePassword *string `json:"databasePassword,omitempty"`
	DockerImage      *string `json:"dockerImage,omitempty"`
}

// DeletePostgresRequest represents the request body for deleting a postgres service
type DeletePostgresRequest struct {
	PostgresID string `json:"postgresId"`
}

// CreatePostgres creates a new postgres service
func (c *Client) CreatePostgres(ctx context.Context, req CreatePostgresRequest) (*CreatePostgresResponse, error) {
	data, err := c.doPostRequest(ctx, "/postgres.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating postgres: %w", err)
	}

	var resp CreatePostgresResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create postgres response: %w", err)
	}

	return &resp, nil
}

// GetPostgres finds a postgres by ID across all projects
func (c *Client) GetPostgres(ctx context.Context, postgresID string) (*Postgres, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for _, env := range proj.Environments {
			for i := range env.Postgres {
				if env.Postgres[i].PostgresID == postgresID {
					return &env.Postgres[i], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("postgres not found: %s", postgresID)
}

// UpdatePostgres updates an existing postgres service
func (c *Client) UpdatePostgres(ctx context.Context, req UpdatePostgresRequest) error {
	_, err := c.doPostRequest(ctx, "/postgres.update", req)
	if err != nil {
		return fmt.Errorf("updating postgres: %w", err)
	}
	return nil
}

// DeletePostgres deletes a postgres service
func (c *Client) DeletePostgres(ctx context.Context, postgresID string) error {
	req := DeletePostgresRequest{PostgresID: postgresID}
	_, err := c.doPostRequest(ctx, "/postgres.delete", req)
	if err != nil {
		return fmt.Errorf("deleting postgres: %w", err)
	}
	return nil
}

// =============================================================================
// MySQL CRUD Operations
// =============================================================================

// CreateMysqlRequest represents the request body for creating a mysql service
type CreateMysqlRequest struct {
	Name                 string  `json:"name"`
	AppName              string  `json:"appName"`
	Description          *string `json:"description,omitempty"`
	EnvironmentID        string  `json:"environmentId"`
	ServerID             *string `json:"serverId,omitempty"`
	DatabaseName         string  `json:"databaseName"`
	DatabaseUser         string  `json:"databaseUser"`
	DatabasePassword     string  `json:"databasePassword"`
	DatabaseRootPassword string  `json:"databaseRootPassword"`
	DockerImage          *string `json:"dockerImage,omitempty"`
}

// CreateMysqlResponse represents the response from creating a mysql service
type CreateMysqlResponse struct {
	MysqlID string `json:"mysqlId"`
}

// UpdateMysqlRequest represents the request body for updating a mysql service
type UpdateMysqlRequest struct {
	MysqlID              string  `json:"mysqlId"`
	Name                 *string `json:"name,omitempty"`
	Description          *string `json:"description,omitempty"`
	DatabaseName         *string `json:"databaseName,omitempty"`
	DatabaseUser         *string `json:"databaseUser,omitempty"`
	DatabasePassword     *string `json:"databasePassword,omitempty"`
	DatabaseRootPassword *string `json:"databaseRootPassword,omitempty"`
	DockerImage          *string `json:"dockerImage,omitempty"`
}

// DeleteMysqlRequest represents the request body for deleting a mysql service
type DeleteMysqlRequest struct {
	MysqlID string `json:"mysqlId"`
}

// CreateMysql creates a new mysql service
func (c *Client) CreateMysql(ctx context.Context, req CreateMysqlRequest) (*CreateMysqlResponse, error) {
	data, err := c.doPostRequest(ctx, "/mysql.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating mysql: %w", err)
	}

	var resp CreateMysqlResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create mysql response: %w", err)
	}

	return &resp, nil
}

// GetMysql finds a mysql by ID across all projects
func (c *Client) GetMysql(ctx context.Context, mysqlID string) (*MySQL, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for _, env := range proj.Environments {
			for i := range env.MySQL {
				if env.MySQL[i].MySQLID == mysqlID {
					return &env.MySQL[i], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("mysql not found: %s", mysqlID)
}

// UpdateMysql updates an existing mysql service
func (c *Client) UpdateMysql(ctx context.Context, req UpdateMysqlRequest) error {
	_, err := c.doPostRequest(ctx, "/mysql.update", req)
	if err != nil {
		return fmt.Errorf("updating mysql: %w", err)
	}
	return nil
}

// DeleteMysql deletes a mysql service
func (c *Client) DeleteMysql(ctx context.Context, mysqlID string) error {
	req := DeleteMysqlRequest{MysqlID: mysqlID}
	_, err := c.doPostRequest(ctx, "/mysql.delete", req)
	if err != nil {
		return fmt.Errorf("deleting mysql: %w", err)
	}
	return nil
}

// =============================================================================
// MariaDB CRUD Operations
// =============================================================================

// CreateMariadbRequest represents the request body for creating a mariadb service
type CreateMariadbRequest struct {
	Name                 string  `json:"name"`
	AppName              string  `json:"appName"`
	Description          *string `json:"description,omitempty"`
	EnvironmentID        string  `json:"environmentId"`
	ServerID             *string `json:"serverId,omitempty"`
	DatabaseName         string  `json:"databaseName"`
	DatabaseUser         string  `json:"databaseUser"`
	DatabasePassword     string  `json:"databasePassword"`
	DatabaseRootPassword string  `json:"databaseRootPassword"`
	DockerImage          *string `json:"dockerImage,omitempty"`
}

// CreateMariadbResponse represents the response from creating a mariadb service
type CreateMariadbResponse struct {
	MariadbID string `json:"mariadbId"`
}

// UpdateMariadbRequest represents the request body for updating a mariadb service
type UpdateMariadbRequest struct {
	MariadbID            string  `json:"mariadbId"`
	Name                 *string `json:"name,omitempty"`
	Description          *string `json:"description,omitempty"`
	DatabaseName         *string `json:"databaseName,omitempty"`
	DatabaseUser         *string `json:"databaseUser,omitempty"`
	DatabasePassword     *string `json:"databasePassword,omitempty"`
	DatabaseRootPassword *string `json:"databaseRootPassword,omitempty"`
	DockerImage          *string `json:"dockerImage,omitempty"`
}

// DeleteMariadbRequest represents the request body for deleting a mariadb service
type DeleteMariadbRequest struct {
	MariadbID string `json:"mariadbId"`
}

// CreateMariadb creates a new mariadb service
func (c *Client) CreateMariadb(ctx context.Context, req CreateMariadbRequest) (*CreateMariadbResponse, error) {
	data, err := c.doPostRequest(ctx, "/mariadb.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating mariadb: %w", err)
	}

	var resp CreateMariadbResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create mariadb response: %w", err)
	}

	return &resp, nil
}

// GetMariadb finds a mariadb by ID across all projects
func (c *Client) GetMariadb(ctx context.Context, mariadbID string) (*MariaDB, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for _, env := range proj.Environments {
			for i := range env.MariaDB {
				if env.MariaDB[i].MariaDBID == mariadbID {
					return &env.MariaDB[i], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("mariadb not found: %s", mariadbID)
}

// UpdateMariadb updates an existing mariadb service
func (c *Client) UpdateMariadb(ctx context.Context, req UpdateMariadbRequest) error {
	_, err := c.doPostRequest(ctx, "/mariadb.update", req)
	if err != nil {
		return fmt.Errorf("updating mariadb: %w", err)
	}
	return nil
}

// DeleteMariadb deletes a mariadb service
func (c *Client) DeleteMariadb(ctx context.Context, mariadbID string) error {
	req := DeleteMariadbRequest{MariadbID: mariadbID}
	_, err := c.doPostRequest(ctx, "/mariadb.delete", req)
	if err != nil {
		return fmt.Errorf("deleting mariadb: %w", err)
	}
	return nil
}

// =============================================================================
// MongoDB CRUD Operations
// =============================================================================

// CreateMongoRequest represents the request body for creating a mongo service
type CreateMongoRequest struct {
	Name             string  `json:"name"`
	AppName          string  `json:"appName"`
	Description      *string `json:"description,omitempty"`
	EnvironmentID    string  `json:"environmentId"`
	ServerID         *string `json:"serverId,omitempty"`
	DatabaseUser     string  `json:"databaseUser"`
	DatabasePassword string  `json:"databasePassword"`
	DockerImage      *string `json:"dockerImage,omitempty"`
}

// CreateMongoResponse represents the response from creating a mongo service
type CreateMongoResponse struct {
	MongoID string `json:"mongoId"`
}

// UpdateMongoRequest represents the request body for updating a mongo service
type UpdateMongoRequest struct {
	MongoID          string  `json:"mongoId"`
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	DatabaseUser     *string `json:"databaseUser,omitempty"`
	DatabasePassword *string `json:"databasePassword,omitempty"`
	DockerImage      *string `json:"dockerImage,omitempty"`
}

// DeleteMongoRequest represents the request body for deleting a mongo service
type DeleteMongoRequest struct {
	MongoID string `json:"mongoId"`
}

// CreateMongo creates a new mongo service
func (c *Client) CreateMongo(ctx context.Context, req CreateMongoRequest) (*CreateMongoResponse, error) {
	data, err := c.doPostRequest(ctx, "/mongo.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating mongo: %w", err)
	}

	var resp CreateMongoResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create mongo response: %w", err)
	}

	return &resp, nil
}

// GetMongo finds a mongo by ID across all projects
func (c *Client) GetMongo(ctx context.Context, mongoID string) (*Mongo, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for _, env := range proj.Environments {
			for i := range env.Mongo {
				if env.Mongo[i].MongoID == mongoID {
					return &env.Mongo[i], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("mongo not found: %s", mongoID)
}

// UpdateMongo updates an existing mongo service
func (c *Client) UpdateMongo(ctx context.Context, req UpdateMongoRequest) error {
	_, err := c.doPostRequest(ctx, "/mongo.update", req)
	if err != nil {
		return fmt.Errorf("updating mongo: %w", err)
	}
	return nil
}

// DeleteMongo deletes a mongo service
func (c *Client) DeleteMongo(ctx context.Context, mongoID string) error {
	req := DeleteMongoRequest{MongoID: mongoID}
	_, err := c.doPostRequest(ctx, "/mongo.delete", req)
	if err != nil {
		return fmt.Errorf("deleting mongo: %w", err)
	}
	return nil
}

// =============================================================================
// Redis CRUD Operations
// =============================================================================

// CreateRedisRequest represents the request body for creating a redis service
type CreateRedisRequest struct {
	Name             string  `json:"name"`
	AppName          string  `json:"appName"`
	Description      *string `json:"description,omitempty"`
	EnvironmentID    string  `json:"environmentId"`
	ServerID         *string `json:"serverId,omitempty"`
	DatabasePassword string  `json:"databasePassword"`
	DockerImage      *string `json:"dockerImage,omitempty"`
}

// CreateRedisResponse represents the response from creating a redis service
type CreateRedisResponse struct {
	RedisID string `json:"redisId"`
}

// UpdateRedisRequest represents the request body for updating a redis service
type UpdateRedisRequest struct {
	RedisID          string  `json:"redisId"`
	Name             *string `json:"name,omitempty"`
	Description      *string `json:"description,omitempty"`
	DatabasePassword *string `json:"databasePassword,omitempty"`
	DockerImage      *string `json:"dockerImage,omitempty"`
}

// DeleteRedisRequest represents the request body for deleting a redis service
type DeleteRedisRequest struct {
	RedisID string `json:"redisId"`
}

// CreateRedis creates a new redis service
func (c *Client) CreateRedis(ctx context.Context, req CreateRedisRequest) (*CreateRedisResponse, error) {
	data, err := c.doPostRequest(ctx, "/redis.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating redis: %w", err)
	}

	var resp CreateRedisResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create redis response: %w", err)
	}

	return &resp, nil
}

// GetRedis finds a redis by ID across all projects
func (c *Client) GetRedis(ctx context.Context, redisID string) (*Redis, error) {
	projects, err := c.GetProjects(ctx)
	if err != nil {
		return nil, err
	}

	for _, proj := range projects {
		for _, env := range proj.Environments {
			for i := range env.Redis {
				if env.Redis[i].RedisID == redisID {
					return &env.Redis[i], nil
				}
			}
		}
	}

	return nil, fmt.Errorf("redis not found: %s", redisID)
}

// UpdateRedis updates an existing redis service
func (c *Client) UpdateRedis(ctx context.Context, req UpdateRedisRequest) error {
	_, err := c.doPostRequest(ctx, "/redis.update", req)
	if err != nil {
		return fmt.Errorf("updating redis: %w", err)
	}
	return nil
}

// DeleteRedis deletes a redis service
func (c *Client) DeleteRedis(ctx context.Context, redisID string) error {
	req := DeleteRedisRequest{RedisID: redisID}
	_, err := c.doPostRequest(ctx, "/redis.delete", req)
	if err != nil {
		return fmt.Errorf("deleting redis: %w", err)
	}
	return nil
}
