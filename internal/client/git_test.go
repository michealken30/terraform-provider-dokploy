package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// =============================================================================
// GitLab Tests
// =============================================================================

func TestCreateGitlab_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitlab.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateGitlabRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "My GitLab" {
			t.Errorf("expected name 'My GitLab', got %s", req.Name)
		}
		if req.GitlabURL != "https://gitlab.example.com" {
			t.Errorf("expected gitlab URL 'https://gitlab.example.com', got %s", req.GitlabURL)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GitlabResponse{
			GitlabID:      "new-gitlab-id",
			GitProviderID: "provider-123",
			Name:          "My GitLab",
			GitlabURL:     "https://gitlab.example.com",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateGitlab(context.Background(), CreateGitlabRequest{
		Name:      "My GitLab",
		GitlabURL: "https://gitlab.example.com",
		AuthID:    "auth-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GitlabID != "new-gitlab-id" {
		t.Errorf("expected gitlab ID new-gitlab-id, got %s", result.GitlabID)
	}
}

func TestGetGitlab_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitlab.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GitlabResponse{
			GitlabID:      "gitlab-123",
			GitProviderID: "provider-123",
			Name:          "My GitLab",
			GitlabURL:     "https://gitlab.example.com",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetGitlab(context.Background(), "gitlab-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GitlabID != "gitlab-123" {
		t.Errorf("expected gitlab ID gitlab-123, got %s", result.GitlabID)
	}
}

func TestUpdateGitlab_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitlab.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req UpdateGitlabRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.GitlabID != "gitlab-1" {
			t.Errorf("expected gitlab ID gitlab-1, got %s", req.GitlabID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateGitlab(context.Background(), UpdateGitlabRequest{
		GitlabID:      "gitlab-1",
		GitProviderID: "provider-1",
		Name:          "Updated GitLab",
		GitlabURL:     "https://gitlab.example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Bitbucket Tests
// =============================================================================

func TestCreateBitbucket_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bitbucket.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateBitbucketRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "My Bitbucket" {
			t.Errorf("expected name 'My Bitbucket', got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BitbucketResponse{
			BitbucketID:   "new-bitbucket-id",
			GitProviderID: "provider-123",
			Name:          "My Bitbucket",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateBitbucket(context.Background(), CreateBitbucketRequest{
		Name:   "My Bitbucket",
		AuthID: "auth-123",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BitbucketID != "new-bitbucket-id" {
		t.Errorf("expected bitbucket ID new-bitbucket-id, got %s", result.BitbucketID)
	}
}

func TestGetBitbucket_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bitbucket.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(BitbucketResponse{
			BitbucketID:   "bitbucket-123",
			GitProviderID: "provider-123",
			Name:          "My Bitbucket",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetBitbucket(context.Background(), "bitbucket-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BitbucketID != "bitbucket-123" {
		t.Errorf("expected bitbucket ID bitbucket-123, got %s", result.BitbucketID)
	}
}

func TestUpdateBitbucket_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/bitbucket.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateBitbucket(context.Background(), UpdateBitbucketRequest{
		BitbucketID:   "bitbucket-1",
		GitProviderID: "provider-1",
		Name:          "Updated Bitbucket",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Gitea Tests
// =============================================================================

func TestCreateGitea_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitea.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateGiteaRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "My Gitea" {
			t.Errorf("expected name 'My Gitea', got %s", req.Name)
		}
		if req.GiteaURL != "https://gitea.example.com" {
			t.Errorf("expected gitea URL 'https://gitea.example.com', got %s", req.GiteaURL)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GiteaResponse{
			GiteaID:       "new-gitea-id",
			GitProviderID: "provider-123",
			Name:          "My Gitea",
			GiteaURL:      "https://gitea.example.com",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateGitea(context.Background(), CreateGiteaRequest{
		Name:     "My Gitea",
		GiteaURL: "https://gitea.example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GiteaID != "new-gitea-id" {
		t.Errorf("expected gitea ID new-gitea-id, got %s", result.GiteaID)
	}
}

func TestGetGitea_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitea.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(GiteaResponse{
			GiteaID:       "gitea-123",
			GitProviderID: "provider-123",
			Name:          "My Gitea",
			GiteaURL:      "https://gitea.example.com",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetGitea(context.Background(), "gitea-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GiteaID != "gitea-123" {
		t.Errorf("expected gitea ID gitea-123, got %s", result.GiteaID)
	}
}

func TestUpdateGitea_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitea.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateGitea(context.Background(), UpdateGiteaRequest{
		GiteaID:       "gitea-1",
		GitProviderID: "provider-1",
		Name:          "Updated Gitea",
		GiteaURL:      "https://gitea.example.com",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// GitHub Tests
// =============================================================================

func TestGetGithub_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/github.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Github{
			GithubID:      "github-123",
			GitProviderID: "provider-123",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetGithub(context.Background(), "github-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.GithubID != "github-123" {
		t.Errorf("expected github ID github-123, got %s", result.GithubID)
	}
}

func TestGetGithub_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": "github not found"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetGithub(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// =============================================================================
// Git Provider Tests
// =============================================================================

func TestGetGitProviders_Success(t *testing.T) {
	providers := []GitProvider{
		{GitProviderID: "prov-1", Name: "GitLab", ProviderType: "gitlab"},
		{GitProviderID: "prov-2", Name: "Bitbucket", ProviderType: "bitbucket"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitProvider.getAll" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(providers)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetGitProviders(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 git providers, got %d", len(result))
	}
}

func TestDeleteGitProvider_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/gitProvider.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req struct {
			GitProviderID string `json:"gitProviderId"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.GitProviderID != "provider-to-delete" {
			t.Errorf("expected git provider ID provider-to-delete, got %s", req.GitProviderID)
		}

		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteGitProvider(context.Background(), "provider-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Error Cases
// =============================================================================

func TestCreateGitlab_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid gitlab URL"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateGitlab(context.Background(), CreateGitlabRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateBitbucket_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid credentials"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateBitbucket(context.Background(), CreateBitbucketRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateGitea_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid gitea URL"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateGitea(context.Background(), CreateGiteaRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestDeleteGitProvider_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "provider in use"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteGitProvider(context.Background(), "provider-in-use")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
