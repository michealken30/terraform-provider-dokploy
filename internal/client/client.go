package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"time"
)

// RetryConfig configures retry behavior for API requests.
type RetryConfig struct {
	MaxRetries     int
	InitialBackoff time.Duration
	MaxBackoff     time.Duration
}

// DefaultRetryConfig returns the default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:     3,
		InitialBackoff: 500 * time.Millisecond,
		MaxBackoff:     30 * time.Second,
	}
}

// Client is the Dokploy API client
type Client struct {
	baseURL     string
	apiKey      string
	httpClient  *http.Client
	retryConfig RetryConfig
}

// New creates a new Dokploy API client
func New(baseURL, apiKey string) *Client {
	return &Client{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		retryConfig: DefaultRetryConfig(),
	}
}

// NewWithRetry creates a new Dokploy API client with custom retry configuration.
func NewWithRetry(baseURL, apiKey string, retryConfig RetryConfig) *Client {
	return &Client{
		baseURL:     baseURL,
		apiKey:      apiKey,
		httpClient:  &http.Client{Timeout: 30 * time.Second},
		retryConfig: retryConfig,
	}
}

// calculateBackoff returns the backoff duration for the given attempt.
func (c *Client) calculateBackoff(attempt int) time.Duration {
	backoff := float64(c.retryConfig.InitialBackoff) * math.Pow(2, float64(attempt))
	if backoff > float64(c.retryConfig.MaxBackoff) {
		backoff = float64(c.retryConfig.MaxBackoff)
	}
	return time.Duration(backoff)
}

// shouldRetry returns true if the request should be retried based on the status code.
func shouldRetry(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests ||
		(statusCode >= 500 && statusCode < 600)
}

// doRequest performs an authenticated GET request with retry support.
func (c *Client) doRequest(ctx context.Context, endpoint string) ([]byte, error) {
	url := fmt.Sprintf("%s/api%s", c.baseURL, endpoint)

	var lastErr error
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.calculateBackoff(attempt - 1)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return body, nil
		}

		lastErr = newAPIError(resp.StatusCode, endpoint, body)
		if !shouldRetry(resp.StatusCode) {
			return nil, lastErr
		}
	}

	return nil, lastErr
}

// doPostRequest performs an authenticated POST request with JSON body and retry support.
func (c *Client) doPostRequest(ctx context.Context, endpoint string, body interface{}) ([]byte, error) {
	url := fmt.Sprintf("%s/api%s", c.baseURL, endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshaling request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.calculateBackoff(attempt - 1)
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff):
			}
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
			lastErr = fmt.Errorf("executing request: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			continue
		}

		if resp.StatusCode == http.StatusOK {
			return respBody, nil
		}

		lastErr = newAPIError(resp.StatusCode, endpoint, respBody)
		if !shouldRetry(resp.StatusCode) {
			return nil, lastErr
		}
	}

	return nil, lastErr
}

// doDeleteRequest performs a delete request with retry support, treating 404 as success.
func (c *Client) doDeleteRequest(ctx context.Context, endpoint string, body interface{}) error {
	url := fmt.Sprintf("%s/api%s", c.baseURL, endpoint)

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return fmt.Errorf("marshaling request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		if attempt > 0 {
			backoff := c.calculateBackoff(attempt - 1)
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
		if err != nil {
			return fmt.Errorf("creating request: %w", err)
		}

		req.Header.Set("x-api-key", c.apiKey)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Accept", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			lastErr = fmt.Errorf("executing request: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("reading response body: %w", err)
			continue
		}

		// Treat 200 OK and 404 Not Found as success (resource deleted or already gone)
		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusNotFound {
			return nil
		}

		lastErr = newAPIError(resp.StatusCode, endpoint, respBody)
		if !shouldRetry(resp.StatusCode) {
			return lastErr
		}
	}

	return lastErr
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

	return nil, newNotFoundError("project", projectID)
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

	return nil, newNotFoundError("project", name)
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

	return nil, newNotFoundError("server", serverID)
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

	return nil, newNotFoundError("server", name)
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

	return nil, newNotFoundError("SSH key", sshKeyID)
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

	return nil, newNotFoundError("SSH key", name)
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
	return nil, newNotFoundError("registry", registryID)
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
	return nil, newNotFoundError("registry", name)
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
	return c.doDeleteRequest(ctx, "/registry.remove", req)
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
	return nil, newNotFoundError("certificate", certificateID)
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
	return nil, newNotFoundError("certificate", name)
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
	return c.doDeleteRequest(ctx, "/certificates.remove", req)
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
	return nil, newNotFoundError("destination", destinationID)
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
	return nil, newNotFoundError("destination", name)
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
	return c.doDeleteRequest(ctx, "/destination.remove", req)
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

	return nil, newNotFoundError("application", applicationID)
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

	return nil, newNotFoundError("environment", environmentID)
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
	return c.doDeleteRequest(ctx, "/project.remove", req)
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
	return c.doDeleteRequest(ctx, "/sshKey.remove", req)
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
	return c.doDeleteRequest(ctx, "/server.remove", req)
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
	return c.doDeleteRequest(ctx, "/environment.remove", req)
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
	return c.doDeleteRequest(ctx, "/application.delete", req)
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

	return nil, newNotFoundError("compose", composeID)
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
	return c.doDeleteRequest(ctx, "/compose.delete", req)
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

	return nil, newNotFoundError("postgres", postgresID)
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
	return c.doDeleteRequest(ctx, "/postgres.delete", req)
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

	return nil, newNotFoundError("mysql", mysqlID)
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
	return c.doDeleteRequest(ctx, "/mysql.delete", req)
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

	return nil, newNotFoundError("mariadb", mariadbID)
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
	return c.doDeleteRequest(ctx, "/mariadb.delete", req)
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

	return nil, newNotFoundError("mongo", mongoID)
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
	return c.doDeleteRequest(ctx, "/mongo.delete", req)
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

	return nil, newNotFoundError("redis", redisID)
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
	return c.doDeleteRequest(ctx, "/redis.delete", req)
}

// =============================================================================
// Domain CRUD Operations
// =============================================================================

// CreateDomainRequest represents the request body for creating a domain
type CreateDomainRequest struct {
	Host               string  `json:"host"`
	Path               *string `json:"path,omitempty"`
	Port               *int    `json:"port,omitempty"`
	HTTPS              bool    `json:"https"`
	CertificateType    string  `json:"certificateType"`
	CustomCertResolver *string `json:"customCertResolver,omitempty"`
	ApplicationID      *string `json:"applicationId,omitempty"`
	ComposeID          *string `json:"composeId,omitempty"`
	ServiceName        *string `json:"serviceName,omitempty"`
	DomainType         *string `json:"domainType,omitempty"`
	InternalPath       *string `json:"internalPath,omitempty"`
	StripPath          bool    `json:"stripPath"`
}

