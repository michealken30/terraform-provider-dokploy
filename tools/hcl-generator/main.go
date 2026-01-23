package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// Models for reading backup JSON files (subset of fields needed for HCL generation)

type Project struct {
	ProjectID    string        `json:"projectId"`
	Name         string        `json:"name"`
	Description  string        `json:"description"`
	Environments []Environment `json:"environments"`
}

type Environment struct {
	EnvironmentID string        `json:"environmentId"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	ProjectID     string        `json:"projectId"`
	IsDefault     bool          `json:"isDefault"`
	Applications  []Application `json:"applications"`
	Compose       []Compose     `json:"compose"`
	Postgres      []Postgres    `json:"postgres"`
	MySQL         []MySQL       `json:"mysql"`
	MariaDB       []MariaDB     `json:"mariadb"`
	Mongo         []Mongo       `json:"mongo"`
	Redis         []Redis       `json:"redis"`
}

type Application struct {
	ApplicationID     string   `json:"applicationId"`
	Name              string   `json:"name"`
	AppName           string   `json:"appName"`
	Description       string   `json:"description"`
	EnvironmentID     string   `json:"environmentId"`
	SourceType        string   `json:"sourceType"`
	Repository        string   `json:"repository"`
	Owner             string   `json:"owner"`
	Branch            string   `json:"branch"`
	BuildPath         string   `json:"buildPath"`
	AutoDeploy        bool     `json:"autoDeploy"`
	BuildType         string   `json:"buildType"`
	Dockerfile        string   `json:"dockerfile"`
	Replicas          int      `json:"replicas"`
	ServerID          *string  `json:"serverId"`
	RegistryID        *string  `json:"registryId"`
	MemoryLimit       *int64   `json:"memoryLimit"`
	MemoryReservation *int64   `json:"memoryReservation"`
	CPULimit          *int64   `json:"cpuLimit"`
	CPUReservation    *int64   `json:"cpuReservation"`
	Domains           []Domain `json:"domains"`
}

type Compose struct {
	ComposeID       string  `json:"composeId"`
	Name            string  `json:"name"`
	AppName         string  `json:"appName"`
	Description     string  `json:"description"`
	EnvironmentID   string  `json:"environmentId"`
	ComposeFile     string  `json:"composeFile"`
	ComposeType     string  `json:"composeType"`
	SourceType      string  `json:"sourceType"`
	Repository      string  `json:"repository"`
	Owner           string  `json:"owner"`
	Branch          string  `json:"branch"`
	AutoDeploy      bool    `json:"autoDeploy"`
	ServerID        *string `json:"serverId"`
	RandomizeCompose bool   `json:"randomizeCompose"`
}

type Domain struct {
	DomainID        string `json:"domainId"`
	Host            string `json:"host"`
	Port            *int   `json:"port"`
	HTTPS           bool   `json:"https"`
	CertificateType string `json:"certificateType"`
	Path            string `json:"path"`
}

type Postgres struct {
	PostgresID       string  `json:"postgresId"`
	Name             string  `json:"name"`
	AppName          string  `json:"appName"`
	Description      string  `json:"description"`
	EnvironmentID    string  `json:"environmentId"`
	DatabaseName     string  `json:"databaseName"`
	DatabaseUser     string  `json:"databaseUser"`
	DockerImage      string  `json:"dockerImage"`
	ServerID         *string `json:"serverId"`
	ExternalPort     *int    `json:"externalPort"`
	MemoryLimit      *int64  `json:"memoryLimit"`
	MemoryReservation *int64 `json:"memoryReservation"`
}

type MySQL struct {
	MySQLID          string  `json:"mysqlId"`
	Name             string  `json:"name"`
	AppName          string  `json:"appName"`
	Description      string  `json:"description"`
	EnvironmentID    string  `json:"environmentId"`
	DatabaseName     string  `json:"databaseName"`
	DatabaseUser     string  `json:"databaseUser"`
	DockerImage      string  `json:"dockerImage"`
	ServerID         *string `json:"serverId"`
	ExternalPort     *int    `json:"externalPort"`
}

type MariaDB struct {
	MariaDBID        string  `json:"mariadbId"`
	Name             string  `json:"name"`
	AppName          string  `json:"appName"`
	Description      string  `json:"description"`
	EnvironmentID    string  `json:"environmentId"`
	DatabaseName     string  `json:"databaseName"`
	DatabaseUser     string  `json:"databaseUser"`
	DockerImage      string  `json:"dockerImage"`
	ServerID         *string `json:"serverId"`
}

type Mongo struct {
	MongoID       string  `json:"mongoId"`
	Name          string  `json:"name"`
	AppName       string  `json:"appName"`
	Description   string  `json:"description"`
	EnvironmentID string  `json:"environmentId"`
	DatabaseUser  string  `json:"databaseUser"`
	DockerImage   string  `json:"dockerImage"`
	ServerID      *string `json:"serverId"`
}

type Redis struct {
	RedisID       string  `json:"redisId"`
	Name          string  `json:"name"`
	AppName       string  `json:"appName"`
	Description   string  `json:"description"`
	EnvironmentID string  `json:"environmentId"`
	DockerImage   string  `json:"dockerImage"`
	ServerID      *string `json:"serverId"`
}

type Server struct {
	ServerID            string `json:"serverId"`
	Name                string `json:"name"`
	Description         string `json:"description"`
	IPAddress           string `json:"ipAddress"`
	Port                int    `json:"port"`
	Username            string `json:"username"`
	SSHKeyID            string `json:"sshKeyId"`
	ServerType          string `json:"serverType"`
	EnableDockerCleanup bool   `json:"enableDockerCleanup"`
}

type SSHKey struct {
	SSHKeyID    string `json:"sshKeyId"`
	Name        string `json:"name"`
	Description string `json:"description"`
	PublicKey   string `json:"publicKey"`
}

type Registry struct {
	RegistryID   string  `json:"registryId"`
	RegistryName string  `json:"registryName"`
	Username     string  `json:"username"`
	RegistryURL  string  `json:"registryUrl"`
	SelfHosted   string  `json:"selfHosted"`
	ImagePrefix  *string `json:"imagePrefix"`
	RegistryType string  `json:"registryType"`
}

// HCL Generator

type Generator struct {
	backupDir string
	outputDir string
	imports   []string
	servers   map[string]Server
	sshKeys   map[string]SSHKey
	projects  map[string]Project
	usedNames map[string]int // Track used resource names to avoid collisions
}

func NewGenerator(backupDir, outputDir string) *Generator {
	return &Generator{
		backupDir: backupDir,
		outputDir: outputDir,
		imports:   []string{},
		servers:   make(map[string]Server),
		sshKeys:   make(map[string]SSHKey),
		projects:  make(map[string]Project),
		usedNames: make(map[string]int),
	}
}

// uniqueName returns a unique resource name, appending a counter if needed
func (g *Generator) uniqueName(resourceType, baseName string) string {
	key := resourceType + "." + baseName
	count := g.usedNames[key]
	g.usedNames[key] = count + 1
	if count == 0 {
		return baseName
	}
	return fmt.Sprintf("%s_%d", baseName, count+1)
}

func (g *Generator) loadJSON(filename string, v interface{}) error {
	data, err := os.ReadFile(filepath.Join(g.backupDir, filename))
	if err != nil {
		return fmt.Errorf("reading %s: %w", filename, err)
	}
	if err := json.Unmarshal(data, v); err != nil {
		return fmt.Errorf("parsing %s: %w", filename, err)
	}
	return nil
}

func sanitizeName(name string) string {
	// Convert to lowercase and replace non-alphanumeric with underscore
	re := regexp.MustCompile(`[^a-zA-Z0-9]+`)
	result := re.ReplaceAllString(name, "_")
	result = strings.Trim(result, "_")
	result = strings.ToLower(result)
	// Ensure it starts with a letter
	if len(result) > 0 && result[0] >= '0' && result[0] <= '9' {
		result = "n" + result
	}
	if result == "" {
		result = "unnamed"
	}
	return result
}

func escapeHCL(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "\n", "\\n")
	return s
}

func (g *Generator) Generate() error {
	// Create output directory
	if err := os.MkdirAll(g.outputDir, 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	// Load all backup data
	var sshKeys []SSHKey
	if err := g.loadJSON("ssh_keys.json", &sshKeys); err != nil {
		return err
	}
	for _, key := range sshKeys {
		g.sshKeys[key.SSHKeyID] = key
	}

	var servers []Server
	if err := g.loadJSON("servers.json", &servers); err != nil {
		return err
	}
	for _, server := range servers {
		g.servers[server.ServerID] = server
	}

	var projects []Project
	if err := g.loadJSON("projects.json", &projects); err != nil {
		return err
	}
	for _, project := range projects {
		g.projects[project.ProjectID] = project
	}

	var registries []Registry
	if err := g.loadJSON("registries.json", &registries); err != nil {
		fmt.Printf("Warning: could not load registries: %v\n", err)
		registries = []Registry{}
	}

	// Generate HCL files
	if err := g.generateProvider(); err != nil {
		return err
	}
	if err := g.generateSSHKeys(sshKeys); err != nil {
		return err
	}
	if err := g.generateServers(servers); err != nil {
		return err
	}
	if err := g.generateProjects(projects); err != nil {
		return err
	}
	if err := g.generateApplications(projects); err != nil {
		return err
	}
	if err := g.generateCompose(projects); err != nil {
		return err
	}
	if err := g.generateDatabases(projects); err != nil {
		return err
	}
	if err := g.generateRegistries(registries); err != nil {
		return err
	}
	if err := g.generateImports(); err != nil {
		return err
	}

	return nil
}

func (g *Generator) generateProvider() error {
	content := `# Dokploy Provider Configuration
# Generated from backup on ` + time.Now().Format("2006-01-02") + `

terraform {
  required_providers {
    dokploy = {
      source  = "registry.terraform.io/reserve-protocol/dokploy"
      version = "~> 0.1"
    }
  }
}

provider "dokploy" {
  host    = var.dokploy_host
  api_key = var.dokploy_api_key
}

variable "dokploy_host" {
  description = "Dokploy server URL"
  type        = string
}

variable "dokploy_api_key" {
  description = "Dokploy API key"
  type        = string
  sensitive   = true
}
`
	return os.WriteFile(filepath.Join(g.outputDir, "provider.tf"), []byte(content), 0644)
}

func (g *Generator) generateSSHKeys(keys []SSHKey) error {
	if len(keys) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("# SSH Keys\n")
	sb.WriteString("# NOTE: Private keys are NOT included for security. Keys must be imported.\n\n")

	for _, key := range keys {
		resourceName := sanitizeName(key.Name)
		g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_ssh_key.%s
  id = "%s"
}`, resourceName, key.SSHKeyID))

		sb.WriteString(fmt.Sprintf(`# SSH Key: %s
resource "dokploy_ssh_key" "%s" {
  name        = "%s"
  description = "%s"
  # public_key = "%s"
  # NOTE: Private key must be provided during import or manual creation
}

`, key.Name, resourceName, escapeHCL(key.Name), escapeHCL(key.Description), escapeHCL(key.PublicKey)))
	}

	return os.WriteFile(filepath.Join(g.outputDir, "ssh_keys.tf"), []byte(sb.String()), 0644)
}

