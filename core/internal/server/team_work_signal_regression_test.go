package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestProjectedSignalWouldRegressTerminalWork(t *testing.T) {
	tests := []struct {
		name     string
		current  protocol.TeamWorkState
		incoming protocol.TeamWorkState
		want     bool
	}{
		{name: "duplicate result", current: protocol.TeamWorkStateOutputReady, incoming: protocol.TeamWorkStateOutputReady, want: true},
		{name: "late running", current: protocol.TeamWorkStateOutputReady, incoming: protocol.TeamWorkStateRunning, want: true},
		{name: "blocked does not silently resume", current: protocol.TeamWorkStateNeedsOperator, incoming: protocol.TeamWorkStateRunning, want: true},
		{name: "validation does not silently resume", current: protocol.TeamWorkStateReviewing, incoming: protocol.TeamWorkStateRunning, want: true},
		{name: "validation may degrade", current: protocol.TeamWorkStateReviewing, incoming: protocol.TeamWorkStateDegraded, want: false},
		{name: "degraded may accept corrected result", current: protocol.TeamWorkStateDegraded, incoming: protocol.TeamWorkStateOutputReady, want: false},
		{name: "normal progress", current: protocol.TeamWorkStateQueued, incoming: protocol.TeamWorkStateRunning, want: false},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := projectedSignalWouldRegress(tc.current, tc.incoming); got != tc.want {
				t.Fatalf("projectedSignalWouldRegress(%q, %q) = %v, want %v", tc.current, tc.incoming, got, tc.want)
			}
		})
	}
}
