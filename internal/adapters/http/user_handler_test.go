package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-challenge/internal/adapters/http/dto"
	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports/mocks"

	"github.com/gin-gonic/gin"
	"go.uber.org/mock/gomock"
)

func TestUserHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		payload    interface{}
		mock       func(m *mocks.MockUserService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			payload: dto.RegisterRequest{
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "password123",
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.User{ID: "1", Name: "Alice", Email: "alice@example.com"}, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   `{"id":"1","name":"Alice","email":"alice@example.com"`,
		},
		{
			name:       "Invalid JSON",
			payload:    "invalid",
			mock:       func(m *mocks.MockUserService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"invalid request payload"}`,
		},
		{
			name: "Email Conflict",
			payload: dto.RegisterRequest{
				Name:     "Alice",
				Email:    "exists@example.com",
				Password: "password123",
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrEmailAlreadyExists)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `{"error":"email already exists"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tt.mock(userService)
			handler := NewUserHandler(userService, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(jsonBytes))

			handler.Register(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantBody)) {
				t.Errorf("expected body to contain %s, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		payload    dto.LoginRequest
		mock       func(m *mocks.MockUserService, t *mocks.MockTokenProvider)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			payload: dto.LoginRequest{
				Email:    "alice@example.com",
				Password: "password123",
			},
			mock: func(m *mocks.MockUserService, tp *mocks.MockTokenProvider) {
				m.EXPECT().Authenticate(gomock.Any(), "alice@example.com", "password123").Return(domain.User{ID: "1"}, nil)
				tp.EXPECT().Create("1").Return("token123", nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `{"token":"token123"}`,
		},
		{
			name: "Invalid Credentials",
			payload: dto.LoginRequest{
				Email:    "alice@example.com",
				Password: "wrong",
			},
			mock: func(m *mocks.MockUserService, tp *mocks.MockTokenProvider) {
				m.EXPECT().Authenticate(gomock.Any(), "alice@example.com", "wrong").Return(domain.User{}, domain.ErrInvalidCredentials)
			},
			wantStatus: http.StatusUnauthorized,
			wantBody:   `{"error":"invalid credentials"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tokenProvider := mocks.NewMockTokenProvider(ctrl)
			tt.mock(userService, tokenProvider)
			handler := NewUserHandler(userService, tokenProvider)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(jsonBytes))

			handler.Login(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantBody)) {
				t.Errorf("expected body to contain %s, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_GetUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		id         string
		mock       func(m *mocks.MockUserService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			id:   "1",
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Get(gomock.Any(), "1").Return(domain.User{ID: "1", Name: "Alice"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":"1","name":"Alice"`,
		},
		{
			name: "Not Found",
			id:   "99",
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Get(gomock.Any(), "99").Return(domain.User{}, domain.ErrNotFound)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"user not found"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tt.mock(userService)
			handler := NewUserHandler(userService, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}
			c.Request, _ = http.NewRequest(http.MethodGet, "/users/"+tt.id, nil)

			handler.GetUser(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantBody)) {
				t.Errorf("expected body to contain %s, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_CreateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		payload    dto.CreateUserRequest
		mock       func(m *mocks.MockUserService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			payload: dto.CreateUserRequest{
				Name:     "Bob",
				Email:    "bob@example.com",
				Password: "password123",
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.User{ID: "2", Name: "Bob", Email: "bob@example.com"}, nil)
			},
			wantStatus: http.StatusCreated,
			wantBody:   `"id":"2","name":"Bob","email":"bob@example.com"`,
		},
		{
			name: "Error",
			payload: dto.CreateUserRequest{
				Name:     "Bob",
				Email:    "bob@error.com",
				Password: "password123",
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrInvalidEmail)
			},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"invalid email"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tt.mock(userService)
			handler := NewUserHandler(userService, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(http.MethodPost, "/users", bytes.NewBuffer(jsonBytes))

			handler.CreateUser(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantBody)) {
				t.Errorf("expected body to contain %s, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_ListUsers(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		mock       func(m *mocks.MockUserService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().List(gomock.Any(), 1, 10).Return([]domain.User{{ID: "1", Name: "Alice"}}, int64(1), nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":"1","name":"Alice"`,
		},
		{
			name: "Error",
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().List(gomock.Any(), 1, 10).Return(nil, int64(0), errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"failed to list users"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tt.mock(userService)
			handler := NewUserHandler(userService, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/users", nil)

			handler.ListUsers(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantBody)) {
				t.Errorf("expected body to contain %s, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_UpdateUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		id         string
		payload    dto.UpdateUserRequest
		mock       func(m *mocks.MockUserService)
		wantStatus int
		wantBody   string
	}{
		{
			name: "Success",
			id:   "1",
			payload: dto.UpdateUserRequest{
				Name: func(s string) *string { return &s }("Bob"),
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Update(gomock.Any(), "1", gomock.Any()).Return(domain.User{ID: "1", Name: "Bob"}, nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `"id":"1","name":"Bob"`,
		},
		{
			name: "Error",
			id:   "1",
			payload: dto.UpdateUserRequest{
				Name: func(s string) *string { return &s }("Bob"),
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Update(gomock.Any(), "1", gomock.Any()).Return(domain.User{}, errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `{"error":"internal server error"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tt.mock(userService)
			handler := NewUserHandler(userService, nil)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Params = []gin.Param{{Key: "id", Value: tt.id}}

			jsonBytes, _ := json.Marshal(tt.payload)
			c.Request, _ = http.NewRequest(http.MethodPatch, "/users/"+tt.id, bytes.NewBuffer(jsonBytes))

			handler.UpdateUser(c)

			if w.Code != tt.wantStatus {
				t.Errorf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(tt.wantBody)) {
				t.Errorf("expected body to contain %s, got %s", tt.wantBody, w.Body.String())
			}
		})
	}
}

func TestUserHandler_DeleteUser(t *testing.T) {
	gin.SetMode(gin.TestMode)

	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	userService := mocks.NewMockUserService(ctrl)
	userService.EXPECT().Delete(gomock.Any(), "1").Return(nil)
	handler := NewUserHandler(userService, nil)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = []gin.Param{{Key: "id", Value: "1"}}
	c.Request, _ = http.NewRequest(http.MethodDelete, "/users/1", nil)

	handler.DeleteUser(c)

	if w.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", w.Code)
	}
	if !bytes.Contains(w.Body.Bytes(), []byte(`{"status":"deleted"}`)) {
		t.Errorf("unexpected body: %s", w.Body.String())
	}
}
