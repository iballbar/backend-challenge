package services

import (
	"context"

	"backend-challenge/internal/core/domain"
	"backend-challenge/internal/core/ports"
)

type LotteryService struct {
	repo ports.LotteryRepository
}

func NewLotteryService(repo ports.LotteryRepository) *LotteryService {
	return &LotteryService{
		repo: repo,
	}
}

func (s *LotteryService) Search(ctx context.Context, pattern string, limit int) ([]domain.LotteryTicket, int64, error) {
	if limit <= 0 {
		limit = 1
	}
	results, remaining, err := s.repo.Search(ctx, pattern, limit)
	if err != nil {
		return nil, 0, err
	}
	if len(results) == 0 {
		return nil, 0, domain.ErrNoTicketsAvailable
	}
	return results, remaining, nil
}
