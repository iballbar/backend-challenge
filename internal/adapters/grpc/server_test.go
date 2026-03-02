package grpc

import (
	"context"
	"testing"

	pb "backend-challenge/internal/adapters/grpc/proto"
	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports/mocks"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func TestServer_CreateUser(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.CreateUserRequest
		mock     func(m *mocks.MockUserService)
		want     *pb.UserResponse
		wantCode codes.Code
	}{
		{
			name: "Success",
			req: &pb.CreateUserRequest{
				Name:     "Alice",
				Email:    "alice@example.com",
				Password: "password123",
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Create(gomock.Any(), domain.CreateUser{
					Name:     "Alice",
					Email:    "alice@example.com",
					Password: "password123",
				}).Return(domain.User{ID: "1", Name: "Alice", Email: "alice@example.com"}, nil)
			},
			want: &pb.UserResponse{
				Id:    "1",
				Name:  "Alice",
				Email: "alice@example.com",
			},
		},
		{
			name: "Email Conflict",
			req: &pb.CreateUserRequest{
				Name:     "Alice",
				Email:    "exists@example.com",
				Password: "password123",
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrEmailAlreadyExists)
			},
			wantCode: codes.AlreadyExists,
		},
		{
			name: "Invalid Email",
			req: &pb.CreateUserRequest{
				Email: "invalid",
			},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(domain.User{}, domain.ErrInvalidEmail)
			},
			wantCode: codes.InvalidArgument,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tt.mock(userService)
			s := NewServer(userService)

			resp, err := s.CreateUser(context.Background(), tt.req)

			if tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}

func TestServer_GetUser(t *testing.T) {
	tests := []struct {
		name     string
		req      *pb.GetUserRequest
		mock     func(m *mocks.MockUserService)
		want     *pb.UserResponse
		wantCode codes.Code
	}{
		{
			name: "Success",
			req:  &pb.GetUserRequest{Id: "1"},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Get(gomock.Any(), "1").Return(domain.User{ID: "1", Name: "Alice", Email: "alice@example.com"}, nil)
			},
			want: &pb.UserResponse{
				Id:    "1",
				Name:  "Alice",
				Email: "alice@example.com",
			},
		},
		{
			name: "Not Found",
			req:  &pb.GetUserRequest{Id: "99"},
			mock: func(m *mocks.MockUserService) {
				m.EXPECT().Get(gomock.Any(), "99").Return(domain.User{}, domain.ErrNotFound)
			},
			wantCode: codes.NotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			userService := mocks.NewMockUserService(ctrl)
			tt.mock(userService)
			s := NewServer(userService)

			resp, err := s.GetUser(context.Background(), tt.req)

			if tt.wantCode != codes.OK {
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tt.wantCode, st.Code())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp)
			}
		})
	}
}
