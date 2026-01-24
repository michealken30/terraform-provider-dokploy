package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

func TestGetProjects_Success(t *testing.T) {
	projects := []Project{
		{ProjectID: "proj-1", Name: "Project 1"},
		{ProjectID: "proj-2", Name: "Project 2"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/project.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetProjects(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 projects, got %d", len(result))
	}
	if result[0].ProjectID != "proj-1" {
		t.Errorf("expected project ID proj-1, got %s", result[0].ProjectID)
	}
}

func TestGetProjects_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetProjects(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetProject_Success(t *testing.T) {
	projects := []Project{
		{ProjectID: "proj-1", Name: "Project 1"},
		{ProjectID: "proj-2", Name: "Project 2"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetProject(context.Background(), "proj-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectID != "proj-2" {
		t.Errorf("expected project ID proj-2, got %s", result.ProjectID)
	}
	if result.Name != "Project 2" {
		t.Errorf("expected name Project 2, got %s", result.Name)
	}
}

func TestGetProject_NotFound(t *testing.T) {
	projects := []Project{
		{ProjectID: "proj-1", Name: "Project 1"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetProject(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestGetProjectByName_Success(t *testing.T) {
	projects := []Project{
		{ProjectID: "proj-1", Name: "Project Alpha"},
		{ProjectID: "proj-2", Name: "Project Beta"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetProjectByName(context.Background(), "Project Beta")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProjectID != "proj-2" {
		t.Errorf("expected project ID proj-2, got %s", result.ProjectID)
	}
}

func TestGetProjectByName_NotFound(t *testing.T) {
	projects := []Project{
		{ProjectID: "proj-1", Name: "Project Alpha"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetProjectByName(context.Background(), "Non-Existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateProject_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/project.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Project" {
			t.Errorf("expected name 'New Project', got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateProjectResponse{
			Project: struct {
				ProjectID string `json:"projectId"`
			}{ProjectID: "new-proj-id"},
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateProject(context.Background(), CreateProjectRequest{
		Name: "New Project",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Project.ProjectID != "new-proj-id" {
		t.Errorf("expected project ID new-proj-id, got %s", result.Project.ProjectID)
	}
}

func TestCreateProject_WithDescription(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		var req CreateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Description == nil || *req.Description != "Test description" {
			t.Errorf("expected description 'Test description', got %v", req.Description)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateProjectResponse{
			Project: struct {
				ProjectID string `json:"projectId"`
			}{ProjectID: "proj-id"},
		})
	})
	defer server.Close()

	client := newTestClient(server)
	desc := "Test description"
	result, err := client.CreateProject(context.Background(), CreateProjectRequest{
		Name:        "New Project",
		Description: &desc,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Project.ProjectID != "proj-id" {
		t.Errorf("expected project ID proj-id, got %s", result.Project.ProjectID)
	}
}

func TestCreateProject_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "name already exists"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateProject(context.Background(), CreateProjectRequest{
		Name: "Duplicate",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateProject_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/project.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ProjectID != "proj-1" {
			t.Errorf("expected project ID proj-1, got %s", req.ProjectID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Name"
	err := client.UpdateProject(context.Background(), UpdateProjectRequest{
		ProjectID: "proj-1",
		Name:      &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateProject_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "project not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateProject(context.Background(), UpdateProjectRequest{
		ProjectID: "non-existent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteProject_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/api/project.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req DeleteProjectRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ProjectID != "proj-to-delete" {
			t.Errorf("expected project ID proj-to-delete, got %s", req.ProjectID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteProject(context.Background(), "proj-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteProject_NotFoundTreatedAsSuccess(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteProject(context.Background(), "already-deleted")

	if err != nil {
		t.Fatalf("expected no error (404 should be success for delete), got: %v", err)
	}
}

func TestDeleteProject_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "project has active services"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteProject(context.Background(), "proj-with-services")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
