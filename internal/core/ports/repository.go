package ports

import (
	"context"

	"backend-challenge/internal/core/domain"
)

//go:generate mockgen -source=repository.go -destination=mocks/mock_repository.go -package=mocks

type UserRepository interface {
	Create(ctx context.Context, input domain.CreateUser) (domain.User, error)
	GetByID(ctx context.Context, id string) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	List(ctx context.Context, page, limit int) ([]domain.User, int64, error)
	Update(ctx context.Context, id string, input domain.UpdateUser) (domain.User, error)
	Delete(ctx context.Context, id string) error
	Count(ctx context.Context) (int64, error)
}

type LotteryRepository interface {
	Search(ctx context.Context, pattern string, limit int) ([]domain.LotteryTicket, int64, error)
}