// CreateDomainResponse represents the response from creating a domain
type CreateDomainResponse struct {
	DomainID string `json:"domainId"`
}

// UpdateDomainRequest represents the request body for updating a domain
type UpdateDomainRequest struct {
	DomainID           string  `json:"domainId"`
	Host               string  `json:"host"`
	Path               *string `json:"path,omitempty"`
	Port               *int    `json:"port,omitempty"`
	HTTPS              bool    `json:"https"`
	CertificateType    string  `json:"certificateType"`
	CustomCertResolver *string `json:"customCertResolver,omitempty"`
	ServiceName        *string `json:"serviceName,omitempty"`
	DomainType         *string `json:"domainType,omitempty"`
	InternalPath       *string `json:"internalPath,omitempty"`
	StripPath          bool    `json:"stripPath"`
}

// DeleteDomainRequest represents the request body for deleting a domain
type DeleteDomainRequest struct {
	DomainID string `json:"domainId"`
}

// GetDomainRequest represents the request body for getting a domain
type GetDomainRequest struct {
	DomainID string `json:"domainId"`
}

// CreateDomain creates a new domain
func (c *Client) CreateDomain(ctx context.Context, req CreateDomainRequest) (*CreateDomainResponse, error) {
	data, err := c.doPostRequest(ctx, "/domain.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating domain: %w", err)
	}

	var resp CreateDomainResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create domain response: %w", err)
	}

	return &resp, nil
}

// GetDomain fetches a single domain by ID
func (c *Client) GetDomain(ctx context.Context, domainID string) (*Domain, error) {
	req := GetDomainRequest{DomainID: domainID}
	data, err := c.doPostRequest(ctx, "/domain.one", req)
	if err != nil {
		return nil, fmt.Errorf("fetching domain: %w", err)
	}

	var domain Domain
	if err := json.Unmarshal(data, &domain); err != nil {
		return nil, fmt.Errorf("parsing domain: %w", err)
	}

	return &domain, nil
}

// UpdateDomain updates an existing domain
func (c *Client) UpdateDomain(ctx context.Context, req UpdateDomainRequest) error {
	_, err := c.doPostRequest(ctx, "/domain.update", req)
	if err != nil {
		return fmt.Errorf("updating domain: %w", err)
	}
	return nil
}

// DeleteDomain deletes a domain
func (c *Client) DeleteDomain(ctx context.Context, domainID string) error {
	req := DeleteDomainRequest{DomainID: domainID}
	return c.doDeleteRequest(ctx, "/domain.delete", req)
}

// =============================================================================
// Port CRUD Operations
// =============================================================================

// CreatePortRequest represents the request body for creating a port
type CreatePortRequest struct {
	PublishedPort int    `json:"publishedPort"`
	TargetPort    int    `json:"targetPort"`
	Protocol      string `json:"protocol"`
	PublishMode   string `json:"publishMode"`
	ApplicationID string `json:"applicationId"`
}

// CreatePortResponse represents the response from creating a port
type CreatePortResponse struct {
	PortID string `json:"portId"`
}

// UpdatePortRequest represents the request body for updating a port
type UpdatePortRequest struct {
	PortID        string `json:"portId"`
	PublishedPort int    `json:"publishedPort"`
	TargetPort    int    `json:"targetPort"`
	Protocol      string `json:"protocol"`
	PublishMode   string `json:"publishMode"`
}

// DeletePortRequest represents the request body for deleting a port
type DeletePortRequest struct {
	PortID string `json:"portId"`
}

// GetPortRequest represents the request body for getting a port
type GetPortRequest struct {
	PortID string `json:"portId"`
}

// CreatePort creates a new port
func (c *Client) CreatePort(ctx context.Context, req CreatePortRequest) (*CreatePortResponse, error) {
	data, err := c.doPostRequest(ctx, "/port.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating port: %w", err)
	}

	var resp CreatePortResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create port response: %w", err)
	}

	return &resp, nil
}

// GetPort fetches a single port by ID
func (c *Client) GetPort(ctx context.Context, portID string) (*Port, error) {
	req := GetPortRequest{PortID: portID}
	data, err := c.doPostRequest(ctx, "/port.one", req)
	if err != nil {
		return nil, fmt.Errorf("fetching port: %w", err)
	}

	var port Port
	if err := json.Unmarshal(data, &port); err != nil {
		return nil, fmt.Errorf("parsing port: %w", err)
	}

	return &port, nil
}

// UpdatePort updates an existing port
func (c *Client) UpdatePort(ctx context.Context, req UpdatePortRequest) error {
	_, err := c.doPostRequest(ctx, "/port.update", req)
	if err != nil {
		return fmt.Errorf("updating port: %w", err)
	}
	return nil
}

// DeletePort deletes a port
func (c *Client) DeletePort(ctx context.Context, portID string) error {
	req := DeletePortRequest{PortID: portID}
	return c.doDeleteRequest(ctx, "/port.delete", req)
}

// =============================================================================
// Mount CRUD Operations
// =============================================================================

// CreateMountRequest represents the request body for creating a mount
type CreateMountRequest struct {
	Type        string  `json:"type"`
	HostPath    *string `json:"hostPath,omitempty"`
	VolumeName  *string `json:"volumeName,omitempty"`
	Content     *string `json:"content,omitempty"`
	FilePath    *string `json:"filePath,omitempty"`
	MountPath   string  `json:"mountPath"`
	ServiceType string  `json:"serviceType"`
	ServiceID   string  `json:"serviceId"`
}

// CreateMountResponse represents the response from creating a mount
type CreateMountResponse struct {
	MountID string `json:"mountId"`
}

// UpdateMountRequest represents the request body for updating a mount
type UpdateMountRequest struct {
	MountID     string  `json:"mountId"`
	Type        *string `json:"type,omitempty"`
	HostPath    *string `json:"hostPath,omitempty"`
	VolumeName  *string `json:"volumeName,omitempty"`
	Content     *string `json:"content,omitempty"`
	FilePath    *string `json:"filePath,omitempty"`
	MountPath   *string `json:"mountPath,omitempty"`
	ServiceType *string `json:"serviceType,omitempty"`
}

// DeleteMountRequest represents the request body for deleting a mount
type DeleteMountRequest struct {
	MountID string `json:"mountId"`
}

// GetMountRequest represents the request body for getting a mount
type GetMountRequest struct {
	MountID string `json:"mountId"`
}

// CreateMount creates a new mount
func (c *Client) CreateMount(ctx context.Context, req CreateMountRequest) (*CreateMountResponse, error) {
	data, err := c.doPostRequest(ctx, "/mounts.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating mount: %w", err)
	}

	var resp CreateMountResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create mount response: %w", err)
	}

	return &resp, nil
}

