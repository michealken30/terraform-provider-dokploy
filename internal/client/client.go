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