func (g *Generator) generateServers(servers []Server) error {
	if len(servers) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("# Servers\n\n")

	for _, server := range servers {
		resourceName := sanitizeName(server.Name)
		sshKeyRef := "# ssh_key_id unknown"
		if sshKey, ok := g.sshKeys[server.SSHKeyID]; ok {
			sshKeyRef = fmt.Sprintf("dokploy_ssh_key.%s.id", sanitizeName(sshKey.Name))
		}

		g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_server.%s
  id = "%s"
}`, resourceName, server.ServerID))

		sb.WriteString(fmt.Sprintf(`resource "dokploy_server" "%s" {
  name                  = "%s"
  description           = "%s"
  ip_address            = "%s"
  port                  = %d
  username              = "%s"
  ssh_key_id            = %s
  server_type           = "%s"
  enable_docker_cleanup = %t
}

`, resourceName, escapeHCL(server.Name), escapeHCL(server.Description),
			server.IPAddress, server.Port, server.Username, sshKeyRef,
			server.ServerType, server.EnableDockerCleanup))
	}

	return os.WriteFile(filepath.Join(g.outputDir, "servers.tf"), []byte(sb.String()), 0644)
}

func (g *Generator) generateProjects(projects []Project) error {
	if len(projects) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("# Projects\n\n")

	for _, project := range projects {
		resourceName := sanitizeName(project.Name)

		g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_project.%s
  id = "%s"
}`, resourceName, project.ProjectID))

		sb.WriteString(fmt.Sprintf(`resource "dokploy_project" "%s" {
  name        = "%s"
  description = "%s"
}

`, resourceName, escapeHCL(project.Name), escapeHCL(project.Description)))

		// Generate environments
		for _, env := range project.Environments {
			envResourceName := fmt.Sprintf("%s_%s", resourceName, sanitizeName(env.Name))

			g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_environment.%s
  id = "%s"
}`, envResourceName, env.EnvironmentID))

			sb.WriteString(fmt.Sprintf(`resource "dokploy_environment" "%s" {
  project_id  = dokploy_project.%s.id
  name        = "%s"
  description = "%s"
  is_default  = %t
}

