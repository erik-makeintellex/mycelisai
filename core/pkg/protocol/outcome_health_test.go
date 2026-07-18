package protocol

import "testing"

func TestOutcomeHealthForProject(t *testing.T) {
	tests := []struct {
		name string
		item OutcomeProject
		want OutcomeHealthState
	}{
		{name: "active ownership is healthy", item: OutcomeProject{Status: OutcomeProjectStatusActive}, want: OutcomeHealthHealthy},
		{name: "needs attention is degraded", item: OutcomeProject{Status: OutcomeProjectStatusNeedsAttention}, want: OutcomeHealthDegraded},
		{name: "output ready is completed", item: OutcomeProject{Status: OutcomeProjectStatusOutputReady}, want: OutcomeHealthCompleted},
		{name: "output refs complete active outcome", item: OutcomeProject{Status: OutcomeProjectStatusActive, OutputRefs: []TeamOutputRef{{OutputID: "out-1"}}}, want: OutcomeHealthCompleted},
		{name: "recovery refs degrade output", item: OutcomeProject{Status: OutcomeProjectStatusOutputReady, RecoveryRefs: []string{"recovery-1"}}, want: OutcomeHealthDegraded},
		{name: "archived stays archived", item: OutcomeProject{Status: OutcomeProjectStatusArchived, RecoveryRefs: []string{"recovery-1"}}, want: OutcomeHealthArchived},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OutcomeHealthForProject(tt.item); got != tt.want {
				t.Fatalf("OutcomeHealthForProject() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutcomeHealthForTeamWork(t *testing.T) {
	tests := []struct {
		state    TeamWorkState
		operator bool
		recovery []string
		want     OutcomeHealthState
	}{
		{state: TeamWorkStateNew, want: OutcomeHealthWaiting},
		{state: TeamWorkStateBriefed, want: OutcomeHealthWaiting},
		{state: TeamWorkStateQueued, want: OutcomeHealthWaiting},
		{state: TeamWorkStatePaused, want: OutcomeHealthWaiting},
		{state: TeamWorkStateRunning, want: OutcomeHealthRunning},
		{state: TeamWorkStateReviewing, want: OutcomeHealthRunning},
		{state: TeamWorkStateOutputReady, want: OutcomeHealthCompleted},
		{state: TeamWorkStateDegraded, want: OutcomeHealthDegraded},
		{state: TeamWorkStateNeedsOperator, want: OutcomeHealthBlocked},
		{state: TeamWorkStateArchived, want: OutcomeHealthArchived},
		{state: TeamWorkStateRunning, operator: true, want: OutcomeHealthBlocked},
		{state: TeamWorkStateRunning, recovery: []string{"retry"}, want: OutcomeHealthBlocked},
	}

	for _, tt := range tests {
		item := TeamWorkItem{State: tt.state, NeedsOperator: tt.operator, RecoveryOptions: tt.recovery}
		if got := OutcomeHealthForTeamWork(item); got != tt.want {
			t.Fatalf("OutcomeHealthForTeamWork(%q) = %q, want %q", tt.state, got, tt.want)
		}
	}
}

func TestOutcomeHealthForTeamStatusEvent(t *testing.T) {
	tests := []struct {
		name string
		item TeamStatusEvent
		want OutcomeHealthState
	}{
		{name: "running event is running", item: TeamStatusEvent{State: TeamWorkStateRunning}, want: OutcomeHealthRunning},
		{name: "degraded event stays degraded", item: TeamStatusEvent{State: TeamWorkStateDegraded}, want: OutcomeHealthDegraded},
		{name: "blocked dependency requires recovery", item: TeamStatusEvent{State: TeamWorkStateRunning, BlockedBy: []string{"provider"}}, want: OutcomeHealthBlocked},
		{name: "output event is completed", item: TeamStatusEvent{State: TeamWorkStateOutputReady}, want: OutcomeHealthCompleted},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := OutcomeHealthForTeamStatusEvent(tt.item); got != tt.want {
				t.Fatalf("OutcomeHealthForTeamStatusEvent() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestOutcomeHealthForRunStatus(t *testing.T) {
	tests := map[string]OutcomeHealthState{
		"pending":   OutcomeHealthWaiting,
		"running":   OutcomeHealthRunning,
		"completed": OutcomeHealthCompleted,
		"failed":    OutcomeHealthBlocked,
		"unknown":   OutcomeHealthHealthy,
	}
	for status, want := range tests {
		if got := OutcomeHealthForRunStatus(status); got != want {
			t.Fatalf("OutcomeHealthForRunStatus(%q) = %q, want %q", status, got, want)
		}
	}
}
