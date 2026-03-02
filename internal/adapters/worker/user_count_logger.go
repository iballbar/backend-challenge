package worker

import (
	"context"
	"log/slog"
	"time"

	"backend-challenge/internal/core/ports"
)

func StartUserCountLogger(ctx context.Context, repo ports.UserRepository, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			count, err := repo.Count(ctx)
			if err != nil {
				slog.Error("failed to count users", "error", err)
				continue
			}
			slog.Info("user count", "count", count)
		}
	}
}
