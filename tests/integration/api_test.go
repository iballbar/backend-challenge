package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"backend-challenge/internal/core/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type loginResponse struct {
	Token string `json:"token"`
}

func getAuthToken(t *testing.T, email, password string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"email":    email,
		"password": password,
	})
	req, _ := http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(body))
	w := httptest.NewRecorder()
	testRouter.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp loginResponse
	err := json.Unmarshal(w.Body.Bytes(), &resp)
	require.NoError(t, err)
	return resp.Token
}

func TestAPI_AuthAndUserFlow(t *testing.T) {
	if os.Getenv("INT") != "1" {
		t.Skip("skipping integration test")
	}
	truncateCollection(t, "users")

	var token string
	var createdUserID string

	t.Run("Auth Flow", func(t *testing.T) {
		tests := []struct {
			name          string
			method        string
			path          string
			body          interface{}
			expectedCode  int
			checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
		}{
			{
				name:   "Register User",
				method: http.MethodPost,
				path:   "/register",
				body: map[string]string{
					"name":     "API Test User",
					"email":    "api@example.com",
					"password": "password123",
				},
				expectedCode: http.StatusCreated,
			},
			{
				name:   "Login User",
				method: http.MethodPost,
				path:   "/login",
				body: map[string]string{
					"email":    "api@example.com",
					"password": "password123",
				},
				expectedCode: http.StatusOK,
				checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
					var resp loginResponse
					err := json.Unmarshal(w.Body.Bytes(), &resp)
					require.NoError(t, err)
					assert.NotEmpty(t, resp.Token)
					token = resp.Token
				},
			},
			{
				name:   "Login Invalid Credentials",
				method: http.MethodPost,
				path:   "/login",
				body: map[string]string{
					"email":    "api@example.com",
					"password": "wrongpassword",
				},
				expectedCode: http.StatusUnauthorized,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				body, _ := json.Marshal(tt.body)
				req, _ := http.NewRequest(tt.method, tt.path, bytes.NewBuffer(body))
				w := httptest.NewRecorder()
				testRouter.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedCode, w.Code)
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
			})
		}
	})

	t.Run("Protected User CRUD Flow", func(t *testing.T) {
		require.NotEmpty(t, token, "token is required for protected tests")

		tests := []struct {
			name          string
			method        string
			path          func() string
			body          interface{}
			expectedCode  int
			checkResponse func(t *testing.T, w *httptest.ResponseRecorder)
		}{
			{
				name:   "Create User (Protected)",
				method: http.MethodPost,
				path:   func() string { return "/users" },
				body: map[string]string{
					"name":     "New User",
					"email":    "new@example.com",
					"password": "password123",
				},
				expectedCode: http.StatusCreated,
				checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
					var resp domain.User
					err := json.Unmarshal(w.Body.Bytes(), &resp)
					require.NoError(t, err)
					assert.Equal(t, "New User", resp.Name)
					createdUserID = resp.ID
				},
			},
			{
				name:         "List Users (Protected)",
				method:       http.MethodGet,
				path:         func() string { return "/users" },
				expectedCode: http.StatusOK,
				checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
					var resp struct {
						Data struct {
							Users []domain.User `json:"users"`
							Total int64         `json:"total"`
						} `json:"data"`
					}
					err := json.Unmarshal(w.Body.Bytes(), &resp)
					require.NoError(t, err, "Response: "+w.Body.String())
					assert.GreaterOrEqual(t, resp.Data.Total, int64(1))
				},
			},
			{
				name:         "Get User By ID (Protected)",
				method:       http.MethodGet,
				path:         func() string { return "/users/" + createdUserID },
				expectedCode: http.StatusOK,
				checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
					var resp domain.User
					err := json.Unmarshal(w.Body.Bytes(), &resp)
					require.NoError(t, err)
					assert.Equal(t, createdUserID, resp.ID)
				},
			},
			{
				name:   "Update User (Protected)",
				method: http.MethodPatch,
				path:   func() string { return "/users/" + createdUserID },
				body: map[string]string{
					"name": "Updated Name",
				},
				expectedCode: http.StatusOK,
				checkResponse: func(t *testing.T, w *httptest.ResponseRecorder) {
					var resp domain.User
					err := json.Unmarshal(w.Body.Bytes(), &resp)
					require.NoError(t, err)
					assert.Equal(t, "Updated Name", resp.Name)
				},
			},
			{
				name:         "Delete User (Protected)",
				method:       http.MethodDelete,
				path:         func() string { return "/users/" + createdUserID },
				expectedCode: http.StatusOK,
			},
			{
				name:         "Get Deleted User (Protected)",
				method:       http.MethodGet,
				path:         func() string { return "/users/" + createdUserID },
				expectedCode: http.StatusNotFound,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				var bodyReader *bytes.Buffer
				if tt.body != nil {
					b, _ := json.Marshal(tt.body)
					bodyReader = bytes.NewBuffer(b)
				} else {
					bodyReader = bytes.NewBuffer(nil)
				}

				req, _ := http.NewRequest(tt.method, tt.path(), bodyReader)
				req.Header.Set("Authorization", "Bearer "+token)
				w := httptest.NewRecorder()
				testRouter.ServeHTTP(w, req)

				assert.Equal(t, tt.expectedCode, w.Code, w.Body.String())
				if tt.checkResponse != nil {
					tt.checkResponse(t, w)
				}
			})
		}
	})
}

func TestAPI_LotteryFlow(t *testing.T) {
	if os.Getenv("INT") != "1" {
		t.Skip("skipping integration test")
	}
	t.Run("Search Wildcard", func(t *testing.T) {
		truncateCollection(t, "lottery_tickets")
		tickets := []lotteryTicketDoc{
			makeTicket("111111", 1),
			makeTicket("222222", 1),
		}
		seedTickets(t, tickets)

		req, _ := http.NewRequest(http.MethodGet, "/lottery/search?pattern=******&limit=5", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data struct {
				Items []map[string]interface{} `json:"items"`
				Count int                      `json:"count"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		assert.Equal(t, 2, resp.Data.Count)
	})

	t.Run("Search Specific", func(t *testing.T) {
		truncateCollection(t, "lottery_tickets")
		tickets := []lotteryTicketDoc{
			makeTicket("123456", 1),
		}
		seedTickets(t, tickets)

		req, _ := http.NewRequest(http.MethodGet, "/lottery/search?pattern=123456&limit=5", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var resp struct {
			Data struct {
				Items []map[string]interface{} `json:"items"`
			} `json:"data"`
		}
		err := json.Unmarshal(w.Body.Bytes(), &resp)
		require.NoError(t, err)
		require.NotEmpty(t, resp.Data.Items)
		assert.Equal(t, "123456", resp.Data.Items[0]["number"])
	})

	t.Run("Search Invalid Pattern", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, "/lottery/search?pattern=123&limit=5", nil)
		w := httptest.NewRecorder()
		testRouter.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
