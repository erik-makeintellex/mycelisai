package swarm

import "context"

// NewTestSoma creates a Soma with pre-populated teams for testing.
// Teams are registered but not started (no NATS subscriptions, no agent goroutines).
// Use this in handler tests that need Soma.ListTeams() to return data.
func NewTestSoma(manifests []*TeamManifest) *Soma {
	ctx, cancel := context.WithCancel(context.Background())
	s := &Soma{
		id:     "test-soma",
		teams:  make(map[string]*Team),
		ctx:    ctx,
		cancel: cancel,
	}
	for _, m := range manifests {
		s.teams[m.ID] = &Team{Manifest: m}
	}
	return s
}