// GetMount fetches a single mount by ID
func (c *Client) GetMount(ctx context.Context, mountID string) (*Mount, error) {
	req := GetMountRequest{MountID: mountID}
	data, err := c.doPostRequest(ctx, "/mounts.one", req)
	if err != nil {
		return nil, fmt.Errorf("fetching mount: %w", err)
	}

	var mount Mount
	if err := json.Unmarshal(data, &mount); err != nil {
		return nil, fmt.Errorf("parsing mount: %w", err)
	}

	return &mount, nil
}

// UpdateMount updates an existing mount
func (c *Client) UpdateMount(ctx context.Context, req UpdateMountRequest) error {
	_, err := c.doPostRequest(ctx, "/mounts.update", req)
	if err != nil {
		return fmt.Errorf("updating mount: %w", err)
	}
	return nil
}

// DeleteMount deletes a mount
func (c *Client) DeleteMount(ctx context.Context, mountID string) error {
	req := DeleteMountRequest{MountID: mountID}
	return c.doDeleteRequest(ctx, "/mounts.remove", req)
}

// =============================================================================
// Security CRUD Operations
// =============================================================================

// CreateSecurityRequest represents the request body for creating a security entry
type CreateSecurityRequest struct {
	ApplicationID string `json:"applicationId"`
	Username      string `json:"username"`
	Password      string `json:"password"`
}

// CreateSecurityResponse represents the response from creating a security entry
type CreateSecurityResponse struct {
	SecurityID string `json:"securityId"`
}

