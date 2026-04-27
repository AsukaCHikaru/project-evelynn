package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/asukachikaru/project-evelynn/server/api"
	"github.com/asukachikaru/project-evelynn/server/db"
	"github.com/asukachikaru/project-evelynn/server/internal/testutil"
)

func TestCreateUser_HappyPath(t *testing.T) {
	conn, terminate, connStr, err := testutil.BootTestDB()
	if err != nil {
		t.Fatalf("Failed to boot test db: %v", err)
	}
	defer terminate()
	err = testutil.MigrateTestDB(connStr)
	if err != nil {
		t.Fatalf("Failed to migrate test db: %v", err)
	}

	q := db.New(conn)
	server := api.NewServer(q)

	payload := map[string]string{"display_name": "test"}
	b, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/user/profile", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	server.CreateUser(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status code 200, got %d", w.Code)
	}

	var resp struct {
		Data  *api.UserProfileResponse `json:"data"`
		Error *api.APIError            `json:"error"`
	}

	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response body: %v", err)
	}

	if resp.Data.DisplayName != "test" {
		t.Fatalf("Expected display name 'test', got '%s'", resp.Data.DisplayName)
	}
	if resp.Data.DailyWordLimit != 10 {
		t.Fatalf("Expected daily word limit 10, got %d", resp.Data.DailyWordLimit)
	}
	if resp.Error != nil {
		t.Fatalf("Expected no error, got %v", resp.Error)
	}

	var name string
	err = conn.QueryRow("SELECT display_name FROM users WHERE display_name = $1", "test").Scan(&name)
	if err != nil {
		t.Fatalf("Failed to query db: %v", err)
	}
	if name != "test" {
		t.Fatalf("Expected display name 'test', got '%s'", name)
	}
}
