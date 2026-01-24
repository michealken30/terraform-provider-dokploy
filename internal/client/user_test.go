package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// =============================================================================
// User Tests
// =============================================================================

func TestGetUsers_Success(t *testing.T) {
	users := []User{
		{ID: "user-1", Name: "Alice", Email: "alice@example.com"},
		{ID: "user-2", Name: "Bob", Email: "bob@example.com"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetUsers(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 users, got %d", len(result))
	}
}

func TestGetUsers_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal error"}`))
	})
	defer server.Close()

	client := NewWithRetry(server.URL, "test-api-key", RetryConfig{MaxRetries: 0})
	_, err := client.GetUsers(context.Background())

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetUser_Success(t *testing.T) {
	users := []User{
		{ID: "user-1", Name: "Alice", Email: "alice@example.com"},
		{ID: "user-2", Name: "Bob", Email: "bob@example.com"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetUser(context.Background(), "user-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ID != "user-2" {
		t.Errorf("expected user ID user-2, got %s", result.ID)
	}
	if result.Name != "Bob" {
		t.Errorf("expected name 'Bob', got %s", result.Name)
	}
}

func TestGetUser_NotFound(t *testing.T) {
	users := []User{
		{ID: "user-1", Name: "Alice", Email: "alice@example.com"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(users)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetUser(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateUser_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ID != "user-1" {
			t.Errorf("expected user ID user-1, got %s", req.ID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Name"
	err := client.UpdateUser(context.Background(), UpdateUserRequest{
		ID:   "user-1",
		Name: &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateUser_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "user not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateUser(context.Background(), UpdateUserRequest{
		ID: "non-existent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteUser_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req struct {
			UserID string `json:"userId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.UserID != "user-to-delete" {
			t.Errorf("expected user ID user-to-delete, got %s", req.UserID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteUser(context.Background(), "user-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteUser_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "cannot delete admin"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteUser(context.Background(), "admin-user")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestAssignUserPermissions_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/user.assignPermissions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UserPermissionsRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.ID != "user-1" {
			t.Errorf("expected user ID user-1, got %s", req.ID)
		}
		if !req.CanCreateProjects {
			t.Errorf("expected CanCreateProjects to be true")
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.AssignUserPermissions(context.Background(), UserPermissionsRequest{
		ID:                "user-1",
		AccessedProjects:  []string{"proj-1", "proj-2"},
		CanCreateProjects: true,
		CanDeleteProjects: false,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestAssignUserPermissions_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "user not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.AssignUserPermissions(context.Background(), UserPermissionsRequest{
		ID: "non-existent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// =============================================================================
// Volume Backup Tests
// =============================================================================

func TestCreateVolumeBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/volumeBackups.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateVolumeBackupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "Daily Volume Backup" {
			t.Errorf("expected name 'Daily Volume Backup', got %s", req.Name)
		}
		if req.VolumeName != "app-data" {
			t.Errorf("expected volume name 'app-data', got %s", req.VolumeName)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(VolumeBackup{
			VolumeBackupID: "new-vb-id",
			Name:           "Daily Volume Backup",
			VolumeName:     "app-data",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	appID := "app-123"
	result, err := client.CreateVolumeBackup(context.Background(), CreateVolumeBackupRequest{
		Name:           "Daily Volume Backup",
		VolumeName:     "app-data",
		Prefix:         "daily",
		CronExpression: "0 0 * * *",
		DestinationID:  "dest-123",
		ApplicationID:  &appID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VolumeBackupID != "new-vb-id" {
		t.Errorf("expected volume backup ID new-vb-id, got %s", result.VolumeBackupID)
	}
}

func TestCreateVolumeBackup_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid cron expression"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateVolumeBackup(context.Background(), CreateVolumeBackupRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetVolumeBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/volumeBackups.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(VolumeBackup{
			VolumeBackupID: "vb-123",
			Name:           "Daily Volume Backup",
			VolumeName:     "app-data",
			CronExpression: "0 0 * * *",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetVolumeBackup(context.Background(), "vb-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VolumeBackupID != "vb-123" {
		t.Errorf("expected volume backup ID vb-123, got %s", result.VolumeBackupID)
	}
}

func TestGetVolumeBackup_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "volume backup not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetVolumeBackup(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestUpdateVolumeBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/volumeBackups.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateVolumeBackupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.VolumeBackupID != "vb-1" {
			t.Errorf("expected volume backup ID vb-1, got %s", req.VolumeBackupID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateVolumeBackup(context.Background(), UpdateVolumeBackupRequest{
		VolumeBackupID: "vb-1",
		Name:           "Updated Volume Backup",
		VolumeName:     "app-data",
		Prefix:         "updated",
		CronExpression: "0 0 * * *",
		DestinationID:  "dest-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateVolumeBackup_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "volume backup not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateVolumeBackup(context.Background(), UpdateVolumeBackupRequest{
		VolumeBackupID: "non-existent",
	})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteVolumeBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/volumeBackups.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req struct {
			VolumeBackupID string `json:"volumeBackupId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.VolumeBackupID != "vb-to-delete" {
			t.Errorf("expected volume backup ID vb-to-delete, got %s", req.VolumeBackupID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteVolumeBackup(context.Background(), "vb-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteVolumeBackup_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "backup running"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteVolumeBackup(context.Background(), "running-backup")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
