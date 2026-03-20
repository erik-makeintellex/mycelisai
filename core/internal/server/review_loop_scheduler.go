package server

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"
)

const defaultLoopSchedulerTick = time.Second

var errScheduledLoopIntervalRequired = errors.New("interval_seconds must be greater than zero for scheduled execution")

type LoopScheduler struct {
	server       *AdminServer
	tickInterval time.Duration
	now          func() time.Time
	execute      func(OrganizationHomePayload, LoopProfile, string) (ReviewLoopResult, error)

	ctx    context.Context
	cancel context.CancelFunc

	mu          sync.Mutex
	lastAttempt map[string]time.Time
	inProgress  map[string]bool
}

type loopSchedulerTickStats struct {
	executed    int
	failed      int
	skippedBusy int
}

func NewLoopScheduler(server *AdminServer) *LoopScheduler {
	return &LoopScheduler{
		server:       server,
		tickInterval: defaultLoopSchedulerTick,
		now:          time.Now,
		execute: func(home OrganizationHomePayload, profile LoopProfile, trigger string) (ReviewLoopResult, error) {
			return server.executeReviewLoop(home, profile, trigger)
		},
		lastAttempt: make(map[string]time.Time),
		inProgress:  make(map[string]bool),
	}
}

func (s *AdminServer) loopScheduler() *LoopScheduler {
	if s.LoopScheduler == nil {
		s.LoopScheduler = NewLoopScheduler(s)
	}
	return s.LoopScheduler
}

func (s *AdminServer) StartLoopScheduler(parent context.Context) {
	scheduler := s.loopScheduler()
	if scheduler.isRunning() {
		return
	}
	scheduler.Start(parent)
}

func (s *AdminServer) StopLoopScheduler() {
	if s.LoopScheduler != nil {
		s.LoopScheduler.Stop()
	}
}

func (ls *LoopScheduler) Start(parent context.Context) {
	ls.mu.Lock()
	if ls.cancel != nil {
		ls.mu.Unlock()
		return
	}
	ls.ctx, ls.cancel = context.WithCancel(parent)
	ls.mu.Unlock()

	go ls.run()
}

func (ls *LoopScheduler) Stop() {
	ls.mu.Lock()
	cancel := ls.cancel
	ls.cancel = nil
	ls.ctx = nil
	ls.mu.Unlock()

	if cancel != nil {
		cancel()
	}
}

func (ls *LoopScheduler) isRunning() bool {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.cancel != nil
}

func (ls *LoopScheduler) run() {
	ticker := time.NewTicker(ls.tickInterval)
	defer ticker.Stop()

	log.Printf("[review-loop-scheduler] active (every %s)", ls.tickInterval)
	ls.runDueLoopsAt(ls.now())

	for {
		select {
		case <-ls.ctx.Done():
			log.Printf("[review-loop-scheduler] stopped")
			return
		case <-ticker.C:
			ls.runDueLoopsAt(ls.now())
		}
	}
}

func (ls *LoopScheduler) runDueLoopsAt(now time.Time) loopSchedulerTickStats {
	stats := loopSchedulerTickStats{}
	summaries := ls.server.organizationStore().List()
	for _, summary := range summaries {
		home, ok := ls.server.organizationStore().Get(summary.ID)
		if !ok {
			continue
		}
		ls.server.loopProfileStore().EnsureDefaults(home)
		for _, profile := range ls.server.loopProfileStore().ListByOrganization(home.ID) {
			if profile.IntervalSeconds == 0 {
				continue
			}
			executed, failed, skippedBusy := ls.tryExecuteScheduledLoop(home, profile, now)
			if executed {
				stats.executed++
			}
			if failed {
				stats.failed++
			}
			if skippedBusy {
				stats.skippedBusy++
			}
		}
	}

	if stats.executed > 0 || stats.failed > 0 || stats.skippedBusy > 0 {
		log.Printf("[review-loop-scheduler] tick executed=%d failed=%d skipped_busy=%d", stats.executed, stats.failed, stats.skippedBusy)
	}
	return stats
}

func (ls *LoopScheduler) tryExecuteScheduledLoop(home OrganizationHomePayload, profile LoopProfile, now time.Time) (executed bool, failed bool, skippedBusy bool) {
	if err := validateLoopProfile(profile); err != nil {
		ls.server.recordFailedLoopResult(home, profile, "scheduled", err)
		log.Printf("[review-loop-scheduler] organization=%s loop=%s invalid_config=%v", home.ID, profile.ID, err)
		return false, true, false
	}
	if profile.IntervalSeconds <= 0 {
		err := errScheduledLoopIntervalRequired
		ls.server.recordFailedLoopResult(home, profile, "scheduled", err)
		log.Printf("[review-loop-scheduler] organization=%s loop=%s invalid_config=%v", home.ID, profile.ID, err)
		return false, true, false
	}

	key := scheduledLoopExecutionKey(home.ID, profile.ID)

	ls.mu.Lock()
	if ls.inProgress[key] {
		ls.mu.Unlock()
		log.Printf("[review-loop-scheduler] organization=%s loop=%s skipped_previous_execution_still_running=true", home.ID, profile.ID)
		return false, false, true
	}
	if lastAttempt, ok := ls.lastAttempt[key]; ok && now.Sub(lastAttempt) < time.Duration(profile.IntervalSeconds)*time.Second {
		ls.mu.Unlock()
		return false, false, false
	}
	ls.inProgress[key] = true
	ls.lastAttempt[key] = now
	ls.mu.Unlock()

	startedAt := time.Now()
	defer func() {
		ls.mu.Lock()
		delete(ls.inProgress, key)
		ls.mu.Unlock()
	}()

	result, err := ls.execute(normalizeOrganizationHome(home), profile, "scheduled")
	if err != nil {
		ls.server.recordFailedLoopResult(home, profile, "scheduled", err)
		log.Printf("[review-loop-scheduler] organization=%s loop=%s failed duration=%s error=%v", home.ID, profile.ID, time.Since(startedAt), err)
		return false, true, false
	}

	log.Printf("[review-loop-scheduler] organization=%s loop=%s completed duration=%s status=%s result_id=%s", home.ID, profile.ID, time.Since(startedAt), result.Review.Status, result.ID)
	return true, false, false
}

func scheduledLoopExecutionKey(orgID, loopID string) string {
	return orgID + "::" + loopID
}
