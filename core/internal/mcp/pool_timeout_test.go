package mcp

import (
	"context"
	"testing"
	"time"
)

func TestWithMCPConnectTimeoutAddsDeadline(t *testing.T) {
	start := time.Now()
	called := false

	err := withMCPConnectTimeout(context.Background(), func(ctx context.Context) error {
		called = true
		deadline, ok := ctx.Deadline()
		if !ok {
			t.Fatal("expected connect context deadline")
		}
		remaining := time.Until(deadline)
		if remaining <= 0 {
			t.Fatalf("expected positive remaining timeout, got %v", remaining)
		}
		if remaining > mcpConnectTimeout || remaining < mcpConnectTimeout-2*time.Second {
			t.Fatalf("remaining timeout = %v, want close to %v", remaining, mcpConnectTimeout)
		}
		if deadline.Sub(start) > mcpConnectTimeout+2*time.Second {
			t.Fatalf("deadline too far in future: %v", deadline.Sub(start))
		}
		return nil
	})

	if err != nil {
		t.Fatalf("withMCPConnectTimeout: %v", err)
	}
	if !called {
		t.Fatal("expected callback to run")
	}
}
