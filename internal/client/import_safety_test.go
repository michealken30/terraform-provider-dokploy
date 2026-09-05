// Package client provides import safety tests that verify import operations
// do not trigger any write behaviors (create, update, delete).
//
// These tests create a mock HTTP server that:
// - Accepts read operations and returns valid mock data
// - Fails immediately if any write operation is attempted
// - Records all requests for verification
//
// This provides verifiable proof that exercising import codepaths is safe
// from any potentially destructive behavior.
package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

// WriteOperationError is returned when a write operation is detected
type WriteOperationError struct {
	Method   string
	Endpoint string
}

func (e WriteOperationError) Error() string {
	return "IMPORT SAFETY VIOLATION: write operation detected: " + e.Method + " " + e.Endpoint
}

// requestRecord tracks a single HTTP request
type requestRecord struct {
	Method   string
	Endpoint string
	Body     string
}

// safetyTestServer creates a mock server for import safety testing
type safetyTestServer struct {
	server        *httptest.Server
	requests      []requestRecord
	mu            sync.Mutex
	writeDetected *WriteOperationError
}

// writeEndpoints are patterns that indicate destructive operations
var writeEndpoints = []string{
	".create",
	".update",
	".delete",
	".remove",
	".deploy",
	".stop",
	".start",
	".redeploy",
	".reload",
	".restart",
	".refreshToken",
	".saveEnvironment",
	".saveBuildType",
	".saveGitProvider",
	".saveDockerProvider",
	".assignPermissions",
}

// isWriteEndpoint checks if an endpoint is a write operation
func isWriteEndpoint(endpoint string) bool {
	for _, pattern := range writeEndpoints {
		if strings.Contains(endpoint, pattern) {
			return true
		}
	}
	return false
}

// newSafetyTestServer creates a new mock server for import safety testing
func newSafetyTestServer(t *testing.T) *safetyTestServer {
	t.Helper()

	sts := &safetyTestServer{
		requests: make([]requestRecord, 0),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		sts.mu.Lock()
		defer sts.mu.Unlock()

		endpoint := r.URL.Path
		if r.URL.RawQuery != "" {
			endpoint += "?" + r.URL.RawQuery
		}

		// Read body for POST requests
		var body string
		if r.Method == http.MethodPost || r.Method == http.MethodDelete {
			buf := make([]byte, 1024)
			n, _ := r.Body.Read(buf)
			body = string(buf[:n])
		}

		// Record request
		sts.requests = append(sts.requests, requestRecord{
			Method:   r.Method,
			Endpoint: endpoint,
			Body:     body,
		})

		// Check for write operations - FAIL IMMEDIATELY
		if isWriteEndpoint(endpoint) {
			sts.writeDetected = &WriteOperationError{
				Method:   r.Method,
				Endpoint: endpoint,
			}
			http.Error(w, "WRITE OPERATION DETECTED - TEST FAILED", http.StatusForbidden)
			return
		}

		// Return mock data for read operations
		w.Header().Set("Content-Type", "application/json")

		switch {
		// List endpoints - return arrays with mock data
		case strings.HasSuffix(r.URL.Path, "/project.all"):
			json.NewEncoder(w).Encode([]Project{mockProject()})

		case strings.HasSuffix(r.URL.Path, "/server.all"):
			json.NewEncoder(w).Encode([]Server{mockServer()})

		case strings.HasSuffix(r.URL.Path, "/sshKey.all"):
			json.NewEncoder(w).Encode([]SSHKey{mockSSHKey()})

		case strings.HasSuffix(r.URL.Path, "/registry.all"):
			json.NewEncoder(w).Encode([]Registry{mockRegistry()})

		case strings.HasSuffix(r.URL.Path, "/certificates.all"):
			json.NewEncoder(w).Encode([]Certificate{mockCertificate()})

		case strings.HasSuffix(r.URL.Path, "/destination.all"):
			json.NewEncoder(w).Encode([]Destination{mockDestination()})

		case strings.HasSuffix(r.URL.Path, "/notification.all"):
			json.NewEncoder(w).Encode([]Notification{mockNotification()})

		case strings.HasSuffix(r.URL.Path, "/user.all"):
			json.NewEncoder(w).Encode([]User{mockUser()})

		case strings.HasSuffix(r.URL.Path, "/organization.all"):
			json.NewEncoder(w).Encode([]Organization{mockOrganization()})

		case strings.HasSuffix(r.URL.Path, "/gitProvider.getAll"):
			json.NewEncoder(w).Encode([]GitProvider{mockGitProvider()})

		// Single resource endpoints (GET with query param)
		case strings.HasSuffix(r.URL.Path, "/organization.one"):
			json.NewEncoder(w).Encode(mockOrganization())

		case strings.HasSuffix(r.URL.Path, "/gitlab.one"):
			json.NewEncoder(w).Encode(mockGitlabResponse())

		case strings.HasSuffix(r.URL.Path, "/bitbucket.one"):
			json.NewEncoder(w).Encode(mockBitbucketResponse())

		case strings.HasSuffix(r.URL.Path, "/gitea.one"):
			json.NewEncoder(w).Encode(mockGiteaResponse())

		// Single resource endpoints (POST with body) - these are READ operations
		case strings.HasSuffix(r.URL.Path, "/domain.one"):
			json.NewEncoder(w).Encode(mockDomain())

		case strings.HasSuffix(r.URL.Path, "/port.one"):
			json.NewEncoder(w).Encode(mockPort())

		case strings.HasSuffix(r.URL.Path, "/mounts.one"):
			json.NewEncoder(w).Encode(mockMount())

		case strings.HasSuffix(r.URL.Path, "/security.one"):
			json.NewEncoder(w).Encode(mockSecurity())

		case strings.HasSuffix(r.URL.Path, "/redirects.one"):
			json.NewEncoder(w).Encode(mockRedirect())

		case strings.HasSuffix(r.URL.Path, "/backup.one"):
			json.NewEncoder(w).Encode(mockBackup())

		case strings.HasSuffix(r.URL.Path, "/schedule.one"):
			json.NewEncoder(w).Encode(mockSchedule())

		case strings.HasSuffix(r.URL.Path, "/volumeBackups.one"):
			json.NewEncoder(w).Encode(mockVolumeBackup())

		case strings.HasSuffix(r.URL.Path, "/github.one"):
			json.NewEncoder(w).Encode(mockGithub())

		case strings.HasSuffix(r.URL.Path, "/deployment.all"),
			strings.HasSuffix(r.URL.Path, "/deployment.allByCompose"):
			json.NewEncoder(w).Encode(mockDeployments())

		default:
			// Unknown endpoint - return empty JSON object
			w.Write([]byte("{}"))
		}
	})

	sts.server = httptest.NewServer(mux)
	return sts
}

