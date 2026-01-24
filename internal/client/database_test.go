package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// =============================================================================
// Postgres Tests
// =============================================================================

func TestGetPostgres_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					Postgres: []Postgres{
						{PostgresID: "pg-1", Name: "Postgres One", DatabaseName: "db1"},
						{PostgresID: "pg-2", Name: "Postgres Two", DatabaseName: "db2"},
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
	result, err := client.GetPostgres(context.Background(), "pg-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PostgresID != "pg-2" {
		t.Errorf("expected postgres ID pg-2, got %s", result.PostgresID)
	}
	if result.DatabaseName != "db2" {
		t.Errorf("expected database name 'db2', got %s", result.DatabaseName)
	}
}

func TestGetPostgres_NotFound(t *testing.T) {
	projects := []Project{{ProjectID: "proj-1", Environments: []Environment{{EnvironmentID: "env-1"}}}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetPostgres(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreatePostgres_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/postgres.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreatePostgresRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Postgres" {
			t.Errorf("expected name 'New Postgres', got %s", req.Name)
		}
		if req.DatabaseName != "testdb" {
			t.Errorf("expected database name 'testdb', got %s", req.DatabaseName)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreatePostgresResponse{PostgresID: "new-pg-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreatePostgres(context.Background(), CreatePostgresRequest{
		Name:             "New Postgres",
		AppName:          "postgres-app",
		EnvironmentID:    "env-123",
		DatabaseName:     "testdb",
		DatabaseUser:     "admin",
		DatabasePassword: "secret",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PostgresID != "new-pg-id" {
		t.Errorf("expected postgres ID new-pg-id, got %s", result.PostgresID)
	}
}

func TestUpdatePostgres_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/postgres.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Postgres"
	err := client.UpdatePostgres(context.Background(), UpdatePostgresRequest{
		PostgresID: "pg-1",
		Name:       &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeletePostgres_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/postgres.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeletePostgres(context.Background(), "pg-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// MySQL Tests
// =============================================================================

func TestGetMysql_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					MySQL: []MySQL{
						{MySQLID: "mysql-1", Name: "MySQL One", DatabaseName: "db1"},
						{MySQLID: "mysql-2", Name: "MySQL Two", DatabaseName: "db2"},
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
	result, err := client.GetMysql(context.Background(), "mysql-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MySQLID != "mysql-2" {
		t.Errorf("expected mysql ID mysql-2, got %s", result.MySQLID)
	}
}

func TestGetMysql_NotFound(t *testing.T) {
	projects := []Project{{ProjectID: "proj-1", Environments: []Environment{{EnvironmentID: "env-1"}}}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetMysql(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateMysql_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mysql.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateMysqlRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.DatabaseRootPassword != "rootpass" {
			t.Errorf("expected root password 'rootpass', got %s", req.DatabaseRootPassword)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateMysqlResponse{MysqlID: "new-mysql-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateMysql(context.Background(), CreateMysqlRequest{
		Name:                 "New MySQL",
		AppName:              "mysql-app",
		EnvironmentID:        "env-123",
		DatabaseName:         "testdb",
		DatabaseUser:         "admin",
		DatabasePassword:     "secret",
		DatabaseRootPassword: "rootpass",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MysqlID != "new-mysql-id" {
		t.Errorf("expected mysql ID new-mysql-id, got %s", result.MysqlID)
	}
}

func TestUpdateMysql_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mysql.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated MySQL"
	err := client.UpdateMysql(context.Background(), UpdateMysqlRequest{
		MysqlID: "mysql-1",
		Name:    &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMysql_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mysql.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteMysql(context.Background(), "mysql-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// MariaDB Tests
// =============================================================================

func TestGetMariadb_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					MariaDB: []MariaDB{
						{MariaDBID: "maria-1", Name: "MariaDB One", DatabaseName: "db1"},
						{MariaDBID: "maria-2", Name: "MariaDB Two", DatabaseName: "db2"},
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
	result, err := client.GetMariadb(context.Background(), "maria-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MariaDBID != "maria-2" {
		t.Errorf("expected mariadb ID maria-2, got %s", result.MariaDBID)
	}
}

func TestGetMariadb_NotFound(t *testing.T) {
	projects := []Project{{ProjectID: "proj-1", Environments: []Environment{{EnvironmentID: "env-1"}}}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetMariadb(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateMariadb_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mariadb.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateMariadbResponse{MariadbID: "new-maria-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateMariadb(context.Background(), CreateMariadbRequest{
		Name:                 "New MariaDB",
		AppName:              "mariadb-app",
		EnvironmentID:        "env-123",
		DatabaseName:         "testdb",
		DatabaseUser:         "admin",
		DatabasePassword:     "secret",
		DatabaseRootPassword: "rootpass",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MariadbID != "new-maria-id" {
		t.Errorf("expected mariadb ID new-maria-id, got %s", result.MariadbID)
	}
}

func TestUpdateMariadb_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mariadb.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated MariaDB"
	err := client.UpdateMariadb(context.Background(), UpdateMariadbRequest{
		MariadbID: "maria-1",
		Name:      &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMariadb_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mariadb.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteMariadb(context.Background(), "maria-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// MongoDB Tests
// =============================================================================

func TestGetMongo_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					Mongo: []Mongo{
						{MongoID: "mongo-1", Name: "Mongo One", DatabaseUser: "user1"},
						{MongoID: "mongo-2", Name: "Mongo Two", DatabaseUser: "user2"},
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
	result, err := client.GetMongo(context.Background(), "mongo-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MongoID != "mongo-2" {
		t.Errorf("expected mongo ID mongo-2, got %s", result.MongoID)
	}
}

func TestGetMongo_NotFound(t *testing.T) {
	projects := []Project{{ProjectID: "proj-1", Environments: []Environment{{EnvironmentID: "env-1"}}}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetMongo(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateMongo_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mongo.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateMongoResponse{MongoID: "new-mongo-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateMongo(context.Background(), CreateMongoRequest{
		Name:             "New Mongo",
		AppName:          "mongo-app",
		EnvironmentID:    "env-123",
		DatabaseUser:     "admin",
		DatabasePassword: "secret",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.MongoID != "new-mongo-id" {
		t.Errorf("expected mongo ID new-mongo-id, got %s", result.MongoID)
	}
}

func TestUpdateMongo_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mongo.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Mongo"
	err := client.UpdateMongo(context.Background(), UpdateMongoRequest{
		MongoID: "mongo-1",
		Name:    &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteMongo_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/mongo.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteMongo(context.Background(), "mongo-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Redis Tests
// =============================================================================

func TestGetRedis_Success(t *testing.T) {
	projects := []Project{
		{
			ProjectID: "proj-1",
			Environments: []Environment{
				{
					EnvironmentID: "env-1",
					Redis: []Redis{
						{RedisID: "redis-1", Name: "Redis One"},
						{RedisID: "redis-2", Name: "Redis Two"},
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
	result, err := client.GetRedis(context.Background(), "redis-2")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RedisID != "redis-2" {
		t.Errorf("expected redis ID redis-2, got %s", result.RedisID)
	}
}

func TestGetRedis_NotFound(t *testing.T) {
	projects := []Project{{ProjectID: "proj-1", Environments: []Environment{{EnvironmentID: "env-1"}}}}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(projects)
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.GetRedis(context.Background(), "non-existent")

	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("expected NotFoundError, got %T", err)
	}
}

func TestCreateRedis_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/redis.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateRedisResponse{RedisID: "new-redis-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateRedis(context.Background(), CreateRedisRequest{
		Name:             "New Redis",
		AppName:          "redis-app",
		EnvironmentID:    "env-123",
		DatabasePassword: "secret",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.RedisID != "new-redis-id" {
		t.Errorf("expected redis ID new-redis-id, got %s", result.RedisID)
	}
}

func TestUpdateRedis_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/redis.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	newName := "Updated Redis"
	err := client.UpdateRedis(context.Background(), UpdateRedisRequest{
		RedisID: "redis-1",
		Name:    &newName,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteRedis_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/redis.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteRedis(context.Background(), "redis-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Error Cases
// =============================================================================

func TestCreatePostgres_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid configuration"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreatePostgres(context.Background(), CreatePostgresRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateMysql_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid configuration"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateMysql(context.Background(), CreateMysqlRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateMariadb_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid configuration"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateMariadb(context.Background(), CreateMariadbRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateMongo_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid configuration"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateMongo(context.Background(), CreateMongoRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateRedis_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid configuration"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateRedis(context.Background(), CreateRedisRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
