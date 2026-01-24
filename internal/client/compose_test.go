package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetCompose_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					Compose: []Compose{
						{ComposeID: "compose-1", Name: "Compose One"},
						{ComposeID: "compose-2", Name: "Compose Two"},
					},
				},
			},
		},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetCompose(context.Background(), "compose-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ComposeID != "compose-2" {
		t.Errorf("expected compose ID compose-2, got %s", result.ComposeID)
	}
	if result.Name != "Compose Two" {
		t.Errorf("expected name 'Compose Two', got %s", result.Name)
	}
}

func TestGetCompose_NotFound(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					Compose: []Compose{
						{ComposeID: "compose-1", Name: "Compose One"},
					},
				},
			},
		},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetCompose(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetCompose_EmptyProjects(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]Project{})
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetCompose(context.Background(), "compose-1")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateCompose_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/compose.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateComposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Compose" {
			t.Errorf("expected name 'New Compose', got %s", req.Name)
		}
		if req.EnvironmentID != "env-123" {
			t.Errorf("expected environment ID 'env-123', got %s", req.EnvironmentID)
		}
		if req.ComposeType != "docker-compose" {
			t.Errorf("expected compose type 'docker-compose', got %s", req.ComposeType)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateComposeResponse{ComposeID: "new-compose-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateCompose(context.Background(), CreateComposeRequest{
		Name:          "New Compose",
		EnvironmentID: "env-123",
		ComposeType:   "docker-compose",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ComposeID != "new-compose-id" {
		t.Errorf("expected compose ID new-compose-id, got %s", result.ComposeID)
	}
}

func TestCreateCompose_WithOptionalFields(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateComposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Description == nil || *req.Description != "Test description" {
			t.Errorf("expected description 'Test description', got %v", req.Description)
		}
		if req.ServerID == nil || *req.ServerID != "srv-123" {
			t.Errorf("expected server ID 'srv-123', got %v", req.ServerID)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateComposeResponse{ComposeID: "compose-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	desc := "Test description"
	serverID := "srv-123"
	result, err := client.CreateCompose(context.Background(), CreateComposeRequest{
		Name:          "New Compose",
		Description:   &desc,
		EnvironmentID: "env-123",
		ServerID:      &serverID,
		ComposeType:   "docker-compose",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ComposeID != "compose-id" {
		t.Errorf("expected compose ID compose-id, got %s", result.ComposeID)
	}
}

func TestCreateCompose_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid compose type"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateCompose(context.Background(), CreateComposeRequest{
		Name:          "Invalid",
		EnvironmentID: "env-123",
		ComposeType:   "invalid",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateCompose_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/compose.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateComposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ComposeID != "compose-1" {
			t.Errorf("expected compose ID compose-1, got %s", req.ComposeID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Compose"
	err := client.UpdateCompose(context.Background(), UpdateComposeRequest{
		ComposeID: "compose-1",
		Name:      &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCompose_WithComposeFile(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req UpdateComposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ComposeFile == nil || *req.ComposeFile != "version: '3'\nservices:\n  web:\n    image: nginx" {
			t.Errorf("expected compose file content, got %v", req.ComposeFile)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	composeFile := "version: '3'\nservices:\n  web:\n    image: nginx"
	err := client.UpdateCompose(context.Background(), UpdateComposeRequest{
		ComposeID:   "compose-1",
		ComposeFile: &composeFile,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateCompose_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "compose not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateCompose(context.Background(), UpdateComposeRequest{
		ComposeID: "non-existent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteCompose_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/compose.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req DeleteComposeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ComposeID != "compose-to-delete" {
			t.Errorf("expected compose ID compose-to-delete, got %s", req.ComposeID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteCompose(context.Background(), "compose-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCompose_NotFoundTreatedAsSuccess(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteCompose(context.Background(), "already-deleted")

	if err != nil {
		t.Fatalf("expected no error (404 should be success for delete), got: %v", err)
	}
}

func TestDeleteCompose_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "compose is running"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteCompose(context.Background(), "running-compose")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
