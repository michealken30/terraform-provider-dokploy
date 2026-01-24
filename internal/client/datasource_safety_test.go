// Package client provides datasource safety tests that verify data source
// read operations do not trigger any write behaviors.
//
// These tests verify that all client methods used by data sources are
// read-only and never call create, update, delete, or other mutating endpoints.
package client

import (
	"context"
	"strings"
	"testing"
)

// =============================================================================
// Datasource Safety Tests - List Operations
// =============================================================================

// TestDatasourceSafety_GetProjects verifies GetProjects only makes read operations
func TestDatasourceSafety_GetProjects(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetProjects(ctx)
	if err != nil {
		t.Fatalf("GetProjects failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/project.all") {
		t.Fatalf("Expected /project.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetServers verifies GetServers only makes read operations
func TestDatasourceSafety_GetServers(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetServers(ctx)
	if err != nil {
		t.Fatalf("GetServers failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/server.all") {
		t.Fatalf("Expected /server.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetSSHKeys verifies GetSSHKeys only makes read operations
func TestDatasourceSafety_GetSSHKeys(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetSSHKeys(ctx)
	if err != nil {
		t.Fatalf("GetSSHKeys failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/sshKey.all") {
		t.Fatalf("Expected /sshKey.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetRegistries verifies GetRegistries only makes read operations
func TestDatasourceSafety_GetRegistries(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetRegistries(ctx)
	if err != nil {
		t.Fatalf("GetRegistries failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/registry.all") {
		t.Fatalf("Expected /registry.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetCertificates verifies GetCertificates only makes read operations
func TestDatasourceSafety_GetCertificates(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetCertificates(ctx)
	if err != nil {
		t.Fatalf("GetCertificates failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/certificates.all") {
		t.Fatalf("Expected /certificates.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetDestinations verifies GetDestinations only makes read operations
func TestDatasourceSafety_GetDestinations(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetDestinations(ctx)
	if err != nil {
		t.Fatalf("GetDestinations failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/destination.all") {
		t.Fatalf("Expected /destination.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetNotifications verifies GetNotifications only makes read operations
func TestDatasourceSafety_GetNotifications(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetNotifications(ctx)
	if err != nil {
		t.Fatalf("GetNotifications failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/notification.all") {
		t.Fatalf("Expected /notification.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetUsers verifies GetUsers only makes read operations
func TestDatasourceSafety_GetUsers(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetUsers(ctx)
	if err != nil {
		t.Fatalf("GetUsers failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/user.all") {
		t.Fatalf("Expected /user.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetOrganizations verifies GetOrganizations only makes read operations
func TestDatasourceSafety_GetOrganizations(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetOrganizations(ctx)
	if err != nil {
		t.Fatalf("GetOrganizations failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/organization.all") {
		t.Fatalf("Expected /organization.all, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetGitProviders verifies GetGitProviders only makes read operations
func TestDatasourceSafety_GetGitProviders(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetGitProviders(ctx)
	if err != nil {
		t.Fatalf("GetGitProviders failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/gitProvider.getAll") {
		t.Fatalf("Expected /gitProvider.getAll, got %s", requests[0].Endpoint)
	}
}

// TestDatasourceSafety_GetEnvironmentsByProjectID verifies GetEnvironmentsByProjectID only makes read operations
func TestDatasourceSafety_GetEnvironmentsByProjectID(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetEnvironmentsByProjectID(ctx, "proj-test-123")
	if err != nil {
		t.Fatalf("GetEnvironmentsByProjectID failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)

	// GetEnvironmentsByProjectID calls GetProject which calls GetProjects
	requests := sts.GetRequests()
	if len(requests) != 1 {
		t.Fatalf("Expected 1 request, got %d", len(requests))
	}
	if !strings.HasSuffix(requests[0].Endpoint, "/project.all") {
		t.Fatalf("Expected /project.all, got %s", requests[0].Endpoint)
	}
}

// =============================================================================
// Datasource Safety Tests - Single Resource by ID
// =============================================================================

// TestDatasourceSafety_GetProjectByName verifies GetProjectByName only makes read operations
func TestDatasourceSafety_GetProjectByName(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetProjectByName(ctx, "test-project")
	if err != nil {
		t.Fatalf("GetProjectByName failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestDatasourceSafety_GetServerByName verifies GetServerByName only makes read operations
func TestDatasourceSafety_GetServerByName(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetServerByName(ctx, "test-server")
	if err != nil {
		t.Fatalf("GetServerByName failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestDatasourceSafety_GetSSHKeyByName verifies GetSSHKeyByName only makes read operations
func TestDatasourceSafety_GetSSHKeyByName(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetSSHKeyByName(ctx, "test-sshkey")
	if err != nil {
		t.Fatalf("GetSSHKeyByName failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestDatasourceSafety_GetGithub verifies GetGithub only makes read operations
func TestDatasourceSafety_GetGithub(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	// Add handler for github.one endpoint
	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetGithub(ctx, "github-test-123")
	if err != nil {
		t.Fatalf("GetGithub failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestDatasourceSafety_GetDeployments_Application verifies GetDeployments for applications only makes read operations
func TestDatasourceSafety_GetDeployments_Application(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetDeployments(ctx, "app-test-123", "application")
	if err != nil {
		t.Fatalf("GetDeployments failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// TestDatasourceSafety_GetDeployments_Compose verifies GetDeployments for compose only makes read operations
func TestDatasourceSafety_GetDeployments_Compose(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	_, err := client.GetDeployments(ctx, "compose-test-123", "compose")
	if err != nil {
		t.Fatalf("GetDeployments failed: %v", err)
	}

	sts.AssertNoWriteOperations(t)
}

// =============================================================================
// Comprehensive Datasource Safety Test
// =============================================================================

// TestDatasourceSafety_AllDatasources runs all datasource read operations and verifies
// no write operations occur for any datasource type.
func TestDatasourceSafety_AllDatasources(t *testing.T) {
	sts := newSafetyTestServer(t)
	defer sts.Close()

	client := sts.Client()
	ctx := context.Background()

	// List operations (used by plural datasources)
	listTests := []struct {
		name string
		fn   func() error
	}{
		{"GetProjects", func() error { _, err := client.GetProjects(ctx); return err }},
		{"GetServers", func() error { _, err := client.GetServers(ctx); return err }},
		{"GetSSHKeys", func() error { _, err := client.GetSSHKeys(ctx); return err }},
		{"GetRegistries", func() error { _, err := client.GetRegistries(ctx); return err }},
		{"GetCertificates", func() error { _, err := client.GetCertificates(ctx); return err }},
		{"GetDestinations", func() error { _, err := client.GetDestinations(ctx); return err }},
		{"GetNotifications", func() error { _, err := client.GetNotifications(ctx); return err }},
		{"GetUsers", func() error { _, err := client.GetUsers(ctx); return err }},
		{"GetOrganizations", func() error { _, err := client.GetOrganizations(ctx); return err }},
		{"GetGitProviders", func() error { _, err := client.GetGitProviders(ctx); return err }},
	}

	// Single resource operations (used by singular datasources)
	singleTests := []struct {
		name string
		fn   func() error
	}{
		{"GetProject", func() error { _, err := client.GetProject(ctx, "proj-test-123"); return err }},
		{"GetProjectByName", func() error { _, err := client.GetProjectByName(ctx, "test-project"); return err }},
		{"GetServer", func() error { _, err := client.GetServer(ctx, "server-test-123"); return err }},
		{"GetServerByName", func() error { _, err := client.GetServerByName(ctx, "test-server"); return err }},
		{"GetSSHKey", func() error { _, err := client.GetSSHKey(ctx, "sshkey-test-123"); return err }},
		{"GetSSHKeyByName", func() error { _, err := client.GetSSHKeyByName(ctx, "test-sshkey"); return err }},
		{"GetApplication", func() error { _, err := client.GetApplication(ctx, "app-test-123"); return err }},
		{"GetUser", func() error { _, err := client.GetUser(ctx, "user-test-123"); return err }},
		{"GetOrganization", func() error { _, err := client.GetOrganization(ctx, "org-test-123"); return err }},
		{"GetNotification", func() error { _, err := client.GetNotification(ctx, "notif-test-123"); return err }},
		{"GetGithub", func() error { _, err := client.GetGithub(ctx, "github-test-123"); return err }},
		{"GetEnvironmentsByProjectID", func() error {
			_, err := client.GetEnvironmentsByProjectID(ctx, "proj-test-123")
			return err
		}},
		{"GetDeployments_Application", func() error {
			_, err := client.GetDeployments(ctx, "app-test-123", "application")
			return err
		}},
		{"GetDeployments_Compose", func() error {
			_, err := client.GetDeployments(ctx, "compose-test-123", "compose")
			return err
		}},
	}

	t.Run("ListOperations", func(t *testing.T) {
		for _, tt := range listTests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err != nil {
					t.Fatalf("%s failed: %v", tt.name, err)
				}
			})
		}
	})

	t.Run("SingleResourceOperations", func(t *testing.T) {
		for _, tt := range singleTests {
			t.Run(tt.name, func(t *testing.T) {
				err := tt.fn()
				if err != nil {
					t.Fatalf("%s failed: %v", tt.name, err)
				}
			})
		}
	})

	// Final verification: no write operations across all calls
	sts.AssertNoWriteOperations(t)

	// Report all endpoints called
	requests := sts.GetRequests()
	t.Logf("Total API calls during datasource operations: %d", len(requests))
	for _, req := range requests {
		t.Logf("  %s %s", req.Method, req.Endpoint)
	}
}
