package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// =============================================================================
// Domain Tests
// =============================================================================

func TestCreateDomain_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/domain.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateDomainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Host != "example.com" {
			t.Errorf("expected host 'example.com', got %s", req.Host)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateDomainResponse{DomainID: "new-domain-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	appID := "app-123"
	result, err := client.CreateDomain(context.Background(), CreateDomainRequest{
		Host:            "example.com",
		HTTPS:           true,
		CertificateType: "letsencrypt",
		ApplicationID:   &appID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DomainID != "new-domain-id" {
		t.Errorf("expected domain ID new-domain-id, got %s", result.DomainID)
	}
}

func TestCreateDomain_WithComposeID(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateDomainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ComposeID == nil || *req.ComposeID != "compose-123" {
			t.Errorf("expected compose ID 'compose-123', got %v", req.ComposeID)
		}
		if req.ServiceName == nil || *req.ServiceName != "web" {
			t.Errorf("expected service name 'web', got %v", req.ServiceName)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateDomainResponse{DomainID: "domain-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	composeID := "compose-123"
	serviceName := "web"
	_, err := client.CreateDomain(context.Background(), CreateDomainRequest{
		Host:            "example.com",
		HTTPS:           true,
		CertificateType: "none",
		ComposeID:       &composeID,
		ServiceName:     &serviceName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetDomain_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/domain.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Domain{
			DomainID: "domain-123",
			Host:     "example.com",
			HTTPS:    true,
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetDomain(context.Background(), "domain-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DomainID != "domain-123" {
		t.Errorf("expected domain ID domain-123, got %s", result.DomainID)
	}
	if result.Host != "example.com" {
		t.Errorf("expected host example.com, got %s", result.Host)
	}
}

func TestGetDomain_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "domain not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetDomain(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateDomain_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/domain.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateDomainRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.DomainID != "domain-1" {
			t.Errorf("expected domain ID domain-1, got %s", req.DomainID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateDomain(context.Background(), UpdateDomainRequest{
		DomainID:        "domain-1",
		Host:            "new.example.com",
		HTTPS:           true,
		CertificateType: "letsencrypt",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteDomain_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/domain.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteDomain(context.Background(), "domain-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Port Tests
// =============================================================================

func TestCreatePort_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/port.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreatePortRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.PublishedPort != 8080 {
			t.Errorf("expected published port 8080, got %d", req.PublishedPort)
		}
		if req.TargetPort != 80 {
			t.Errorf("expected target port 80, got %d", req.TargetPort)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreatePortResponse{PortID: "new-port-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreatePort(context.Background(), CreatePortRequest{
		PublishedPort: 8080,
		TargetPort:    80,
		Protocol:      "tcp",
		PublishMode:   "host",
		ApplicationID: "app-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PortID != "new-port-id" {
		t.Errorf("expected port ID new-port-id, got %s", result.PortID)
	}
}

func TestGetPort_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/port.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Port{
			PortID:        "port-123",
			PublishedPort: 8080,
			TargetPort:    80,
			Protocol:      "tcp",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetPort(context.Background(), "port-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PortID != "port-123" {
		t.Errorf("expected port ID port-123, got %s", result.PortID)
	}
}

func TestUpdatePort_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/port.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdatePort(context.Background(), UpdatePortRequest{
		PortID:        "port-1",
		PublishedPort: 9090,
		TargetPort:    80,
		Protocol:      "tcp",
		PublishMode:   "host",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeletePort_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/port.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeletePort(context.Background(), "port-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Mount Tests
// =============================================================================

func TestCreateMount_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mounts.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateMountRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Type != "volume" {
			t.Errorf("expected type 'volume', got %s", req.Type)
		}
		if req.MountPath != "/data" {
			t.Errorf("expected mount path '/data', got %s", req.MountPath)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateMountResponse{MountID: "new-mount-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	volumeName := "my-volume"
	result, err := client.CreateMount(context.Background(), CreateMountRequest{
		Type:        "volume",
		VolumeName:  &volumeName,
		MountPath:   "/data",
		ServiceType: "application",
		ServiceID:   "app-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MountID != "new-mount-id" {
		t.Errorf("expected mount ID new-mount-id, got %s", result.MountID)
	}
}

func TestGetMount_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mounts.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Mount{
			MountID:   "mount-123",
			Type:      "volume",
			MountPath: "/data",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetMount(context.Background(), "mount-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MountID != "mount-123" {
		t.Errorf("expected mount ID mount-123, got %s", result.MountID)
	}
}

func TestUpdateMount_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mounts.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newPath := "/new-data"
	err := client.UpdateMount(context.Background(), UpdateMountRequest{
		MountID:   "mount-1",
		MountPath: &newPath,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMount_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mounts.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteMount(context.Background(), "mount-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Security Tests
// =============================================================================

func TestCreateSecurity_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateSecurityRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Username != "admin" {
			t.Errorf("expected username 'admin', got %s", req.Username)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateSecurityResponse{SecurityID: "new-security-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateSecurity(context.Background(), CreateSecurityRequest{
		ApplicationID: "app-123",
		Username:      "admin",
		Password:      "secret",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SecurityID != "new-security-id" {
		t.Errorf("expected security ID new-security-id, got %s", result.SecurityID)
	}
}

func TestGetSecurity_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Security{
			SecurityID: "security-123",
			Username:   "admin",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetSecurity(context.Background(), "security-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SecurityID != "security-123" {
		t.Errorf("expected security ID security-123, got %s", result.SecurityID)
	}
}

func TestUpdateSecurity_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateSecurity(context.Background(), UpdateSecurityRequest{
		SecurityID: "security-1",
		Username:   "newadmin",
		Password:   "newsecret",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSecurity_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/security.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteSecurity(context.Background(), "security-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Redirect Tests
// =============================================================================

func TestCreateRedirect_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/redirects.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateRedirectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Regex != "^/old/(.*)" {
			t.Errorf("expected regex '^/old/(.*)', got %s", req.Regex)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateRedirectResponse{RedirectID: "new-redirect-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateRedirect(context.Background(), CreateRedirectRequest{
		Regex:         "^/old/(.*)",
		Replacement:   "/new/$1",
		Permanent:     true,
		ApplicationID: "app-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RedirectID != "new-redirect-id" {
		t.Errorf("expected redirect ID new-redirect-id, got %s", result.RedirectID)
	}
}

func TestGetRedirect_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/redirects.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Redirect{
			RedirectID:  "redirect-123",
			Regex:       "^/old/(.*)",
			Replacement: "/new/$1",
			Permanent:   true,
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetRedirect(context.Background(), "redirect-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RedirectID != "redirect-123" {
		t.Errorf("expected redirect ID redirect-123, got %s", result.RedirectID)
	}
}

func TestUpdateRedirect_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/redirects.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateRedirect(context.Background(), UpdateRedirectRequest{
		RedirectID:  "redirect-1",
		Regex:       "^/updated/(.*)",
		Replacement: "/new/$1",
		Permanent:   false,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRedirect_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/redirects.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteRedirect(context.Background(), "redirect-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Error Cases
// =============================================================================

func TestCreateDomain_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid domain"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateDomain(context.Background(), CreateDomainRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreatePort_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "port already in use"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreatePort(context.Background(), CreatePortRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateMount_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid mount path"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateMount(context.Background(), CreateMountRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSecurity_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "username required"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateSecurity(context.Background(), CreateSecurityRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRedirect_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid regex"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateRedirect(context.Background(), CreateRedirectRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
