package job

import (
	"context"
	"log"
	"time"
)

func RunTTL(ctx context.Context, ttlExecutorFunction func(), tickDuration time.Duration) {
	ticker := time.NewTicker(tickDuration)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			ttlExecutorFunction()
		case <-ctx.Done():
			log.Print("Closing the job")
			return
		}
	}
}