`, envResourceName, resourceName, escapeHCL(env.Name), escapeHCL(env.Description), env.IsDefault))
		}
	}

	return os.WriteFile(filepath.Join(g.outputDir, "projects.tf"), []byte(sb.String()), 0644)
}

func (g *Generator) generateApplications(projects []Project) error {
	var sb strings.Builder
	sb.WriteString("# Applications\n\n")
	hasContent := false

	for _, project := range projects {
		projectName := sanitizeName(project.Name)
		for _, env := range project.Environments {
			envName := sanitizeName(env.Name)
			for _, app := range env.Applications {
				hasContent = true
				baseName := fmt.Sprintf("%s_%s_%s", projectName, envName, sanitizeName(app.Name))
				resourceName := g.uniqueName("dokploy_application", baseName)

				g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_application.%s
  id = "%s"
}`, resourceName, app.ApplicationID))

				serverRef := "null"
				if app.ServerID != nil {
					if server, ok := g.servers[*app.ServerID]; ok {
						serverRef = fmt.Sprintf("dokploy_server.%s.id", sanitizeName(server.Name))
					}
				}

				sb.WriteString(fmt.Sprintf(`resource "dokploy_application" "%s" {
  environment_id = dokploy_environment.%s_%s.id
  name           = "%s"
  description    = "%s"
  source_type    = "%s"
  repository     = "%s"
  owner          = "%s"
  branch         = "%s"
  build_path     = "%s"
  build_type     = "%s"
  dockerfile     = "%s"
  auto_deploy    = %t
  replicas       = %d
  server_id      = %s
}

`, resourceName, projectName, envName, escapeHCL(app.Name), escapeHCL(app.Description),
					app.SourceType, app.Repository, app.Owner, app.Branch,
					app.BuildPath, app.BuildType, app.Dockerfile, app.AutoDeploy,
					app.Replicas, serverRef))
			}
		}
	}

	if !hasContent {
		return nil
	}

	return os.WriteFile(filepath.Join(g.outputDir, "applications.tf"), []byte(sb.String()), 0644)
}

