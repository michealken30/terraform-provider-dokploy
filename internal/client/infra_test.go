package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// =============================================================================
// SSH Key Tests
// =============================================================================

func TestGetSSHKeys_Success(t *testing.T) {
	keys := []SSHKey{
		{SSHKeyID: "ssh-1", Name: "Key One"},
		{SSHKeyID: "ssh-2", Name: "Key Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sshKey.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(keys)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetSSHKeys(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 SSH keys, got %d", len(result))
	}
}

func TestGetSSHKey_Success(t *testing.T) {
	keys := []SSHKey{
		{SSHKeyID: "ssh-1", Name: "Key One"},
		{SSHKeyID: "ssh-2", Name: "Key Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(keys)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetSSHKey(context.Background(), "ssh-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SSHKeyID != "ssh-2" {
		t.Errorf("expected SSH key ID ssh-2, got %s", result.SSHKeyID)
	}
}

func TestGetSSHKey_NotFound(t *testing.T) {
	keys := []SSHKey{{SSHKeyID: "ssh-1", Name: "Key One"}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(keys)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetSSHKey(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetSSHKeyByName_Success(t *testing.T) {
	keys := []SSHKey{
		{SSHKeyID: "ssh-1", Name: "Production Key"},
		{SSHKeyID: "ssh-2", Name: "Staging Key"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(keys)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetSSHKeyByName(context.Background(), "Staging Key")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SSHKeyID != "ssh-2" {
		t.Errorf("expected SSH key ID ssh-2, got %s", result.SSHKeyID)
	}
}

func TestGetSSHKeyByName_NotFound(t *testing.T) {
	keys := []SSHKey{{SSHKeyID: "ssh-1", Name: "Key One"}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(keys)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetSSHKeyByName(context.Background(), "Non-Existent Key")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateSSHKey_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sshKey.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateSSHKeyRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Key" {
			t.Errorf("expected name 'New Key', got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateSSHKeyResponse{SSHKeyID: "new-ssh-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateSSHKey(context.Background(), CreateSSHKeyRequest{
		Name:           "New Key",
		PrivateKey:     "private-key-content",
		PublicKey:      "public-key-content",
		OrganizationID: "org-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SSHKeyID != "new-ssh-id" {
		t.Errorf("expected SSH key ID new-ssh-id, got %s", result.SSHKeyID)
	}
}

func TestUpdateSSHKey_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sshKey.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Key"
	err := client.UpdateSSHKey(context.Background(), UpdateSSHKeyRequest{
		SSHKeyID: "ssh-1",
		Name:     &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSSHKey_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/sshKey.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteSSHKey(context.Background(), "ssh-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Registry Tests
// =============================================================================

func TestGetRegistries_Success(t *testing.T) {
	registries := []Registry{
		{RegistryID: "reg-1", RegistryName: "Registry One"},
		{RegistryID: "reg-2", RegistryName: "Registry Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/registry.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registries)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetRegistries(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 registries, got %d", len(result))
	}
}

func TestGetRegistry_Success(t *testing.T) {
	registries := []Registry{
		{RegistryID: "reg-1", RegistryName: "Registry One"},
		{RegistryID: "reg-2", RegistryName: "Registry Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registries)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetRegistry(context.Background(), "reg-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RegistryID != "reg-2" {
		t.Errorf("expected registry ID reg-2, got %s", result.RegistryID)
	}
}

func TestGetRegistry_NotFound(t *testing.T) {
	registries := []Registry{{RegistryID: "reg-1", RegistryName: "Registry One"}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registries)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetRegistry(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetRegistryByName_Success(t *testing.T) {
	registries := []Registry{
		{RegistryID: "reg-1", RegistryName: "Docker Hub"},
		{RegistryID: "reg-2", RegistryName: "GitHub Registry"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(registries)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetRegistryByName(context.Background(), "GitHub Registry")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RegistryID != "reg-2" {
		t.Errorf("expected registry ID reg-2, got %s", result.RegistryID)
	}
}

func TestCreateRegistry_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/registry.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateRegistryRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.RegistryName != "New Registry" {
			t.Errorf("expected name 'New Registry', got %s", req.RegistryName)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateRegistryResponse{RegistryID: "new-reg-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateRegistry(context.Background(), CreateRegistryRequest{
		RegistryName:   "New Registry",
		Username:       "user",
		Password:       "pass",
		RegistryURL:    "https://registry.example.com",
		RegistryType:   "selfHosted",
		OrganizationID: "org-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RegistryID != "new-reg-id" {
		t.Errorf("expected registry ID new-reg-id, got %s", result.RegistryID)
	}
}

func TestUpdateRegistry_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/registry.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Registry"
	err := client.UpdateRegistry(context.Background(), UpdateRegistryRequest{
		RegistryID:   "reg-1",
		RegistryName: &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRegistry_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/registry.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteRegistry(context.Background(), "reg-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Certificate Tests
// =============================================================================

func TestGetCertificates_Success(t *testing.T) {
	certs := []Certificate{
		{CertificateID: "cert-1", Name: "Cert One"},
		{CertificateID: "cert-2", Name: "Cert Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/certificates.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(certs)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetCertificates(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 certificates, got %d", len(result))
	}
}

func TestGetCertificate_Success(t *testing.T) {
	certs := []Certificate{
		{CertificateID: "cert-1", Name: "Cert One"},
		{CertificateID: "cert-2", Name: "Cert Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(certs)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetCertificate(context.Background(), "cert-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CertificateID != "cert-2" {
		t.Errorf("expected certificate ID cert-2, got %s", result.CertificateID)
	}
}

func TestGetCertificate_NotFound(t *testing.T) {
	certs := []Certificate{{CertificateID: "cert-1", Name: "Cert One"}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(certs)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetCertificate(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetCertificateByName_Success(t *testing.T) {
	certs := []Certificate{
		{CertificateID: "cert-1", Name: "Wildcard Cert"},
		{CertificateID: "cert-2", Name: "Domain Cert"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(certs)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetCertificateByName(context.Background(), "Domain Cert")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CertificateID != "cert-2" {
		t.Errorf("expected certificate ID cert-2, got %s", result.CertificateID)
	}
}

func TestCreateCertificate_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/certificates.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateCertificateRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Cert" {
			t.Errorf("expected name 'New Cert', got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateCertificateResponse{CertificateID: "new-cert-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateCertificate(context.Background(), CreateCertificateRequest{
		Name:            "New Cert",
		CertificateData: "cert-data",
		PrivateKey:      "private-key",
		OrganizationID:  "org-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.CertificateID != "new-cert-id" {
		t.Errorf("expected certificate ID new-cert-id, got %s", result.CertificateID)
	}
}

func TestUpdateCertificate_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/certificates.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Cert"
	err := client.UpdateCertificate(context.Background(), UpdateCertificateRequest{
		CertificateID: "cert-1",
		Name:          &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCertificate_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/certificates.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteCertificate(context.Background(), "cert-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Destination Tests
// =============================================================================

func TestGetDestinations_Success(t *testing.T) {
	destinations := []Destination{
		{DestinationID: "dest-1", Name: "Destination One"},
		{DestinationID: "dest-2", Name: "Destination Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/destination.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(destinations)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetDestinations(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 destinations, got %d", len(result))
	}
}

func TestGetDestination_Success(t *testing.T) {
	destinations := []Destination{
		{DestinationID: "dest-1", Name: "Destination One"},
		{DestinationID: "dest-2", Name: "Destination Two"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(destinations)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetDestination(context.Background(), "dest-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DestinationID != "dest-2" {
		t.Errorf("expected destination ID dest-2, got %s", result.DestinationID)
	}
}

func TestGetDestination_NotFound(t *testing.T) {
	destinations := []Destination{{DestinationID: "dest-1", Name: "Destination One"}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(destinations)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetDestination(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetDestinationByName_Success(t *testing.T) {
	destinations := []Destination{
		{DestinationID: "dest-1", Name: "S3 Backup"},
		{DestinationID: "dest-2", Name: "GCS Backup"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(destinations)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetDestinationByName(context.Background(), "GCS Backup")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DestinationID != "dest-2" {
		t.Errorf("expected destination ID dest-2, got %s", result.DestinationID)
	}
}

func TestCreateDestination_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/destination.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateDestinationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Destination" {
			t.Errorf("expected name 'New Destination', got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateDestinationResponse{DestinationID: "new-dest-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateDestination(context.Background(), CreateDestinationRequest{
		Name:            "New Destination",
		AccessKey:       "access-key",
		SecretAccessKey: "secret-key",
		Bucket:          "my-bucket",
		Region:          "us-east-1",
		Endpoint:        "https://s3.amazonaws.com",
		OrganizationID:  "org-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DestinationID != "new-dest-id" {
		t.Errorf("expected destination ID new-dest-id, got %s", result.DestinationID)
	}
}

func TestUpdateDestination_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/destination.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Destination"
	err := client.UpdateDestination(context.Background(), UpdateDestinationRequest{
		DestinationID: "dest-1",
		Name:          &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteDestination_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/destination.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteDestination(context.Background(), "dest-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Error Cases
// =============================================================================

func TestCreateSSHKey_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid key format"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateSSHKey(context.Background(), CreateSSHKeyRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRegistry_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid registry URL"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateRegistry(context.Background(), CreateRegistryRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateCertificate_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid certificate"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateCertificate(context.Background(), CreateCertificateRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateDestination_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid S3 credentials"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateDestination(context.Background(), CreateDestinationRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
