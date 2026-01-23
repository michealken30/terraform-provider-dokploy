package client

import (
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
