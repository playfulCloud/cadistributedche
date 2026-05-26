package jobs

import (
	"context"
	"log/slog"
	"time"
)

func RunTTL(ctx context.Context, ttlExecutorFunction func() int, tickDuration time.Duration) {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	slog.Info("ttl cleanup worker started", "interval", tickDuration.String())

	for {
		select {
		case <-ticker.C:
			ttlExecutorFunction()
		case <-ctx.Done():
			slog.Info("ttl cleanup worker stopped", "reason", ctx.Err())
			return
		}
	}
}
