package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetServers_Success(t *testing.T) {
	servers := []Server{
		{ServerID: "srv-1", Name: "Server 1", IPAddress: "192.168.1.1"},
		{ServerID: "srv-2", Name: "Server 2", IPAddress: "192.168.1.2"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/server.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetServers(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 servers, got %d", len(result))
	}
	if result[0].ServerID != "srv-1" {
		t.Errorf("expected server ID srv-1, got %s", result[0].ServerID)
	}
}

func TestGetServers_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetServers(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetServer_Success(t *testing.T) {
	servers := []Server{
		{ServerID: "srv-1", Name: "Server 1", IPAddress: "192.168.1.1"},
		{ServerID: "srv-2", Name: "Server 2", IPAddress: "192.168.1.2"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetServer(context.Background(), "srv-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ServerID != "srv-2" {
		t.Errorf("expected server ID srv-2, got %s", result.ServerID)
	}
	if result.IPAddress != "192.168.1.2" {
		t.Errorf("expected IP 192.168.1.2, got %s", result.IPAddress)
	}
}

func TestGetServer_NotFound(t *testing.T) {
	servers := []Server{
		{ServerID: "srv-1", Name: "Server 1"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetServer(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetServerByName_Success(t *testing.T) {
	servers := []Server{
		{ServerID: "srv-1", Name: "Production Server"},
		{ServerID: "srv-2", Name: "Staging Server"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetServerByName(context.Background(), "Staging Server")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ServerID != "srv-2" {
		t.Errorf("expected server ID srv-2, got %s", result.ServerID)
	}
}

func TestGetServerByName_NotFound(t *testing.T) {
	servers := []Server{
		{ServerID: "srv-1", Name: "Production Server"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(servers)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetServerByName(context.Background(), "Non-Existent Server")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateServer_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/server.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Server" {
			t.Errorf("expected name 'New Server', got %s", req.Name)
		}
		if req.IPAddress != "10.0.0.1" {
			t.Errorf("expected IP '10.0.0.1', got %s", req.IPAddress)
		}
		if req.Port != 22 {
			t.Errorf("expected port 22, got %d", req.Port)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateServerResponse{ServerID: "new-srv-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateServer(context.Background(), CreateServerRequest{
		Name:       "New Server",
		IPAddress:  "10.0.0.1",
		Port:       22,
		Username:   "root",
		ServerType: "deploy",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ServerID != "new-srv-id" {
		t.Errorf("expected server ID new-srv-id, got %s", result.ServerID)
	}
}

func TestCreateServer_WithSSHKey(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.SSHKeyID == nil || *req.SSHKeyID != "ssh-key-123" {
			t.Errorf("expected SSH key ID 'ssh-key-123', got %v", req.SSHKeyID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateServerResponse{ServerID: "srv-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	sshKeyID := "ssh-key-123"
	result, err := client.CreateServer(context.Background(), CreateServerRequest{
		Name:       "New Server",
		IPAddress:  "10.0.0.1",
		Port:       22,
		Username:   "root",
		ServerType: "deploy",
		SSHKeyID:   &sshKeyID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ServerID != "srv-id" {
		t.Errorf("expected server ID srv-id, got %s", result.ServerID)
	}
}

func TestCreateServer_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid IP address"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateServer(context.Background(), CreateServerRequest{
		Name:       "Bad Server",
		IPAddress:  "invalid-ip",
		Port:       22,
		Username:   "root",
		ServerType: "deploy",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateServer_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/server.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ServerID != "srv-1" {
			t.Errorf("expected server ID srv-1, got %s", req.ServerID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Server"
	err := client.UpdateServer(context.Background(), UpdateServerRequest{
		ServerID: "srv-1",
		Name:     &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateServer_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "server not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateServer(context.Background(), UpdateServerRequest{
		ServerID: "non-existent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteServer_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/server.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req DeleteServerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ServerID != "srv-to-delete" {
			t.Errorf("expected server ID srv-to-delete, got %s", req.ServerID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteServer(context.Background(), "srv-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteServer_NotFoundTreatedAsSuccess(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteServer(context.Background(), "already-deleted")

	if err != nil {
		t.Fatalf("expected no error (404 should be success for delete), got: %v", err)
	}
}

func TestDeleteServer_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "server has active deployments"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteServer(context.Background(), "srv-with-deployments")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
