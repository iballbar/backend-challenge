package grpc

import (
	"context"

	pb "backend-challenge/internal/adapters/grpc/proto"
	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	pb.UnimplementedUserServiceServer
	users ports.UserService
}

func NewServer(users ports.UserService) *Server {
	return &Server{users: users}
}

func (s *Server) CreateUser(ctx context.Context, req *pb.CreateUserRequest) (*pb.UserResponse, error) {
	created, err := s.users.Create(ctx, domain.CreateUser{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.UserResponse{
		Id:    created.ID,
		Name:  created.Name,
		Email: created.Email,
	}, nil
}

func (s *Server) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.UserResponse, error) {
	u, err := s.users.Get(ctx, req.Id)
	if err != nil {
		return nil, mapError(err)
	}

	return &pb.UserResponse{
		Id:    u.ID,
		Name:  u.Name,
		Email: u.Email,
	}, nil
}

func mapError(err error) error {
	switch err {
	case domain.ErrNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrEmailAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrInvalidEmail:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
