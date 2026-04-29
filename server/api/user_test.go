package api_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/asukachikaru/project-evelynn/server/api"
	"github.com/asukachikaru/project-evelynn/server/db"
	"github.com/asukachikaru/project-evelynn/server/internal/testutil"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

var (
	conn      *sql.DB
	terminate func()
	server    *api.Server
	connStr   string
)

func TestMain(m *testing.M) {
	conn, terminate, connStr, _ = testutil.BootTestDB()
	testutil.MigrateTestDB(connStr)

	q := db.New(conn)
	server = api.NewServer(q)

	code := m.Run()

	terminate()
	os.Exit(code)
}

func TestUserProfile(t *testing.T) {
	t.Run("POST", func(t *testing.T) {
		t.Run("Happy path",
			func(t *testing.T) {
				testutil.TruncateTestDB(conn)
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

				err := json.NewDecoder(w.Body).Decode(&resp)
				require.NoError(t, err)
				require.Equal(t, "test", resp.Data.DisplayName)
				require.Equal(t, int32(10), resp.Data.DailyWordLimit)
				require.Nil(t, resp.Error)

				var name string
				err = conn.QueryRow("SELECT display_name FROM users WHERE display_name = $1", "test").Scan(&name)
				require.NoError(t, err)
				require.Equal(t, "test", name)
			})
		t.Run("Missing display name",
			func(t *testing.T) {
				testutil.TruncateTestDB(conn)
				payload := map[string]string{"display_name": ""}
				b, _ := json.Marshal(payload)
				w := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/api/user/profile", bytes.NewReader(b))
				req.Header.Set("Content-Type", "application/json")

				server.CreateUser(w, req)
				require.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

				var count int
				err := conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			})
		t.Run("Duplicate display name",
			func(t *testing.T) {
				testutil.TruncateTestDB(conn)
				_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
				require.NoError(t, err)

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
			})
		t.Run("Malformed JSON",
			func(t *testing.T) {
				testutil.TruncateTestDB(conn)
				invalidJSON := []byte(`{"display_name": "test"`) // Missing closing brace
				w := httptest.NewRecorder()
				req := httptest.NewRequest("POST", "/api/user/profile", bytes.NewReader(invalidJSON))
				req.Header.Set("Content-Type", "application/json")

				server.CreateUser(w, req)

				require.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

				var count int
				err := conn.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
				require.NoError(t, err)
				require.Equal(t, 0, count)
			})
	})
	t.Run("GET", func(t *testing.T) {
		t.Run("Happy path", func(t *testing.T) {
			testutil.TruncateTestDB(conn)
			_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
			require.NoError(t, err)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/profile", nil)

			server.GetUserProfile(w, req)

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
		})
		t.Run("No user found", func(t *testing.T) {
			testutil.TruncateTestDB(conn)

			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", "/api/user/profile", nil)

			server.GetUserProfile(w, req)

			require.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
		})
	})
	t.Run("PATCH", func(t *testing.T) {
		t.Run("Update display_name and daily_word_limit", func(t *testing.T) {
			testutil.TruncateTestDB(conn)
			_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
			require.NoError(t, err)

			payload := map[string]interface{}{"display_name": "updated_test", "daily_word_limit": 20}
			b, _ := json.Marshal(payload)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/api/user/profile", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			server.UpdateUser(w, req)

			require.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
			var resp struct {
				Data  *api.UserProfileResponse `json:"data"`
				Error *api.APIError            `json:"error"`
			}

			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, "updated_test", resp.Data.DisplayName)
			require.Equal(t, int32(20), resp.Data.DailyWordLimit)
			require.Nil(t, resp.Error)
		})
		t.Run("Update only display_name", func(t *testing.T) {
			testutil.TruncateTestDB(conn)
			_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
			require.NoError(t, err)

			payload := map[string]interface{}{"display_name": "updated_test"}
			b, _ := json.Marshal(payload)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/api/user/profile", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			server.UpdateUser(w, req)

			require.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
			var resp struct {
				Data  *api.UserProfileResponse `json:"data"`
				Error *api.APIError            `json:"error"`
			}

			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, "updated_test", resp.Data.DisplayName)
			require.Equal(t, int32(10), resp.Data.DailyWordLimit) // Unchanged
			require.Nil(t, resp.Error)
		})
		t.Run("Update only daily_word_limit", func(t *testing.T) {
			testutil.TruncateTestDB(conn)
			_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
			require.NoError(t, err)

			payload := map[string]interface{}{"daily_word_limit": 20}
			b, _ := json.Marshal(payload)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/api/user/profile", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			server.UpdateUser(w, req)

			require.Equal(t, http.StatusOK, w.Code, "Expected status code 200")
			var resp struct {
				Data  *api.UserProfileResponse `json:"data"`
				Error *api.APIError            `json:"error"`
			}

			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, "test", resp.Data.DisplayName) // Unchanged
			require.Equal(t, int32(20), resp.Data.DailyWordLimit)
			require.Nil(t, resp.Error)
		})
		t.Run("display_name is empty", func(t *testing.T) {
			testutil.TruncateTestDB(conn)
			_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
			require.NoError(t, err)

			payload := map[string]interface{}{"display_name": ""}
			b, _ := json.Marshal(payload)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/api/user/profile", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			server.UpdateUser(w, req)

			require.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

			var resp struct {
				Data  *api.UserProfileResponse `json:"data"`
				Error *api.APIError            `json:"error"`
			}

			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, "Display name cannot be empty", resp.Error.Message)

			var displayName string
			err = conn.QueryRow("SELECT display_name FROM users WHERE display_name = $1", "test").Scan(&displayName)
			require.NoError(t, err)
			require.Equal(t, "test", displayName) // Unchanged
		})
		t.Run("daily_word_limit is negative", func(t *testing.T) {
			testutil.TruncateTestDB(conn)
			_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
			require.NoError(t, err)

			payload := map[string]interface{}{"daily_word_limit": -1}
			b, _ := json.Marshal(payload)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/api/user/profile", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			server.UpdateUser(w, req)

			require.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")
			var resp struct {
				Data  *api.UserProfileResponse `json:"data"`
				Error *api.APIError            `json:"error"`
			}
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, "Daily word limit must be between 1 and 20", resp.Error.Message)
			var dailyWordLimit int32
			err = conn.QueryRow("SELECT daily_word_limit FROM users WHERE display_name = $1", "test").Scan(&dailyWordLimit)
			require.NoError(t, err)
			require.Equal(t, int32(10), dailyWordLimit) // Unchanged
		})
		t.Run("daily_word_limit is zero", func(t *testing.T) {
			testutil.TruncateTestDB(conn)
			_, err := conn.Exec("INSERT INTO users (display_name, user_hash_id) VALUES ($1, $2)", "test", uuid.New().String())
			require.NoError(t, err)

			payload := map[string]interface{}{"daily_word_limit": 0}
			b, _ := json.Marshal(payload)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/api/user/profile", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			server.UpdateUser(w, req)

			require.Equal(t, http.StatusBadRequest, w.Code, "Expected status code 400")

			var resp struct {
				Data  *api.UserProfileResponse `json:"data"`
				Error *api.APIError            `json:"error"`
			}
			err = json.NewDecoder(w.Body).Decode(&resp)
			require.NoError(t, err)
			require.Equal(t, "Daily word limit must be between 1 and 20", resp.Error.Message)
			var dailyWordLimit int32
			err = conn.QueryRow("SELECT daily_word_limit FROM users WHERE display_name = $1", "test").Scan(&dailyWordLimit)
			require.NoError(t, err)
			require.Equal(t, int32(10), dailyWordLimit) // Unchanged
		})
		t.Run("User does not exist", func(t *testing.T) {
			testutil.TruncateTestDB(conn)

			payload := map[string]interface{}{"display_name": "updated_test", "daily_word_limit": 20}
			b, _ := json.Marshal(payload)
			w := httptest.NewRecorder()
			req := httptest.NewRequest("PATCH", "/api/user/profile", bytes.NewReader(b))
			req.Header.Set("Content-Type", "application/json")

			server.UpdateUser(w, req)

			require.Equal(t, http.StatusNotFound, w.Code, "Expected status code 404")
		})
	})
}