func (sts *safetyTestServer) Close() {
	sts.server.Close()
}

func (sts *safetyTestServer) Client() *Client {
	return New(sts.server.URL, "test-api-key")
}

func (sts *safetyTestServer) AssertNoWriteOperations(t *testing.T) {
	t.Helper()
	sts.mu.Lock()
	defer sts.mu.Unlock()

	if sts.writeDetected != nil {
		t.Fatalf("Write operation detected: %s %s", sts.writeDetected.Method, sts.writeDetected.Endpoint)
	}

	// Double-check all recorded requests
	for _, req := range sts.requests {
		if isWriteEndpoint(req.Endpoint) {
			t.Fatalf("Write operation in request log: %s %s", req.Method, req.Endpoint)
		}
	}
}

func (sts *safetyTestServer) GetRequests() []requestRecord {
	sts.mu.Lock()
	defer sts.mu.Unlock()
	return append([]requestRecord{}, sts.requests...)
}

// =============================================================================
// Helper Functions
// =============================================================================

func ptrStr(s string) *string {
	return &s
}

// =============================================================================
// Mock Data Factories
// =============================================================================

func mockProject() Project {
	return Project{
		ProjectID:   "proj-test-123",
		Name:        "test-project",
		Description: "Test project",
		Environments: []Environment{
			mockEnvironment(),
		},
	}
}

func mockEnvironment() Environment {
	return Environment{
		EnvironmentID: "env-test-123",
		Name:          "test-env",
		Description:   "Test environment",
		Applications:  []Application{mockApplication()},
		Compose:       []Compose{mockCompose()},
		Postgres:      []Postgres{mockPostgres()},
		MySQL:         []MySQL{mockMySQL()},
		MariaDB:       []MariaDB{mockMariaDB()},
		Mongo:         []Mongo{mockMongo()},
		Redis:         []Redis{mockRedis()},
	}
}

func mockApplication() Application {
	return Application{
		ApplicationID: "app-test-123",
		Name:          "test-app",
		AppName:       "test-app-name",
		Description:   "Test application",
		EnvironmentID: "env-test-123",
	}
}