func (g *Generator) generateCompose(projects []Project) error {
	var sb strings.Builder
	sb.WriteString("# Compose Services\n\n")
	hasContent := false

	for _, project := range projects {
		projectName := sanitizeName(project.Name)
		for _, env := range project.Environments {
			envName := sanitizeName(env.Name)
			for _, comp := range env.Compose {
				hasContent = true
				baseName := fmt.Sprintf("%s_%s_%s", projectName, envName, sanitizeName(comp.Name))
				resourceName := g.uniqueName("dokploy_compose", baseName)

				g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_compose.%s
  id = "%s"
}`, resourceName, comp.ComposeID))

				serverRef := "null"
				if comp.ServerID != nil {
					if server, ok := g.servers[*comp.ServerID]; ok {
						serverRef = fmt.Sprintf("dokploy_server.%s.id", sanitizeName(server.Name))
					}
				}

				sb.WriteString(fmt.Sprintf(`resource "dokploy_compose" "%s" {
  environment_id     = dokploy_environment.%s_%s.id
  name               = "%s"
  description        = "%s"
  source_type        = "%s"
  repository         = "%s"
  owner              = "%s"
  branch             = "%s"
  compose_type       = "%s"
  auto_deploy        = %t
  randomize_compose  = %t
  server_id          = %s
  # compose_file content omitted - configure via source_type
}

`, resourceName, projectName, envName, escapeHCL(comp.Name), escapeHCL(comp.Description),
					comp.SourceType, comp.Repository, comp.Owner, comp.Branch,
					comp.ComposeType, comp.AutoDeploy, comp.RandomizeCompose, serverRef))
			}
		}
	}

	if !hasContent {
		return nil
	}

	return os.WriteFile(filepath.Join(g.outputDir, "compose.tf"), []byte(sb.String()), 0644)
}

func (g *Generator) generateDatabases(projects []Project) error {
	var sb strings.Builder
	sb.WriteString("# Databases\n")
	sb.WriteString("# NOTE: Database passwords are NOT included for security.\n\n")
	hasContent := false

	for _, project := range projects {
		projectName := sanitizeName(project.Name)
		for _, env := range project.Environments {
			envName := sanitizeName(env.Name)

			// Postgres
			for _, pg := range env.Postgres {
				hasContent = true
				baseName := fmt.Sprintf("%s_%s_%s", projectName, envName, sanitizeName(pg.Name))
				resourceName := g.uniqueName("dokploy_postgres", baseName)

				g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_postgres.%s
  id = "%s"
}`, resourceName, pg.PostgresID))

				serverRef := "null"
				if pg.ServerID != nil {
					if server, ok := g.servers[*pg.ServerID]; ok {
						serverRef = fmt.Sprintf("dokploy_server.%s.id", sanitizeName(server.Name))
					}
				}

				sb.WriteString(fmt.Sprintf(`resource "dokploy_postgres" "%s" {
  environment_id = dokploy_environment.%s_%s.id
  name           = "%s"
  description    = "%s"
  database_name  = "%s"
  database_user  = "%s"
  docker_image   = "%s"
  server_id      = %s
  # database_password must be provided separately
}

`, resourceName, projectName, envName, escapeHCL(pg.Name), escapeHCL(pg.Description),
					pg.DatabaseName, pg.DatabaseUser, pg.DockerImage, serverRef))
			}

			// MySQL
			for _, mysql := range env.MySQL {
				hasContent = true
				baseName := fmt.Sprintf("%s_%s_%s", projectName, envName, sanitizeName(mysql.Name))
				resourceName := g.uniqueName("dokploy_mysql", baseName)

				g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_mysql.%s
  id = "%s"
}`, resourceName, mysql.MySQLID))

				serverRef := "null"
				if mysql.ServerID != nil {
					if server, ok := g.servers[*mysql.ServerID]; ok {
						serverRef = fmt.Sprintf("dokploy_server.%s.id", sanitizeName(server.Name))
					}
				}

				sb.WriteString(fmt.Sprintf(`resource "dokploy_mysql" "%s" {
  environment_id = dokploy_environment.%s_%s.id
  name           = "%s"
  description    = "%s"
  database_name  = "%s"
  database_user  = "%s"
  docker_image   = "%s"
  server_id      = %s
  # database_password must be provided separately
}

