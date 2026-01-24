package client

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"
)

// =============================================================================
// Backup Tests
// =============================================================================

func TestCreateBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/backup.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateBackupRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Schedule != "0 0 * * *" {
			t.Errorf("expected schedule '0 0 * * *', got %s", req.Schedule)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateBackupResponse{BackupID: "new-backup-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	postgresID := "pg-123"
	result, err := client.CreateBackup(context.Background(), CreateBackupRequest{
		Schedule:      "0 0 * * *",
		Prefix:        "daily",
		DestinationID: "dest-123",
		Database:      "testdb",
		DatabaseType:  "postgres",
		BackupType:    "full",
		PostgresID:    &postgresID,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BackupID != "new-backup-id" {
		t.Errorf("expected backup ID new-backup-id, got %s", result.BackupID)
	}
}

func TestGetBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/backup.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Backup{
			BackupID: "backup-123",
			Schedule: "0 0 * * *",
			Enabled:  true,
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetBackup(context.Background(), "backup-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.BackupID != "backup-123" {
		t.Errorf("expected backup ID backup-123, got %s", result.BackupID)
	}
}

func TestUpdateBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/backup.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	enabled := true
	err := client.UpdateBackup(context.Background(), UpdateBackupRequest{
		BackupID: "backup-1",
		Enabled:  &enabled,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteBackup_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/backup.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteBackup(context.Background(), "backup-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Schedule Tests
// =============================================================================

func TestCreateSchedule_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/schedule.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateScheduleRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "Daily Cleanup" {
			t.Errorf("expected name 'Daily Cleanup', got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateScheduleResponse{ScheduleID: "new-schedule-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	appID := "app-123"
	result, err := client.CreateSchedule(context.Background(), CreateScheduleRequest{
		Name:           "Daily Cleanup",
		CronExpression: "0 0 * * *",
		Command:        "cleanup.sh",
		ShellType:      "bash",
		ScheduleType:   "application",
		ApplicationID:  &appID,
		Enabled:        true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ScheduleID != "new-schedule-id" {
		t.Errorf("expected schedule ID new-schedule-id, got %s", result.ScheduleID)
	}
}

func TestGetSchedule_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/schedule.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Schedule{
			ScheduleID:     "schedule-123",
			Name:           "Daily Cleanup",
			CronExpression: "0 0 * * *",
			Enabled:        true,
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetSchedule(context.Background(), "schedule-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ScheduleID != "schedule-123" {
		t.Errorf("expected schedule ID schedule-123, got %s", result.ScheduleID)
	}
}

func TestUpdateSchedule_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/schedule.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	enabled := false
	err := client.UpdateSchedule(context.Background(), UpdateScheduleRequest{
		ScheduleID: "schedule-1",
		Enabled:    &enabled,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteSchedule_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/schedule.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteSchedule(context.Background(), "schedule-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Notification Tests
// =============================================================================

func TestGetNotifications_Success(t *testing.T) {
	notifications := []Notification{
		{NotificationID: "notif-1", Name: "Slack Alert", NotificationType: "slack"},
		{NotificationID: "notif-2", Name: "Discord Alert", NotificationType: "discord"},
	}

	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.all" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(notifications)
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetNotifications(context.Background())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Errorf("expected 2 notifications, got %d", len(result))
	}
}

func TestGetNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(NotificationResponse{
			NotificationID:   "notif-123",
			Name:             "Slack Alert",
			NotificationType: "slack",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetNotification(context.Background(), "notif-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotificationID != "notif-123" {
		t.Errorf("expected notification ID notif-123, got %s", result.NotificationID)
	}
}

func TestCreateSlackNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.createSlack" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateSlackNotificationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Channel != "#alerts" {
			t.Errorf("expected channel '#alerts', got %s", req.Channel)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateNotificationResponse{NotificationID: "new-slack-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateSlackNotification(context.Background(), CreateSlackNotificationRequest{
		NotificationBase: NotificationBase{
			Name:      "Slack Alert",
			AppDeploy: true,
		},
		WebhookURL: "https://hooks.slack.com/...",
		Channel:    "#alerts",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotificationID != "new-slack-id" {
		t.Errorf("expected notification ID new-slack-id, got %s", result.NotificationID)
	}
}

func TestCreateDiscordNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.createDiscord" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateNotificationResponse{NotificationID: "new-discord-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateDiscordNotification(context.Background(), CreateDiscordNotificationRequest{
		NotificationBase: NotificationBase{
			Name:      "Discord Alert",
			AppDeploy: true,
		},
		WebhookURL: "https://discord.com/api/webhooks/...",
		Decoration: true,
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotificationID != "new-discord-id" {
		t.Errorf("expected notification ID new-discord-id, got %s", result.NotificationID)
	}
}

func TestCreateTelegramNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.createTelegram" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateNotificationResponse{NotificationID: "new-telegram-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateTelegramNotification(context.Background(), CreateTelegramNotificationRequest{
		NotificationBase: NotificationBase{
			Name:      "Telegram Alert",
			AppDeploy: true,
		},
		BotToken: "123456:ABC-DEF",
		ChatID:   "-1001234567890",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotificationID != "new-telegram-id" {
		t.Errorf("expected notification ID new-telegram-id, got %s", result.NotificationID)
	}
}

func TestCreateEmailNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.createEmail" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateNotificationResponse{NotificationID: "new-email-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateEmailNotification(context.Background(), CreateEmailNotificationRequest{
		NotificationBase: NotificationBase{
			Name:      "Email Alert",
			AppDeploy: true,
		},
		SmtpServer:  "smtp.example.com",
		SmtpPort:    587,
		Username:    "user@example.com",
		Password:    "secret",
		FromAddress: "noreply@example.com",
		ToAddresses: []string{"admin@example.com"},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.NotificationID != "new-email-id" {
		t.Errorf("expected notification ID new-email-id, got %s", result.NotificationID)
	}
}

func TestUpdateSlackNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.updateSlack" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateSlackNotification(context.Background(), UpdateSlackNotificationRequest{
		NotificationID: "notif-1",
		CreateSlackNotificationRequest: CreateSlackNotificationRequest{
			NotificationBase: NotificationBase{Name: "Updated Slack"},
			WebhookURL:       "https://hooks.slack.com/...",
			Channel:          "#new-channel",
		},
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateDiscordNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.updateDiscord" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateDiscordNotification(context.Background(), UpdateDiscordNotificationRequest{
		NotificationID: "notif-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateTelegramNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.updateTelegram" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateTelegramNotification(context.Background(), UpdateTelegramNotificationRequest{
		NotificationID: "notif-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUpdateEmailNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.updateEmail" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateEmailNotification(context.Background(), UpdateEmailNotificationRequest{
		NotificationID: "notif-1",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteNotification_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/notification.remove" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteNotification(context.Background(), "notif-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Organization Tests
// =============================================================================

func TestCreateOrganization_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/organization.create" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		var req CreateOrganizationRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if req.Name != "New Org" {
			t.Errorf("expected name 'New Org', got %s", req.Name)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(CreateOrganizationResponse{OrganizationID: "new-org-id"})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.CreateOrganization(context.Background(), CreateOrganizationRequest{
		Name: "New Org",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OrganizationID != "new-org-id" {
		t.Errorf("expected organization ID new-org-id, got %s", result.OrganizationID)
	}
}

func TestGetOrganization_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/organization.one" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(Organization{
			OrganizationID: "org-123",
			Name:           "Test Org",
		})
	})
	defer server.Close()

	client := newTestClient(server)
	result, err := client.GetOrganization(context.Background(), "org-123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.OrganizationID != "org-123" {
		t.Errorf("expected organization ID org-123, got %s", result.OrganizationID)
	}
}

func TestUpdateOrganization_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/organization.update" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.UpdateOrganization(context.Background(), UpdateOrganizationRequest{
		OrganizationID: "org-1",
		Name:           "Updated Org",
	})

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteOrganization_Success(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/organization.delete" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
	})
	defer server.Close()

	client := newTestClient(server)
	err := client.DeleteOrganization(context.Background(), "org-to-delete")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// =============================================================================
// Error Cases
// =============================================================================

func TestCreateBackup_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid schedule"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateBackup(context.Background(), CreateBackupRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSchedule_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid cron expression"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateSchedule(context.Background(), CreateScheduleRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateSlackNotification_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid webhook URL"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateSlackNotification(context.Background(), CreateSlackNotificationRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestCreateOrganization_Error(t *testing.T) {
	server := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "name already exists"}`))
	})
	defer server.Close()

	client := newTestClient(server)
	_, err := client.CreateOrganization(context.Background(), CreateOrganizationRequest{})

	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