func mockCompose() Compose {
	return Compose{
		ComposeID:     "compose-test-123",
		Name:          "test-compose",
		AppName:       "test-compose-name",
		Description:   "Test compose",
		EnvironmentID: "env-test-123",
		ComposeType:   "docker-compose",
	}
}

func mockPostgres() Postgres {
	return Postgres{
		PostgresID:       "pg-test-123",
		Name:             "test-postgres",
		AppName:          "test-postgres-app",
		Description:      "Test postgres",
		EnvironmentID:    "env-test-123",
		DatabaseName:     "testdb",
		DatabaseUser:     "testuser",
		DatabasePassword: "testpass",
	}
}

func mockMySQL() MySQL {
	return MySQL{
		MySQLID:              "mysql-test-123",
		Name:                 "test-mysql",
		AppName:              "test-mysql-app",
		Description:          "Test mysql",
		EnvironmentID:        "env-test-123",
		DatabaseName:         "testdb",
		DatabaseUser:         "testuser",
		DatabasePassword:     "testpass",
		DatabaseRootPassword: "testrootpass",
	}
}

func mockMariaDB() MariaDB {
	return MariaDB{
		MariaDBID:            "mariadb-test-123",
		Name:                 "test-mariadb",
		AppName:              "test-mariadb-app",
		Description:          "Test mariadb",
		EnvironmentID:        "env-test-123",
		DatabaseName:         "testdb",
		DatabaseUser:         "testuser",
		DatabasePassword:     "testpass",
		DatabaseRootPassword: "testrootpass",
	}
}

func mockMongo() Mongo {
	return Mongo{
		MongoID:          "mongo-test-123",
		Name:             "test-mongo",
		AppName:          "test-mongo-app",
		Description:      "Test mongo",
		EnvironmentID:    "env-test-123",
		DatabaseUser:     "testuser",
		DatabasePassword: "testpass",
	}
}

func mockRedis() Redis {
	return Redis{
		RedisID:          "redis-test-123",
		Name:             "test-redis",
		AppName:          "test-redis-app",
		Description:      "Test redis",
		EnvironmentID:    "env-test-123",
		DatabasePassword: "testpass",
	}
}

func mockServer() Server {
	return Server{
		ServerID:    "server-test-123",
		Name:        "test-server",
		Description: "Test server",
		IPAddress:   "192.168.1.100",
		Port:        22,
		SSHKeyID:    "sshkey-test-123",
	}
}

func mockSSHKey() SSHKey {
	return SSHKey{
		SSHKeyID:    "sshkey-test-123",
		Name:        "test-sshkey",
		Description: "Test SSH key",
		PublicKey:   "ssh-rsa AAAA...",
	}
}

func mockRegistry() Registry {
	return Registry{
		RegistryID:   "registry-test-123",
		RegistryName: "test-registry",
		Username:     "testuser",
		RegistryURL:  "https://registry.example.com",
		RegistryType: "selfHosted",
	}
}

func mockCertificate() Certificate {
	return Certificate{
		CertificateID:   "cert-test-123",
		Name:            "test-cert",
		CertificateData: "-----BEGIN CERTIFICATE-----\nMIIC...",
		AutoRenew:       false,
	}
}

func mockDestination() Destination {
	return Destination{
		DestinationID: "dest-test-123",
		Name:          "test-destination",
	}
}

func mockNotification() Notification {
	return Notification{
		NotificationID:   "notif-test-123",
		Name:             "test-notification",
		NotificationType: "slack",
	}
}

func mockUser() User {
	return User{
		ID:    "user-test-123",
		Name:  "Test User",
		Email: "test@example.com",
	}
}

func mockOrganization() Organization {
	return Organization{
		OrganizationID: "org-test-123",
		Name:           "test-org",
	}
}

func mockGitProvider() GitProvider {
	return GitProvider{
		GitProviderID: "gitprov-test-123",
		Name:          "test-gitprovider",
		ProviderType:  "gitlab",
	}
}

func mockGitlabResponse() GitlabResponse {
	return GitlabResponse{
		GitlabID:      "gitlab-test-123",
		Name:          "test-gitlab",
		GitProviderID: "gitprov-test-123",
	}
}

func mockBitbucketResponse() BitbucketResponse {
	return BitbucketResponse{
		BitbucketID:   "bitbucket-test-123",
		GitProviderID: "gitprov-test-123",
	}
}

func mockGiteaResponse() GiteaResponse {
	return GiteaResponse{
		GiteaID:       "gitea-test-123",
		Name:          "test-gitea",
		GitProviderID: "gitprov-test-123",
	}
}

