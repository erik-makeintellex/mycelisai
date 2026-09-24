package server

import (
	"context"
	"log/slog"
	"time"
)

// StartInvocationRecovery records uncertainty only. It never dispatches effects.
func StartInvocationRecovery(ctx context.Context, s *AdminServer) {
	if s == nil || s.Invocations == nil || s.getDB() == nil {
		return
	}
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			probe, cancel := context.WithTimeout(ctx, 4*time.Second)
			count, err := s.Invocations.RecoverExpired(probe)
			cancel()
			if err != nil && ctx.Err() == nil {
				slog.Error("invocation recovery unavailable", "component", "invocation_recovery")
			} else if count > 0 {
				slog.Info("expired effects require reconciliation", "component", "invocation_recovery", "count", count)
			}
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
			}
		}
	}()
}