`, resourceName, projectName, envName, escapeHCL(mysql.Name), escapeHCL(mysql.Description),
					mysql.DatabaseName, mysql.DatabaseUser, mysql.DockerImage, serverRef))
			}

			// MariaDB
			for _, mariadb := range env.MariaDB {
				hasContent = true
				baseName := fmt.Sprintf("%s_%s_%s", projectName, envName, sanitizeName(mariadb.Name))
				resourceName := g.uniqueName("dokploy_mariadb", baseName)

				g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_mariadb.%s
  id = "%s"
}`, resourceName, mariadb.MariaDBID))

				serverRef := "null"
				if mariadb.ServerID != nil {
					if server, ok := g.servers[*mariadb.ServerID]; ok {
						serverRef = fmt.Sprintf("dokploy_server.%s.id", sanitizeName(server.Name))
					}
				}

				sb.WriteString(fmt.Sprintf(`resource "dokploy_mariadb" "%s" {
  environment_id = dokploy_environment.%s_%s.id
  name           = "%s"
  description    = "%s"
  database_name  = "%s"
  database_user  = "%s"
  docker_image   = "%s"
  server_id      = %s
  # database_password must be provided separately
}

`, resourceName, projectName, envName, escapeHCL(mariadb.Name), escapeHCL(mariadb.Description),
					mariadb.DatabaseName, mariadb.DatabaseUser, mariadb.DockerImage, serverRef))
			}

			// Mongo
			for _, mongo := range env.Mongo {
				hasContent = true
				baseName := fmt.Sprintf("%s_%s_%s", projectName, envName, sanitizeName(mongo.Name))
				resourceName := g.uniqueName("dokploy_mongo", baseName)

				g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_mongo.%s
  id = "%s"
}`, resourceName, mongo.MongoID))

				serverRef := "null"
				if mongo.ServerID != nil {
					if server, ok := g.servers[*mongo.ServerID]; ok {
						serverRef = fmt.Sprintf("dokploy_server.%s.id", sanitizeName(server.Name))
					}
				}

				sb.WriteString(fmt.Sprintf(`resource "dokploy_mongo" "%s" {
  environment_id = dokploy_environment.%s_%s.id
  name           = "%s"
  description    = "%s"
  database_user  = "%s"
  docker_image   = "%s"
  server_id      = %s
  # database_password must be provided separately
}

