package controller

import (
	"context"
	"time"
)

// RunDispatcher processes durable commands until the context is cancelled.
// Polling is only a wake mechanism; the journal remains the source of truth.
func (service *Service) RunDispatcher(ctx context.Context, owner string, interval time.Duration) {
	if service.Executor == nil {
		return
	}
	if interval <= 0 {
		interval = 100 * time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		processed, _ := service.DispatchOnce(ctx, owner)
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}
