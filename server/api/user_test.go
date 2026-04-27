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
	"github.com/stretchr/testify/require"
)

func TestCreateUser_HappyPath(t *testing.T) {
	conn, terminate, connStr, err := testutil.BootTestDB()
	require.NoError(t, err, "Failed to boot test db")
	defer terminate()
	err = testutil.MigrateTestDB(connStr)
	require.NoError(t, err, "Failed to migrate test db")

	q := db.New(conn)
	server := api.NewServer(q)

	payload := map[string]string{"display_name": "test"}
	b, _ := json.Marshal(payload)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/api/user/profile", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")

	server.CreateUser(w, req)

	require.Equal(t, http.StatusOK, w.Code, "Expected status code 200")

	var resp struct {
		Data  *api.UserProfileResponse `json:"data"`
		Error *api.APIError            `json:"error"`
	}

	err = json.NewDecoder(w.Body).Decode(&resp)
	require.NoError(t, err)
	require.Equal(t, "test", resp.Data.DisplayName)
	require.Equal(t, int32(10), resp.Data.DailyWordLimit)
	require.Nil(t, resp.Error)

	var name string
	err = conn.QueryRow("SELECT display_name FROM users WHERE display_name = $1", "test").Scan(&name)
	require.NoError(t, err)
	require.Equal(t, "test", name)
}