`, resourceName, projectName, envName, escapeHCL(mongo.Name), escapeHCL(mongo.Description),
					mongo.DatabaseUser, mongo.DockerImage, serverRef))
			}

			// Redis
			for _, redis := range env.Redis {
				hasContent = true
				baseName := fmt.Sprintf("%s_%s_%s", projectName, envName, sanitizeName(redis.Name))
				resourceName := g.uniqueName("dokploy_redis", baseName)

				g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_redis.%s
  id = "%s"
}`, resourceName, redis.RedisID))

				serverRef := "null"
				if redis.ServerID != nil {
					if server, ok := g.servers[*redis.ServerID]; ok {
						serverRef = fmt.Sprintf("dokploy_server.%s.id", sanitizeName(server.Name))
					}
				}

				sb.WriteString(fmt.Sprintf(`resource "dokploy_redis" "%s" {
  environment_id = dokploy_environment.%s_%s.id
  name           = "%s"
  description    = "%s"
  docker_image   = "%s"
  server_id      = %s
  # database_password must be provided separately
}

`, resourceName, projectName, envName, escapeHCL(redis.Name), escapeHCL(redis.Description),
					redis.DockerImage, serverRef))
			}
		}
	}

	if !hasContent {
		return nil
	}

	return os.WriteFile(filepath.Join(g.outputDir, "databases.tf"), []byte(sb.String()), 0644)
}

func (g *Generator) generateRegistries(registries []Registry) error {
	if len(registries) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("# Container Registries\n")
	sb.WriteString("# NOTE: Passwords are NOT included for security.\n\n")

	for _, reg := range registries {
		resourceName := sanitizeName(reg.RegistryName)

		g.imports = append(g.imports, fmt.Sprintf(`import {
  to = dokploy_registry.%s
  id = "%s"
}`, resourceName, reg.RegistryID))

		imagePrefix := ""
		if reg.ImagePrefix != nil {
			imagePrefix = *reg.ImagePrefix
		}

		sb.WriteString(fmt.Sprintf(`resource "dokploy_registry" "%s" {
  registry_name = "%s"
  registry_url  = "%s"
  registry_type = "%s"
  username      = "%s"
  self_hosted   = "%s"
  image_prefix  = "%s"
  # password must be provided separately
}

`, resourceName, escapeHCL(reg.RegistryName), reg.RegistryURL, reg.RegistryType,
			reg.Username, reg.SelfHosted, imagePrefix))
	}

	return os.WriteFile(filepath.Join(g.outputDir, "registries.tf"), []byte(sb.String()), 0644)
}

func (g *Generator) generateImports() error {
	if len(g.imports) == 0 {
		return nil
	}

	var sb strings.Builder
	sb.WriteString("# Import blocks for existing resources\n")
	sb.WriteString("# Run 'terraform plan' to preview imports\n")
	sb.WriteString("# Run 'terraform apply' to import into state\n\n")

	for _, imp := range g.imports {
		sb.WriteString(imp)
		sb.WriteString("\n\n")
	}

	return os.WriteFile(filepath.Join(g.outputDir, "imports.tf"), []byte(sb.String()), 0644)
}

func main() {
	backupDir := flag.String("backup-dir", "", "Path to backup directory containing JSON files")
	outputDir := flag.String("output", "./generated", "Output directory for HCL files")
	flag.Parse()

	if *backupDir == "" {
		fmt.Println("Usage: hcl-generator --backup-dir <path> [--output <path>]")
		fmt.Println("")
		fmt.Println("Generate Terraform HCL configuration from Dokploy backup files.")
		fmt.Println("")
		fmt.Println("Options:")
		fmt.Println("  --backup-dir  Path to backup directory containing JSON files (required)")
		fmt.Println("  --output      Output directory for generated HCL files (default: ./generated)")
		os.Exit(1)
	}

	generator := NewGenerator(*backupDir, *outputDir)
	if err := generator.Generate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generated HCL files in %s\n", *outputDir)
	fmt.Println("")
	fmt.Println("Files created:")
	files, _ := os.ReadDir(*outputDir)
	for _, f := range files {
		fmt.Printf("  - %s\n", f.Name())
	}
	fmt.Println("")
	fmt.Println("Next steps:")
	fmt.Println("  1. Review generated files and adjust as needed")
	fmt.Println("  2. Set DOKPLOY_API_KEY environment variable")
	fmt.Println("  3. Run 'terraform init' to initialize")
	fmt.Println("  4. Run 'terraform plan' to preview imports")
	fmt.Println("  5. Run 'terraform apply' to import existing resources")
}