func mockDomain() Domain {
	return Domain{
		DomainID:      "domain-test-123",
		Host:          "example.com",
		ApplicationID: ptrStr("app-test-123"),
	}
}

func mockPort() Port {
	return Port{
		PortID:        "port-test-123",
		PublishedPort: 8080,
		TargetPort:    80,
		Protocol:      "tcp",
		ApplicationID: ptrStr("app-test-123"),
	}
}

func mockMount() Mount {
	return Mount{
		MountID:       "mount-test-123",
		Type:          "bind",
		HostPath:      ptrStr("/host/path"),
		MountPath:     "/container/path",
		ApplicationID: ptrStr("app-test-123"),
	}
}

func mockSecurity() Security {
	return Security{
		SecurityID:    "security-test-123",
		Username:      "testuser",
		Password:      "testpass",
		ApplicationID: "app-test-123",
	}
}

func mockRedirect() Redirect {
	return Redirect{
		RedirectID:    "redirect-test-123",
		Regex:         "^/old/(.*)",
		Replacement:   "/new/$1",
		Permanent:     false,
		ApplicationID: "app-test-123",
	}
}

func mockBackup() Backup {
	return Backup{
		BackupID:      "backup-test-123",
		Schedule:      "0 0 * * *",
		Enabled:       true,
		DestinationID: "dest-test-123",
		PostgresID:    ptrStr("pg-test-123"),
	}
}

func mockSchedule() Schedule {
	return Schedule{
		ScheduleID:     "schedule-test-123",
		CronExpression: "*/5 * * * *",
		Command:        "echo hello",
		ScheduleType:   "application",
		ApplicationID:  ptrStr("app-test-123"),
	}
}

func mockVolumeBackup() VolumeBackup {
	return VolumeBackup{
		VolumeBackupID: "volbackup-test-123",
		Name:           "test-volume-backup",
		VolumeName:     "test-volume",
		Prefix:         "backup-",
		CronExpression: "0 0 * * *",
		DestinationID:  "dest-test-123",
	}
}

func mockGithub() Github {
	return Github{
		GithubID:      "github-test-123",
		GitProviderID: "gitprov-test-123",
	}
}

func mockDeployments() []Deployment {
	title := "Test deployment"
	desc := "Test deployment description"
	appID := "app-test-123"
	createdAt := "2024-01-01T00:00:00Z"
	return []Deployment{
		{
			DeploymentID:  "deploy-test-123",
			Title:         &title,
			Status:        "done",
			Description:   &desc,
			ApplicationID: &appID,
			CreatedAt:     &createdAt,
		},
	}
}

// =============================================================================
// Import Safety Tests
// =============================================================================

