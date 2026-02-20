package swarm

import (
	"context"
	"fmt"
	"log"
	"sync/atomic"
	"time"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// TeamScheduler triggers a team on a configurable interval.
// Reuses the Archivist timer-loop pattern: ticker + context cancellation.
type TeamScheduler struct {
	teamID            string
	interval          time.Duration
	nc                *nats.Conn
	ctx               context.Context
	cancel            context.CancelFunc
	triggerInProgress atomic.Bool
}

// Start runs the scheduler loop. Triggers immediately on first run, then on each tick.
// Blocks until the context is cancelled.
func (ts *TeamScheduler) Start() {
	ticker := time.NewTicker(ts.interval)
	defer ticker.Stop()

	log.Printf("TeamScheduler [%s]: active (every %s)", ts.teamID, ts.interval)
	ts.trigger() // Immediate first run

	for {
		select {
		case <-ts.ctx.Done():
			log.Printf("TeamScheduler [%s]: stopped", ts.teamID)
			return
		case <-ticker.C:
			ts.trigger()
		}
	}
}

// Stop cancels the scheduler loop.
func (ts *TeamScheduler) Stop() {
	if ts.cancel != nil {
		ts.cancel()
	}
}

func (ts *TeamScheduler) trigger() {
	// Phase 0 safety: prevent overlapping triggers
	if !ts.triggerInProgress.CompareAndSwap(false, true) {
		log.Printf("TeamScheduler [%s]: skipping — previous trigger still in progress", ts.teamID)
		return
	}
	defer ts.triggerInProgress.Store(false)

	subject := fmt.Sprintf(protocol.TopicTeamInternalTrigger, ts.teamID)
	payload := fmt.Sprintf(`{"triggered_by":"scheduler","timestamp":"%s"}`, time.Now().Format(time.RFC3339))
	if err := ts.nc.Publish(subject, []byte(payload)); err != nil {
		log.Printf("TeamScheduler [%s]: trigger failed: %v", ts.teamID, err)
	} else {
		log.Printf("TeamScheduler [%s]: triggered", ts.teamID)
	}
}