// UpdateSecurityRequest represents the request body for updating a security entry
type UpdateSecurityRequest struct {
	SecurityID string `json:"securityId"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// DeleteSecurityRequest represents the request body for deleting a security entry
type DeleteSecurityRequest struct {
	SecurityID string `json:"securityId"`
}

// GetSecurityRequest represents the request body for getting a security entry
type GetSecurityRequest struct {
	SecurityID string `json:"securityId"`
}

// Security represents a security (basic auth) configuration
type Security struct {
	SecurityID    string `json:"securityId"`
	Username      string `json:"username"`
	Password      string `json:"password"`
	ApplicationID string `json:"applicationId"`
}

// CreateSecurity creates a new security entry
func (c *Client) CreateSecurity(ctx context.Context, req CreateSecurityRequest) (*CreateSecurityResponse, error) {
	data, err := c.doPostRequest(ctx, "/security.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating security: %w", err)
	}

	var resp CreateSecurityResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create security response: %w", err)
	}

	return &resp, nil
}

// GetSecurity fetches a single security entry by ID
func (c *Client) GetSecurity(ctx context.Context, securityID string) (*Security, error) {
	req := GetSecurityRequest{SecurityID: securityID}
	data, err := c.doPostRequest(ctx, "/security.one", req)
	if err != nil {
		return nil, fmt.Errorf("fetching security: %w", err)
	}

	var security Security
	if err := json.Unmarshal(data, &security); err != nil {
		return nil, fmt.Errorf("parsing security: %w", err)
	}

	return &security, nil
}

// UpdateSecurity updates an existing security entry
func (c *Client) UpdateSecurity(ctx context.Context, req UpdateSecurityRequest) error {
	_, err := c.doPostRequest(ctx, "/security.update", req)
	if err != nil {
		return fmt.Errorf("updating security: %w", err)
	}
	return nil
}

// DeleteSecurity deletes a security entry
func (c *Client) DeleteSecurity(ctx context.Context, securityID string) error {
	req := DeleteSecurityRequest{SecurityID: securityID}
	return c.doDeleteRequest(ctx, "/security.delete", req)
}

// =============================================================================
// Redirect CRUD Operations
// =============================================================================

// CreateRedirectRequest represents the request body for creating a redirect
type CreateRedirectRequest struct {
	Regex         string `json:"regex"`
	Replacement   string `json:"replacement"`
	Permanent     bool   `json:"permanent"`
	ApplicationID string `json:"applicationId"`
}

// CreateRedirectResponse represents the response from creating a redirect
type CreateRedirectResponse struct {
	RedirectID string `json:"redirectId"`
}

// UpdateRedirectRequest represents the request body for updating a redirect
type UpdateRedirectRequest struct {
	RedirectID  string `json:"redirectId"`
	Regex       string `json:"regex"`
	Replacement string `json:"replacement"`
	Permanent   bool   `json:"permanent"`
}

// DeleteRedirectRequest represents the request body for deleting a redirect
type DeleteRedirectRequest struct {
	RedirectID string `json:"redirectId"`
}

// GetRedirectRequest represents the request body for getting a redirect
type GetRedirectRequest struct {
	RedirectID string `json:"redirectId"`
}

// Redirect represents a redirect configuration
type Redirect struct {
	RedirectID    string `json:"redirectId"`
	Regex         string `json:"regex"`
	Replacement   string `json:"replacement"`
	Permanent     bool   `json:"permanent"`
	ApplicationID string `json:"applicationId"`
}

// CreateRedirect creates a new redirect
func (c *Client) CreateRedirect(ctx context.Context, req CreateRedirectRequest) (*CreateRedirectResponse, error) {
	data, err := c.doPostRequest(ctx, "/redirects.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating redirect: %w", err)
	}

	var resp CreateRedirectResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create redirect response: %w", err)
	}

	return &resp, nil
}

// GetRedirect fetches a single redirect by ID
func (c *Client) GetRedirect(ctx context.Context, redirectID string) (*Redirect, error) {
	req := GetRedirectRequest{RedirectID: redirectID}
	data, err := c.doPostRequest(ctx, "/redirects.one", req)
	if err != nil {
		return nil, fmt.Errorf("fetching redirect: %w", err)
	}

	var redirect Redirect
	if err := json.Unmarshal(data, &redirect); err != nil {
		return nil, fmt.Errorf("parsing redirect: %w", err)
	}

	return &redirect, nil
}

// UpdateRedirect updates an existing redirect
func (c *Client) UpdateRedirect(ctx context.Context, req UpdateRedirectRequest) error {
	_, err := c.doPostRequest(ctx, "/redirects.update", req)
	if err != nil {
		return fmt.Errorf("updating redirect: %w", err)
	}
	return nil
}

// DeleteRedirect deletes a redirect
func (c *Client) DeleteRedirect(ctx context.Context, redirectID string) error {
	req := DeleteRedirectRequest{RedirectID: redirectID}
	return c.doDeleteRequest(ctx, "/redirects.delete", req)
}

// =============================================================================
// Backup CRUD Operations
// =============================================================================

// CreateBackupRequest represents the request body for creating a backup
type CreateBackupRequest struct {
	Schedule        string  `json:"schedule"`
	Enabled         *bool   `json:"enabled,omitempty"`
	Prefix          string  `json:"prefix"`
	DestinationID   string  `json:"destinationId"`
	KeepLatestCount *int    `json:"keepLatestCount,omitempty"`
	Database        string  `json:"database"`
	DatabaseType    string  `json:"databaseType"`
	BackupType      string  `json:"backupType"`
	PostgresID      *string `json:"postgresId,omitempty"`
	MysqlID         *string `json:"mysqlId,omitempty"`
	MariadbID       *string `json:"mariadbId,omitempty"`
	MongoID         *string `json:"mongoId,omitempty"`
	ComposeID       *string `json:"composeId,omitempty"`
	ServiceName     *string `json:"serviceName,omitempty"`
}

// CreateBackupResponse represents the response from creating a backup
type CreateBackupResponse struct {
	BackupID string `json:"backupId"`
}

// UpdateBackupRequest represents the request body for updating a backup
type UpdateBackupRequest struct {
	BackupID        string  `json:"backupId"`
	Schedule        *string `json:"schedule,omitempty"`
	Enabled         *bool   `json:"enabled,omitempty"`
	Prefix          *string `json:"prefix,omitempty"`
	DestinationID   *string `json:"destinationId,omitempty"`
	KeepLatestCount *int    `json:"keepLatestCount,omitempty"`
	Database        *string `json:"database,omitempty"`
}

// DeleteBackupRequest represents the request body for deleting a backup
type DeleteBackupRequest struct {
	BackupID string `json:"backupId"`
}

// GetBackupRequest represents the request body for getting a backup
type GetBackupRequest struct {
	BackupID string `json:"backupId"`
}

// Backup represents a backup configuration
type Backup struct {
	BackupID        string  `json:"backupId"`
	Schedule        string  `json:"schedule"`
	Enabled         bool    `json:"enabled"`
	Prefix          string  `json:"prefix"`
	DestinationID   string  `json:"destinationId"`
	KeepLatestCount *int    `json:"keepLatestCount"`
	Database        string  `json:"database"`
	DatabaseType    string  `json:"databaseType"`
	BackupType      string  `json:"backupType"`
	PostgresID      *string `json:"postgresId"`
	MysqlID         *string `json:"mysqlId"`
	MariadbID       *string `json:"mariadbId"`
	MongoID         *string `json:"mongoId"`
	ComposeID       *string `json:"composeId"`
	ServiceName     *string `json:"serviceName"`
}

// CreateBackup creates a new backup
func (c *Client) CreateBackup(ctx context.Context, req CreateBackupRequest) (*CreateBackupResponse, error) {
	data, err := c.doPostRequest(ctx, "/backup.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating backup: %w", err)
	}

	var resp CreateBackupResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create backup response: %w", err)
	}

	return &resp, nil
}

// GetBackup fetches a single backup by ID
func (c *Client) GetBackup(ctx context.Context, backupID string) (*Backup, error) {
	req := GetBackupRequest{BackupID: backupID}
	data, err := c.doPostRequest(ctx, "/backup.one", req)
	if err != nil {
		return nil, fmt.Errorf("fetching backup: %w", err)
	}

	var backup Backup
	if err := json.Unmarshal(data, &backup); err != nil {
		return nil, fmt.Errorf("parsing backup: %w", err)
	}

	return &backup, nil
}

// UpdateBackup updates an existing backup
func (c *Client) UpdateBackup(ctx context.Context, req UpdateBackupRequest) error {
	_, err := c.doPostRequest(ctx, "/backup.update", req)
	if err != nil {
		return fmt.Errorf("updating backup: %w", err)
	}
	return nil
}

// DeleteBackup deletes a backup
func (c *Client) DeleteBackup(ctx context.Context, backupID string) error {
	req := DeleteBackupRequest{BackupID: backupID}
	return c.doDeleteRequest(ctx, "/backup.remove", req)
}

// =============================================================================
// Schedule CRUD Operations
// =============================================================================

// CreateScheduleRequest represents the request body for creating a schedule
type CreateScheduleRequest struct {
	Name           string  `json:"name"`
	CronExpression string  `json:"cronExpression"`
	Command        string  `json:"command"`
	ShellType      string  `json:"shellType"`
	ScheduleType   string  `json:"scheduleType"`
	AppName        *string `json:"appName,omitempty"`
	ServiceName    *string `json:"serviceName,omitempty"`
	Script         *string `json:"script,omitempty"`
	ApplicationID  *string `json:"applicationId,omitempty"`
	ComposeID      *string `json:"composeId,omitempty"`
	ServerID       *string `json:"serverId,omitempty"`
	Enabled        bool    `json:"enabled"`
	Timezone       *string `json:"timezone,omitempty"`
}

// CreateScheduleResponse represents the response from creating a schedule
type CreateScheduleResponse struct {
	ScheduleID string `json:"scheduleId"`
}

// UpdateScheduleRequest represents the request body for updating a schedule
type UpdateScheduleRequest struct {
	ScheduleID     string  `json:"scheduleId"`
	Name           *string `json:"name,omitempty"`
	CronExpression *string `json:"cronExpression,omitempty"`
	Command        *string `json:"command,omitempty"`
	ShellType      *string `json:"shellType,omitempty"`
	Script         *string `json:"script,omitempty"`
	Enabled        *bool   `json:"enabled,omitempty"`
	Timezone       *string `json:"timezone,omitempty"`
}

// DeleteScheduleRequest represents the request body for deleting a schedule
type DeleteScheduleRequest struct {
	ScheduleID string `json:"scheduleId"`
}

// GetScheduleRequest represents the request body for getting a schedule
type GetScheduleRequest struct {
	ScheduleID string `json:"scheduleId"`
}

// Schedule represents a scheduled task configuration
type Schedule struct {
	ScheduleID     string  `json:"scheduleId"`
	Name           string  `json:"name"`
	CronExpression string  `json:"cronExpression"`
	Command        string  `json:"command"`
	ShellType      string  `json:"shellType"`
	ScheduleType   string  `json:"scheduleType"`
	AppName        string  `json:"appName"`
	ServiceName    *string `json:"serviceName"`
	Script         *string `json:"script"`
	ApplicationID  *string `json:"applicationId"`
	ComposeID      *string `json:"composeId"`
	ServerID       *string `json:"serverId"`
	Enabled        bool    `json:"enabled"`
	Timezone       *string `json:"timezone"`
}

// CreateSchedule creates a new schedule
func (c *Client) CreateSchedule(ctx context.Context, req CreateScheduleRequest) (*CreateScheduleResponse, error) {
	data, err := c.doPostRequest(ctx, "/schedule.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating schedule: %w", err)
	}

	var resp CreateScheduleResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create schedule response: %w", err)
	}

	return &resp, nil
}

// GetSchedule fetches a single schedule by ID
func (c *Client) GetSchedule(ctx context.Context, scheduleID string) (*Schedule, error) {
	req := GetScheduleRequest{ScheduleID: scheduleID}
	data, err := c.doPostRequest(ctx, "/schedule.one", req)
	if err != nil {
		return nil, fmt.Errorf("fetching schedule: %w", err)
	}

	var schedule Schedule
	if err := json.Unmarshal(data, &schedule); err != nil {
		return nil, fmt.Errorf("parsing schedule: %w", err)
	}

	return &schedule, nil
}

// UpdateSchedule updates an existing schedule
func (c *Client) UpdateSchedule(ctx context.Context, req UpdateScheduleRequest) error {
	_, err := c.doPostRequest(ctx, "/schedule.update", req)
	if err != nil {
		return fmt.Errorf("updating schedule: %w", err)
	}
	return nil
}

// DeleteSchedule deletes a schedule
func (c *Client) DeleteSchedule(ctx context.Context, scheduleID string) error {
	req := DeleteScheduleRequest{ScheduleID: scheduleID}
	return c.doDeleteRequest(ctx, "/schedule.delete", req)
}

// =============================================================================
// Notification CRUD Operations
// =============================================================================

// NotificationBase contains common fields for all notification types
type NotificationBase struct {
	Name            string `json:"name"`
	AppDeploy       bool   `json:"appDeploy"`
	AppBuildError   bool   `json:"appBuildError"`
	DatabaseBackup  bool   `json:"databaseBackup"`
	VolumeBackup    bool   `json:"volumeBackup"`
	DokployRestart  bool   `json:"dokployRestart"`
	DockerCleanup   bool   `json:"dockerCleanup"`
	ServerThreshold bool   `json:"serverThreshold"`
}

// CreateSlackNotificationRequest represents the request for creating a Slack notification
type CreateSlackNotificationRequest struct {
	NotificationBase
	WebhookURL string `json:"webhookUrl"`
	Channel    string `json:"channel"`
}

// CreateDiscordNotificationRequest represents the request for creating a Discord notification
type CreateDiscordNotificationRequest struct {
	NotificationBase
	WebhookURL string `json:"webhookUrl"`
	Decoration bool   `json:"decoration"`
}

// CreateTelegramNotificationRequest represents the request for creating a Telegram notification
type CreateTelegramNotificationRequest struct {
	NotificationBase
	BotToken        string `json:"botToken"`
	ChatID          string `json:"chatId"`
	MessageThreadID string `json:"messageThreadId"`
}

// CreateEmailNotificationRequest represents the request for creating an Email notification
type CreateEmailNotificationRequest struct {
	NotificationBase
	SmtpServer  string   `json:"smtpServer"`
	SmtpPort    int      `json:"smtpPort"`
	Username    string   `json:"username"`
	Password    string   `json:"password"`
	FromAddress string   `json:"fromAddress"`
	ToAddresses []string `json:"toAddresses"`
}

// UpdateSlackNotificationRequest represents the request for updating a Slack notification
type UpdateSlackNotificationRequest struct {
	NotificationID string `json:"notificationId"`
	CreateSlackNotificationRequest
}

// UpdateDiscordNotificationRequest represents the request for updating a Discord notification
type UpdateDiscordNotificationRequest struct {
	NotificationID string `json:"notificationId"`
	CreateDiscordNotificationRequest
}

// UpdateTelegramNotificationRequest represents the request for updating a Telegram notification
type UpdateTelegramNotificationRequest struct {
	NotificationID string `json:"notificationId"`
	CreateTelegramNotificationRequest
}

// UpdateEmailNotificationRequest represents the request for updating an Email notification
type UpdateEmailNotificationRequest struct {
	NotificationID string `json:"notificationId"`
	CreateEmailNotificationRequest
}

// CreateNotificationResponse represents the response from creating a notification
type CreateNotificationResponse struct {
	NotificationID string `json:"notificationId"`
}

// DeleteNotificationRequest represents the request for deleting a notification
type DeleteNotificationRequest struct {
	NotificationID string `json:"notificationId"`
}

// NotificationResponse represents a notification from the API
type NotificationResponse struct {
	NotificationID  string  `json:"notificationId"`
	Name            string  `json:"name"`
	NotificationType string `json:"notificationType"`
	AppDeploy       bool    `json:"appDeploy"`
	AppBuildError   bool    `json:"appBuildError"`
	DatabaseBackup  bool    `json:"databaseBackup"`
	VolumeBackup    bool    `json:"volumeBackup"`
	DokployRestart  bool    `json:"dokployRestart"`
	DockerCleanup   bool    `json:"dockerCleanup"`
	ServerThreshold bool    `json:"serverThreshold"`
	OrganizationID  string  `json:"organizationId"`
	// Slack fields
	WebhookURL *string `json:"webhookUrl,omitempty"`
	Channel    *string `json:"channel,omitempty"`
	// Discord fields
	Decoration *bool `json:"decoration,omitempty"`
	// Telegram fields
	BotToken        *string `json:"botToken,omitempty"`
	ChatID          *string `json:"chatId,omitempty"`
	MessageThreadID *string `json:"messageThreadId,omitempty"`
	// Email fields
	SmtpServer  *string  `json:"smtpServer,omitempty"`
	SmtpPort    *int     `json:"smtpPort,omitempty"`
	Username    *string  `json:"username,omitempty"`
	Password    *string  `json:"password,omitempty"`
	FromAddress *string  `json:"fromAddress,omitempty"`
	ToAddresses []string `json:"toAddresses,omitempty"`
}

// CreateSlackNotification creates a Slack notification
func (c *Client) CreateSlackNotification(ctx context.Context, req CreateSlackNotificationRequest) (*CreateNotificationResponse, error) {
	data, err := c.doPostRequest(ctx, "/notification.createSlack", req)
	if err != nil {
		return nil, fmt.Errorf("creating slack notification: %w", err)
	}
	var resp CreateNotificationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create notification response: %w", err)
	}
	return &resp, nil
}

// CreateDiscordNotification creates a Discord notification
func (c *Client) CreateDiscordNotification(ctx context.Context, req CreateDiscordNotificationRequest) (*CreateNotificationResponse, error) {
	data, err := c.doPostRequest(ctx, "/notification.createDiscord", req)
	if err != nil {
		return nil, fmt.Errorf("creating discord notification: %w", err)
	}
	var resp CreateNotificationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create notification response: %w", err)
	}
	return &resp, nil
}

// CreateTelegramNotification creates a Telegram notification
func (c *Client) CreateTelegramNotification(ctx context.Context, req CreateTelegramNotificationRequest) (*CreateNotificationResponse, error) {
	data, err := c.doPostRequest(ctx, "/notification.createTelegram", req)
	if err != nil {
		return nil, fmt.Errorf("creating telegram notification: %w", err)
	}
	var resp CreateNotificationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create notification response: %w", err)
	}
	return &resp, nil
}

// CreateEmailNotification creates an Email notification
func (c *Client) CreateEmailNotification(ctx context.Context, req CreateEmailNotificationRequest) (*CreateNotificationResponse, error) {
	data, err := c.doPostRequest(ctx, "/notification.createEmail", req)
	if err != nil {
		return nil, fmt.Errorf("creating email notification: %w", err)
	}
	var resp CreateNotificationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create notification response: %w", err)
	}
	return &resp, nil
}

// GetNotification fetches a single notification by ID
func (c *Client) GetNotification(ctx context.Context, notificationID string) (*NotificationResponse, error) {
	endpoint := fmt.Sprintf("/notification.one?notificationId=%s", notificationID)
	data, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching notification: %w", err)
	}
	var notification NotificationResponse
	if err := json.Unmarshal(data, &notification); err != nil {
		return nil, fmt.Errorf("parsing notification: %w", err)
	}
	return &notification, nil
}

// UpdateSlackNotification updates a Slack notification
func (c *Client) UpdateSlackNotification(ctx context.Context, req UpdateSlackNotificationRequest) error {
	_, err := c.doPostRequest(ctx, "/notification.updateSlack", req)
	if err != nil {
		return fmt.Errorf("updating slack notification: %w", err)
	}
	return nil
}

// UpdateDiscordNotification updates a Discord notification
func (c *Client) UpdateDiscordNotification(ctx context.Context, req UpdateDiscordNotificationRequest) error {
	_, err := c.doPostRequest(ctx, "/notification.updateDiscord", req)
	if err != nil {
		return fmt.Errorf("updating discord notification: %w", err)
	}
	return nil
}

// UpdateTelegramNotification updates a Telegram notification
func (c *Client) UpdateTelegramNotification(ctx context.Context, req UpdateTelegramNotificationRequest) error {
	_, err := c.doPostRequest(ctx, "/notification.updateTelegram", req)
	if err != nil {
		return fmt.Errorf("updating telegram notification: %w", err)
	}
	return nil
}

// UpdateEmailNotification updates an Email notification
func (c *Client) UpdateEmailNotification(ctx context.Context, req UpdateEmailNotificationRequest) error {
	_, err := c.doPostRequest(ctx, "/notification.updateEmail", req)
	if err != nil {
		return fmt.Errorf("updating email notification: %w", err)
	}
	return nil
}

// DeleteNotification deletes a notification
func (c *Client) DeleteNotification(ctx context.Context, notificationID string) error {
	req := DeleteNotificationRequest{NotificationID: notificationID}
	return c.doDeleteRequest(ctx, "/notification.remove", req)
}

// =============================================================================
// Organization CRUD Operations
// =============================================================================

// CreateOrganizationRequest represents the request for creating an organization
type CreateOrganizationRequest struct {
	Name string  `json:"name"`
	Logo *string `json:"logo,omitempty"`
}

// CreateOrganizationResponse represents the response from creating an organization
type CreateOrganizationResponse struct {
	OrganizationID string `json:"organizationId"`
}

// UpdateOrganizationRequest represents the request for updating an organization
type UpdateOrganizationRequest struct {
	OrganizationID string  `json:"organizationId"`
	Name           string  `json:"name"`
	Logo           *string `json:"logo,omitempty"`
}

// DeleteOrganizationRequest represents the request for deleting an organization
type DeleteOrganizationRequest struct {
	OrganizationID string `json:"organizationId"`
}

// Organization represents an organization
type Organization struct {
	OrganizationID string `json:"organizationId"`
	Name           string `json:"name"`
	Logo           string `json:"logo"`
}

// CreateOrganization creates a new organization
func (c *Client) CreateOrganization(ctx context.Context, req CreateOrganizationRequest) (*CreateOrganizationResponse, error) {
	data, err := c.doPostRequest(ctx, "/organization.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating organization: %w", err)
	}
	var resp CreateOrganizationResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create organization response: %w", err)
	}
	return &resp, nil
}

// GetOrganization fetches a single organization by ID
func (c *Client) GetOrganization(ctx context.Context, organizationID string) (*Organization, error) {
	endpoint := fmt.Sprintf("/organization.one?organizationId=%s", organizationID)
	data, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching organization: %w", err)
	}
	var org Organization
	if err := json.Unmarshal(data, &org); err != nil {
		return nil, fmt.Errorf("parsing organization: %w", err)
	}
	return &org, nil
}

// UpdateOrganization updates an existing organization
func (c *Client) UpdateOrganization(ctx context.Context, req UpdateOrganizationRequest) error {
	_, err := c.doPostRequest(ctx, "/organization.update", req)
	if err != nil {
		return fmt.Errorf("updating organization: %w", err)
	}
	return nil
}

// DeleteOrganization deletes an organization
func (c *Client) DeleteOrganization(ctx context.Context, organizationID string) error {
	req := DeleteOrganizationRequest{OrganizationID: organizationID}
	return c.doDeleteRequest(ctx, "/organization.delete", req)
}

// --- GitLab Provider ---

// CreateGitlabRequest is the request to create a GitLab provider
type CreateGitlabRequest struct {
	Name         string  `json:"name"`
	GitlabURL    string  `json:"gitlabUrl"`
	AuthID       string  `json:"authId"`
	ApplicationID *string `json:"applicationId,omitempty"`
	RedirectURI  *string `json:"redirectUri,omitempty"`
	Secret       *string `json:"secret,omitempty"`
	AccessToken  *string `json:"accessToken,omitempty"`
	RefreshToken *string `json:"refreshToken,omitempty"`
	GroupName    *string `json:"groupName,omitempty"`
	ExpiresAt    *int64  `json:"expiresAt,omitempty"`
}

// UpdateGitlabRequest is the request to update a GitLab provider
type UpdateGitlabRequest struct {
	GitlabID      string  `json:"gitlabId"`
	GitProviderID string  `json:"gitProviderId"`
	Name          string  `json:"name"`
	GitlabURL     string  `json:"gitlabUrl"`
	ApplicationID *string `json:"applicationId,omitempty"`
	RedirectURI   *string `json:"redirectUri,omitempty"`
	Secret        *string `json:"secret,omitempty"`
	AccessToken   *string `json:"accessToken,omitempty"`
	RefreshToken  *string `json:"refreshToken,omitempty"`
	GroupName     *string `json:"groupName,omitempty"`
	ExpiresAt     *int64  `json:"expiresAt,omitempty"`
}

// GitlabResponse is the response when fetching a GitLab provider
type GitlabResponse struct {
	GitlabID      string  `json:"gitlabId"`
	GitProviderID string  `json:"gitProviderId"`
	Name          string  `json:"name"`
	GitlabURL     string  `json:"gitlabUrl"`
	ApplicationID *string `json:"applicationId,omitempty"`
	RedirectURI   *string `json:"redirectUri,omitempty"`
	Secret        *string `json:"secret,omitempty"`
	AccessToken   *string `json:"accessToken,omitempty"`
	RefreshToken  *string `json:"refreshToken,omitempty"`
	GroupName     *string `json:"groupName,omitempty"`
	ExpiresAt     *int64  `json:"expiresAt,omitempty"`
}

// CreateGitlab creates a new GitLab provider
func (c *Client) CreateGitlab(ctx context.Context, req CreateGitlabRequest) (*GitlabResponse, error) {
	data, err := c.doPostRequest(ctx, "/gitlab.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating gitlab provider: %w", err)
	}
	var resp GitlabResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create gitlab response: %w", err)
	}
	return &resp, nil
}

// GetGitlab fetches a GitLab provider by ID
func (c *Client) GetGitlab(ctx context.Context, gitlabID string) (*GitlabResponse, error) {
	endpoint := fmt.Sprintf("/gitlab.one?gitlabId=%s", gitlabID)
	data, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching gitlab provider: %w", err)
	}
	var resp GitlabResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing gitlab response: %w", err)
	}
	return &resp, nil
}

// UpdateGitlab updates an existing GitLab provider
func (c *Client) UpdateGitlab(ctx context.Context, req UpdateGitlabRequest) error {
	_, err := c.doPostRequest(ctx, "/gitlab.update", req)
	if err != nil {
		return fmt.Errorf("updating gitlab provider: %w", err)
	}
	return nil
}

// DeleteGitProvider deletes a git provider (used for gitlab, bitbucket, gitea)
func (c *Client) DeleteGitProvider(ctx context.Context, gitProviderID string) error {
	req := struct {
		GitProviderID string `json:"gitProviderId"`
	}{GitProviderID: gitProviderID}
	_, err := c.doPostRequest(ctx, "/gitProvider.remove", req)
	if err != nil {
		return fmt.Errorf("deleting git provider: %w", err)
	}
	return nil
}

// --- Bitbucket Provider ---

// CreateBitbucketRequest is the request to create a Bitbucket provider
type CreateBitbucketRequest struct {
	Name                   string  `json:"name"`
	AuthID                 string  `json:"authId"`
	BitbucketUsername      *string `json:"bitbucketUsername,omitempty"`
	BitbucketEmail         *string `json:"bitbucketEmail,omitempty"`
	AppPassword            *string `json:"appPassword,omitempty"`
	ApiToken               *string `json:"apiToken,omitempty"`
	BitbucketWorkspaceName *string `json:"bitbucketWorkspaceName,omitempty"`
}

// UpdateBitbucketRequest is the request to update a Bitbucket provider
type UpdateBitbucketRequest struct {
	BitbucketID            string  `json:"bitbucketId"`
	GitProviderID          string  `json:"gitProviderId"`
	Name                   string  `json:"name"`
	BitbucketUsername      *string `json:"bitbucketUsername,omitempty"`
	BitbucketEmail         *string `json:"bitbucketEmail,omitempty"`
	AppPassword            *string `json:"appPassword,omitempty"`
	ApiToken               *string `json:"apiToken,omitempty"`
	BitbucketWorkspaceName *string `json:"bitbucketWorkspaceName,omitempty"`
	OrganizationID         *string `json:"organizationId,omitempty"`
}

// BitbucketResponse is the response when fetching a Bitbucket provider
type BitbucketResponse struct {
	BitbucketID            string  `json:"bitbucketId"`
	GitProviderID          string  `json:"gitProviderId"`
	Name                   string  `json:"name"`
	BitbucketUsername      *string `json:"bitbucketUsername,omitempty"`
	BitbucketEmail         *string `json:"bitbucketEmail,omitempty"`
	AppPassword            *string `json:"appPassword,omitempty"`
	ApiToken               *string `json:"apiToken,omitempty"`
	BitbucketWorkspaceName *string `json:"bitbucketWorkspaceName,omitempty"`
	OrganizationID         *string `json:"organizationId,omitempty"`
}

// CreateBitbucket creates a new Bitbucket provider
func (c *Client) CreateBitbucket(ctx context.Context, req CreateBitbucketRequest) (*BitbucketResponse, error) {
	data, err := c.doPostRequest(ctx, "/bitbucket.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating bitbucket provider: %w", err)
	}
	var resp BitbucketResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create bitbucket response: %w", err)
	}
	return &resp, nil
}

// GetBitbucket fetches a Bitbucket provider by ID
func (c *Client) GetBitbucket(ctx context.Context, bitbucketID string) (*BitbucketResponse, error) {
	endpoint := fmt.Sprintf("/bitbucket.one?bitbucketId=%s", bitbucketID)
	data, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching bitbucket provider: %w", err)
	}
	var resp BitbucketResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing bitbucket response: %w", err)
	}
	return &resp, nil
}

// UpdateBitbucket updates an existing Bitbucket provider
func (c *Client) UpdateBitbucket(ctx context.Context, req UpdateBitbucketRequest) error {
	_, err := c.doPostRequest(ctx, "/bitbucket.update", req)
	if err != nil {
		return fmt.Errorf("updating bitbucket provider: %w", err)
	}
	return nil
}

// --- Gitea Provider ---

// CreateGiteaRequest is the request to create a Gitea provider
type CreateGiteaRequest struct {
	Name             string  `json:"name"`
	GiteaURL         string  `json:"giteaUrl"`
	RedirectURI      *string `json:"redirectUri,omitempty"`
	ClientID         *string `json:"clientId,omitempty"`
	ClientSecret     *string `json:"clientSecret,omitempty"`
	AccessToken      *string `json:"accessToken,omitempty"`
	RefreshToken     *string `json:"refreshToken,omitempty"`
	ExpiresAt        *int64  `json:"expiresAt,omitempty"`
	Scopes           *string `json:"scopes,omitempty"`
	GiteaUsername    *string `json:"giteaUsername,omitempty"`
	OrganizationName *string `json:"organizationName,omitempty"`
}

// UpdateGiteaRequest is the request to update a Gitea provider
type UpdateGiteaRequest struct {
	GiteaID          string  `json:"giteaId"`
	GitProviderID    string  `json:"gitProviderId"`
	Name             string  `json:"name"`
	GiteaURL         string  `json:"giteaUrl"`
	RedirectURI      *string `json:"redirectUri,omitempty"`
	ClientID         *string `json:"clientId,omitempty"`
	ClientSecret     *string `json:"clientSecret,omitempty"`
	AccessToken      *string `json:"accessToken,omitempty"`
	RefreshToken     *string `json:"refreshToken,omitempty"`
	ExpiresAt        *int64  `json:"expiresAt,omitempty"`
	Scopes           *string `json:"scopes,omitempty"`
	GiteaUsername    *string `json:"giteaUsername,omitempty"`
	OrganizationName *string `json:"organizationName,omitempty"`
}

// GiteaResponse is the response when fetching a Gitea provider
type GiteaResponse struct {
	GiteaID          string  `json:"giteaId"`
	GitProviderID    string  `json:"gitProviderId"`
	Name             string  `json:"name"`
	GiteaURL         string  `json:"giteaUrl"`
	RedirectURI      *string `json:"redirectUri,omitempty"`
	ClientID         *string `json:"clientId,omitempty"`
	ClientSecret     *string `json:"clientSecret,omitempty"`
	AccessToken      *string `json:"accessToken,omitempty"`
	RefreshToken     *string `json:"refreshToken,omitempty"`
	ExpiresAt        *int64  `json:"expiresAt,omitempty"`
	Scopes           *string `json:"scopes,omitempty"`
	GiteaUsername    *string `json:"giteaUsername,omitempty"`
	OrganizationName *string `json:"organizationName,omitempty"`
}

// CreateGitea creates a new Gitea provider
func (c *Client) CreateGitea(ctx context.Context, req CreateGiteaRequest) (*GiteaResponse, error) {
	data, err := c.doPostRequest(ctx, "/gitea.create", req)
	if err != nil {
		return nil, fmt.Errorf("creating gitea provider: %w", err)
	}
	var resp GiteaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing create gitea response: %w", err)
	}
	return &resp, nil
}

// GetGitea fetches a Gitea provider by ID
func (c *Client) GetGitea(ctx context.Context, giteaID string) (*GiteaResponse, error) {
	endpoint := fmt.Sprintf("/gitea.one?giteaId=%s", giteaID)
	data, err := c.doRequest(ctx, endpoint)
	if err != nil {
		return nil, fmt.Errorf("fetching gitea provider: %w", err)
	}
	var resp GiteaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("parsing gitea response: %w", err)
	}
	return &resp, nil
}

// UpdateGitea updates an existing Gitea provider
func (c *Client) UpdateGitea(ctx context.Context, req UpdateGiteaRequest) error {
	_, err := c.doPostRequest(ctx, "/gitea.update", req)
	if err != nil {
		return fmt.Errorf("updating gitea provider: %w", err)
	}
	return nil
}

// --- User Management ---

// User represents a Dokploy user
type User struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	Image            *string `json:"image,omitempty"`
	IsRegistered     bool    `json:"isRegistered"`
	EmailVerified    bool    `json:"emailVerified"`
	TwoFactorEnabled *bool   `json:"twoFactorEnabled,omitempty"`
	Banned           *bool   `json:"banned,omitempty"`
	BanReason        *string `json:"banReason,omitempty"`
	CreatedAt        *string `json:"createdAt,omitempty"`
	UpdatedAt        *string `json:"updatedAt,omitempty"`
	Role             *string `json:"role,omitempty"`
	// Permissions
	AccessedProjects        []string `json:"accessedProjects,omitempty"`
	AccessedEnvironments    []string `json:"accessedEnvironments,omitempty"`
	AccessedServices        []string `json:"accessedServices,omitempty"`
	CanCreateProjects       *bool    `json:"canCreateProjects,omitempty"`
	CanCreateServices       *bool    `json:"canCreateServices,omitempty"`
	CanDeleteProjects       *bool    `json:"canDeleteProjects,omitempty"`
	CanDeleteServices       *bool    `json:"canDeleteServices,omitempty"`
	CanAccessToDocker       *bool    `json:"canAccessToDocker,omitempty"`
	CanAccessToTraefikFiles *bool    `json:"canAccessToTraefikFiles,omitempty"`
	CanAccessToAPI          *bool    `json:"canAccessToAPI,omitempty"`
	CanAccessToSSHKeys      *bool    `json:"canAccessToSSHKeys,omitempty"`
	CanAccessToGitProviders *bool    `json:"canAccessToGitProviders,omitempty"`
	CanDeleteEnvironments   *bool    `json:"canDeleteEnvironments,omitempty"`
	CanCreateEnvironments   *bool    `json:"canCreateEnvironments,omitempty"`
}

// UpdateUserRequest is the request to update a user
type UpdateUserRequest struct {
	ID       string  `json:"id"`
	Name     *string `json:"name,omitempty"`
	Email    *string `json:"email,omitempty"`
	Image    *string `json:"image,omitempty"`
	Password *string `json:"password,omitempty"`
}

// UserPermissionsRequest is the request to assign user permissions
type UserPermissionsRequest struct {
	ID                      string   `json:"id"`
	AccessedProjects        []string `json:"accessedProjects"`
	AccessedEnvironments    []string `json:"accessedEnvironments"`
	AccessedServices        []string `json:"accessedServices"`
	CanCreateProjects       bool     `json:"canCreateProjects"`
	CanCreateServices       bool     `json:"canCreateServices"`
	CanDeleteProjects       bool     `json:"canDeleteProjects"`
	CanDeleteServices       bool     `json:"canDeleteServices"`
	CanAccessToDocker       bool     `json:"canAccessToDocker"`
	CanAccessToTraefikFiles bool     `json:"canAccessToTraefikFiles"`
	CanAccessToAPI          bool     `json:"canAccessToAPI"`
	CanAccessToSSHKeys      bool     `json:"canAccessToSSHKeys"`
	CanAccessToGitProviders bool     `json:"canAccessToGitProviders"`
	CanDeleteEnvironments   bool     `json:"canDeleteEnvironments"`
	CanCreateEnvironments   bool     `json:"canCreateEnvironments"`
}

// GetUsers fetches all users
func (c *Client) GetUsers(ctx context.Context) ([]User, error) {
	data, err := c.doRequest(ctx, "/user.all")
	if err != nil {
		return nil, fmt.Errorf("fetching users: %w", err)
	}
	var users []User
	if err := json.Unmarshal(data, &users); err != nil {
		return nil, fmt.Errorf("parsing users: %w", err)
	}
	return users, nil
}

// GetUser fetches a specific user by ID
func (c *Client) GetUser(ctx context.Context, userID string) (*User, error) {
	users, err := c.GetUsers(ctx)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.ID == userID {
			return &user, nil
		}
	}
	return nil, fmt.Errorf("user not found: %s", userID)
}

// UpdateUser updates an existing user
func (c *Client) UpdateUser(ctx context.Context, req UpdateUserRequest) error {
	_, err := c.doPostRequest(ctx, "/user.update", req)
	if err != nil {
		return fmt.Errorf("updating user: %w", err)
	}
	return nil
}

// DeleteUser deletes a user
func (c *Client) DeleteUser(ctx context.Context, userID string) error {
	req := struct {
		UserID string `json:"userId"`
	}{UserID: userID}
	_, err := c.doPostRequest(ctx, "/user.remove", req)
	if err != nil {
		return fmt.Errorf("deleting user: %w", err)
	}
	return nil
}

// AssignUserPermissions assigns permissions to a user
func (c *Client) AssignUserPermissions(ctx context.Context, req UserPermissionsRequest) error {
	_, err := c.doPostRequest(ctx, "/user.assignPermissions", req)
	if err != nil {
		return fmt.Errorf("assigning user permissions: %w", err)
	}
	return nil
}
