package ports

import (
	"backend-challenge/internal/core/domain"
	"context"
)

//go:generate mockgen -source service.go -destination mocks/mock_service.go -package mocks

type UserService interface {
	Authenticate(ctx context.Context, email, password string) (domain.User, error)
	Create(ctx context.Context, input domain.CreateUser) (domain.User, error)
	Get(ctx context.Context, id string) (domain.User, error)
	Update(ctx context.Context, id string, input domain.UpdateUser) (domain.User, error)
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, page, limit int) ([]domain.User, int64, error)
	Count(ctx context.Context) (int64, error)
}

type LotteryService interface {
	Search(ctx context.Context, pattern string, limit int) ([]domain.LotteryTicket, int64, error)
}
