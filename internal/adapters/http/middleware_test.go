package http

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-challenge/internal/core/ports/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestAuthMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		authHeader     string
		mockToken      func(m *mocks.MockTokenProvider)
		expectedStatus int
		expectedUserID string
	}{
		{
			name:       "Success",
			authHeader: "Bearer valid-token",
			mockToken: func(m *mocks.MockTokenProvider) {
				m.EXPECT().Parse("valid-token").Return("user-123", nil)
			},
			expectedStatus: http.StatusOK,
			expectedUserID: "user-123",
		},
		{
			name:           "Missing Token",
			authHeader:     "",
			mockToken:      func(m *mocks.MockTokenProvider) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "Invalid Format - No Bearer",
			authHeader:     "valid-token",
			mockToken:      func(m *mocks.MockTokenProvider) {},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:       "Invalid Token",
			authHeader: "Bearer invalid-token",
			mockToken: func(m *mocks.MockTokenProvider) {
				m.EXPECT().Parse("invalid-token").Return("", assert.AnError)
			},
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockTokenProvider := mocks.NewMockTokenProvider(ctrl)
			tt.mockToken(mockTokenProvider)

			router := gin.New()
			router.Use(AuthMiddleware(mockTokenProvider))
			router.GET("/test", func(c *gin.Context) {
				uid, exists := c.Get(userIDKey)
				if exists {
					c.String(http.StatusOK, uid.(string))
				} else {
					c.Status(http.StatusOK)
				}
			})

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)
			if tt.expectedUserID != "" {
				assert.Equal(t, tt.expectedUserID, w.Body.String())
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.Use(LoggingMiddleware())
	router.GET("/test", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
