package http

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestLotteryHandler_SearchLottery(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		query      string
		mock       func(m *mocks.MockLotteryService)
		wantStatus int
		wantBody   string
	}{
		{
			name:  "Success",
			query: "?pattern=123456&limit=10",
			mock: func(m *mocks.MockLotteryService) {
				m.EXPECT().Search(gomock.Any(), "123456", 10).Return([]domain.LotteryTicket{
					{ID: "1", Number: "123456", Set: 1},
				}, int64(0), nil)
			},
			wantStatus: http.StatusOK,
			wantBody:   `"items":[{"id":"1","number":"123456","set":1}]`,
		},
		{
			name:       "Invalid Parameters - Pattern Length",
			query:      "?pattern=123&limit=10",
			mock:       func(m *mocks.MockLotteryService) {},
			wantStatus: http.StatusBadRequest,
			wantBody:   `"error":"validation failed"`,
		},
		{
			name:  "No Tickets Available",
			query: "?pattern=000000&limit=10",
			mock: func(m *mocks.MockLotteryService) {
				m.EXPECT().Search(gomock.Any(), "000000", 10).Return(nil, int64(0), domain.ErrNoTicketsAvailable)
			},
			wantStatus: http.StatusNotFound,
			wantBody:   `"error":"no tickets available"`,
		},
		{
			name:  "Internal Server Error",
			query: "?pattern=111111&limit=10",
			mock: func(m *mocks.MockLotteryService) {
				m.EXPECT().Search(gomock.Any(), "111111", 10).Return(nil, int64(0), errors.New("db error"))
			},
			wantStatus: http.StatusInternalServerError,
			wantBody:   `"error":"internal server error"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockService := mocks.NewMockLotteryService(ctrl)
			tt.mock(mockService)
			handler := NewLotteryHandler(mockService)

			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest(http.MethodGet, "/lottery/search"+tt.query, nil)

			handler.SearchLottery(c)

			assert.Equal(t, tt.wantStatus, w.Code)
			assert.Contains(t, w.Body.String(), tt.wantBody)
		})
	}
}