// TestImportSafety_Project verifies GetProject only makes read operations
func TestImportSafety_Project(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetProject(ctx, "proj-test-123")
	if err != nil {
		t.Fatalf("GetProject failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	// Verify only /project.all was called (GetProject uses GetProjects internally)
	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/project.all") {
		t.Fatalf("Expected /project.all, got %s", requests[0].Endpoint)
	}
}

// TestImportSafety_Environment verifies GetEnvironment only makes read operations
func TestImportSafety_Environment(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetEnvironment(ctx, "env-test-123")
	if err != nil {
		t.Fatalf("GetEnvironment failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Application verifies GetApplication only makes read operations
func TestImportSafety_Application(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetApplication(ctx, "app-test-123")
	if err != nil {
		t.Fatalf("GetApplication failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Compose verifies GetCompose only makes read operations
func TestImportSafety_Compose(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetCompose(ctx, "compose-test-123")
	if err != nil {
		t.Fatalf("GetCompose failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Postgres verifies GetPostgres only makes read operations
func TestImportSafety_Postgres(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetPostgres(ctx, "pg-test-123")
	if err != nil {
		t.Fatalf("GetPostgres failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_MySQL verifies GetMysql only makes read operations
func TestImportSafety_MySQL(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetMysql(ctx, "mysql-test-123")
	if err != nil {
		t.Fatalf("GetMysql failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_MariaDB verifies GetMariadb only makes read operations
func TestImportSafety_MariaDB(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetMariadb(ctx, "mariadb-test-123")
	if err != nil {
		t.Fatalf("GetMariadb failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Mongo verifies GetMongo only makes read operations
func TestImportSafety_Mongo(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetMongo(ctx, "mongo-test-123")
	if err != nil {
		t.Fatalf("GetMongo failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Redis verifies GetRedis only makes read operations
func TestImportSafety_Redis(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetRedis(ctx, "redis-test-123")
	if err != nil {
		t.Fatalf("GetRedis failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Server verifies GetServer only makes read operations
func TestImportSafety_Server(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetServer(ctx, "server-test-123")
	if err != nil {
		t.Fatalf("GetServer failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_SSHKey verifies GetSSHKey only makes read operations
func TestImportSafety_SSHKey(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetSSHKey(ctx, "sshkey-test-123")
	if err != nil {
		t.Fatalf("GetSSHKey failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Registry verifies GetRegistry only makes read operations
func TestImportSafety_Registry(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetRegistry(ctx, "registry-test-123")
	if err != nil {
		t.Fatalf("GetRegistry failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Certificate verifies GetCertificate only makes read operations
func TestImportSafety_Certificate(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetCertificate(ctx, "cert-test-123")
	if err != nil {
		t.Fatalf("GetCertificate failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Destination verifies GetDestination only makes read operations
func TestImportSafety_Destination(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetDestination(ctx, "dest-test-123")
	if err != nil {
		t.Fatalf("GetDestination failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Notification verifies GetNotification only makes read operations
func TestImportSafety_Notification(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetNotification(ctx, "notif-test-123")
	if err != nil {
		t.Fatalf("GetNotification failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Organization verifies GetOrganization only makes read operations
func TestImportSafety_Organization(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetOrganization(ctx, "org-test-123")
	if err != nil {
		t.Fatalf("GetOrganization failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	// Verify the endpoint used
	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.Contains(requests[0].Endpoint, "/organization.one") {
		t.Fatalf("Expected /organization.one, got %s", requests[0].Endpoint)
	}
}

// TestImportSafety_Gitlab verifies GetGitlab only makes read operations
func TestImportSafety_Gitlab(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetGitlab(ctx, "gitlab-test-123")
	if err != nil {
		t.Fatalf("GetGitlab failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Bitbucket verifies GetBitbucket only makes read operations
func TestImportSafety_Bitbucket(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetBitbucket(ctx, "bitbucket-test-123")
	if err != nil {
		t.Fatalf("GetBitbucket failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Gitea verifies GetGitea only makes read operations
func TestImportSafety_Gitea(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetGitea(ctx, "gitea-test-123")
	if err != nil {
		t.Fatalf("GetGitea failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Domain verifies GetDomain only makes read operations
func TestImportSafety_Domain(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetDomain(ctx, "domain-test-123")
	if err != nil {
		t.Fatalf("GetDomain failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	// Verify domain.one uses the documented read-only GET contract.
	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if requests[0].Method != http.MethodGet {
		t.Fatalf("Expected GET method, got %s", requests[0].Method)
	}
	if requests[0].Endpoint != "/api/domain.one?domainId=domain-test-123" {
		t.Fatalf("Expected domain.one query endpoint, got %s", requests[0].Endpoint)
	}
}

// TestImportSafety_Port verifies GetPort only makes read operations
func TestImportSafety_Port(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetPort(ctx, "port-test-123")
	if err != nil {
		t.Fatalf("GetPort failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Mount verifies GetMount only makes read operations
func TestImportSafety_Mount(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetMount(ctx, "mount-test-123")
	if err != nil {
		t.Fatalf("GetMount failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Security verifies GetSecurity only makes read operations
func TestImportSafety_Security(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetSecurity(ctx, "security-test-123")
	if err != nil {
		t.Fatalf("GetSecurity failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Redirect verifies GetRedirect only makes read operations
func TestImportSafety_Redirect(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetRedirect(ctx, "redirect-test-123")
	if err != nil {
		t.Fatalf("GetRedirect failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Backup verifies GetBackup only makes read operations
func TestImportSafety_Backup(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetBackup(ctx, "backup-test-123")
	if err != nil {
		t.Fatalf("GetBackup failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_Schedule verifies GetSchedule only makes read operations
func TestImportSafety_Schedule(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetSchedule(ctx, "schedule-test-123")
	if err != nil {
		t.Fatalf("GetSchedule failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_User verifies GetUser only makes read operations
func TestImportSafety_User(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetUser(ctx, "user-test-123")
	if err != nil {
		t.Fatalf("GetUser failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestImportSafety_VolumeBackup verifies GetVolumeBackup only makes read operations
func TestImportSafety_VolumeBackup(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetVolumeBackup(ctx, "volbackup-test-123")
	if err != nil {
		t.Fatalf("GetVolumeBackup failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// =============================================================================
// Comprehensive Import Safety Test
// =============================================================================

// TestImportSafety_AllResources runs all import codepaths and verifies
// no write operations occur for any resource type.
func TestImportSafety_AllResources(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	// Exercise all Get* methods used during import
	tests := []struct {
		name string
		fn   func() error
	}{
		{"Project", func() error { _, err := client.GetProject(ctx, "proj-test-123"); return err }},
		{"Environment", func() error { _, err := client.GetEnvironment(ctx, "env-test-123"); return err }},
		{"Application", func() error { _, err := client.GetApplication(ctx, "app-test-123"); return err }},
		{"Compose", func() error { _, err := client.GetCompose(ctx, "compose-test-123"); return err }},
		{"Postgres", func() error { _, err := client.GetPostgres(ctx, "pg-test-123"); return err }},
		{"MySQL", func() error { _, err := client.GetMysql(ctx, "mysql-test-123"); return err }},
		{"MariaDB", func() error { _, err := client.GetMariadb(ctx, "mariadb-test-123"); return err }},
		{"Mongo", func() error { _, err := client.GetMongo(ctx, "mongo-test-123"); return err }},
		{"Redis", func() error { _, err := client.GetRedis(ctx, "redis-test-123"); return err }},
		{"Server", func() error { _, err := client.GetServer(ctx, "server-test-123"); return err }},
		{"SSHKey", func() error { _, err := client.GetSSHKey(ctx, "sshkey-test-123"); return err }},
		{"Registry", func() error { _, err := client.GetRegistry(ctx, "registry-test-123"); return err }},
		{"Certificate", func() error { _, err := client.GetCertificate(ctx, "cert-test-123"); return err }},
		{"Destination", func() error { _, err := client.GetDestination(ctx, "dest-test-123"); return err }},
		{"Notification", func() error { _, err := client.GetNotification(ctx, "notif-test-123"); return err }},
		{"Organization", func() error { _, err := client.GetOrganization(ctx, "org-test-123"); return err }},
		{"Gitlab", func() error { _, err := client.GetGitlab(ctx, "gitlab-test-123"); return err }},
		{"Bitbucket", func() error { _, err := client.GetBitbucket(ctx, "bitbucket-test-123"); return err }},
		{"Gitea", func() error { _, err := client.GetGitea(ctx, "gitea-test-123"); return err }},
		{"Domain", func() error { _, err := client.GetDomain(ctx, "domain-test-123"); return err }},
		{"Port", func() error { _, err := client.GetPort(ctx, "port-test-123"); return err }},
		{"Mount", func() error { _, err := client.GetMount(ctx, "mount-test-123"); return err }},
		{"Security", func() error { _, err := client.GetSecurity(ctx, "security-test-123"); return err }},
		{"Redirect", func() error { _, err := client.GetRedirect(ctx, "redirect-test-123"); return err }},
		{"Backup", func() error { _, err := client.GetBackup(ctx, "backup-test-123"); return err }},
		{"Schedule", func() error { _, err := client.GetSchedule(ctx, "schedule-test-123"); return err }},
		{"User", func() error { _, err := client.GetUser(ctx, "user-test-123"); return err }},
		{"VolumeBackup", func() error { _, err := client.GetVolumeBackup(ctx, "volbackup-test-123"); return err }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.fn()
			if err != nil {
				t.Fatalf("Get%s failed: %v", tt.name, err)
			}
		})
	}

	// Final verification: no write operations across all calls
	sts.AssertNoWriteOperations(t)

	// Report all endpoints called
	requests := sts.GetRequests()
	t.Logf("Total API calls during import operations: %d", len(requests))
	for _, req := range requests {
		t.Logf("  %s %s", req.Method, req.Endpoint)
	}
}

// TestImportSafety_WriteOperationDetection verifies the safety mechanism works
// by attempting a write operation and confirming it's detected
func TestImportSafety_WriteOperationDetection(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	// Attempt a create operation - should fail
	_, err := client.CreateProject(ctx, CreateProjectRequest{Name: "test"})
	if err == nil {
		t.Fatal("Expected CreateProject to fail on safety test server")
	}

	// Verify the write was detected
	if sts.writeDetected == nil {
		t.Fatal("Write operation was not detected")
	}
	if !strings.Contains(sts.writeDetected.Endpoint, ".create") {
		t.Fatalf("Expected .create endpoint, got %s", sts.writeDetected.Endpoint)
	}
}
