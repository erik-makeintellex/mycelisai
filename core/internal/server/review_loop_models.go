package server

import (
	"sort"
	"sync"
)

type LoopProfileType string
type LoopOwnerType string
type ReviewLoopEventKind string

const (
	LoopProfileTypeReview LoopProfileType = "review"

	LoopOwnerTypeTeam      LoopOwnerType = "team"
	LoopOwnerTypeAgentType LoopOwnerType = "agent_type"

	ReviewLoopEventOrganizationCreated         ReviewLoopEventKind = "organization_created"
	ReviewLoopEventTeamLeadActionCompleted     ReviewLoopEventKind = "team_lead_action_completed"
	ReviewLoopEventOrganizationAIEngineChanged ReviewLoopEventKind = "organization_ai_engine_changed"
	ReviewLoopEventResponseContractChanged     ReviewLoopEventKind = "response_contract_changed"

	DefaultDepartmentReviewLoopID = "department-readiness-review"
	DefaultAgentTypeReviewLoopID  = "agent-type-readiness-review"
)

type LoopOwnerRef struct {
	Type LoopOwnerType `json:"type"`
	ID   string        `json:"id"`
}

type LoopProfile struct {
	ID              string                `json:"id"`
	Name            string                `json:"name"`
	Type            LoopProfileType       `json:"type"`
	Description     string                `json:"description,omitempty"`
	Owner           LoopOwnerRef          `json:"owner"`
	IntervalSeconds int                   `json:"interval_seconds,omitempty"`
	EventTriggers   []ReviewLoopEventKind `json:"event_triggers,omitempty"`
}

type LoopOwnerResolution struct {
	Type      LoopOwnerType `json:"type"`
	ID        string        `json:"id"`
	Name      string        `json:"name"`
	HelpsWith string        `json:"helps_with"`
}

type ReviewLoopStructuredOutput struct {
	Status       string   `json:"status"`
	FlaggedItems int      `json:"flagged_items,omitempty"`
	Findings     []string `json:"findings"`
	Suggestions  []string `json:"suggestions"`
}

type LoopActivityStatus string

const (
	LoopActivityStatusSuccess LoopActivityStatus = "success"
	LoopActivityStatusWarning LoopActivityStatus = "warning"
	LoopActivityStatusFailed  LoopActivityStatus = "failed"
)

type ReviewLoopResult struct {
	ID               string                     `json:"id"`
	LoopID           string                     `json:"loop_id"`
	LoopName         string                     `json:"loop_name"`
	LoopType         LoopProfileType            `json:"loop_type"`
	OrganizationID   string                     `json:"organization_id"`
	OrganizationName string                     `json:"organization_name"`
	Trigger          string                     `json:"trigger"`
	Owner            LoopOwnerResolution        `json:"owner"`
	ActivityStatus   LoopActivityStatus         `json:"activity_status"`
	ActivitySummary  string                     `json:"activity_summary"`
	Review           ReviewLoopStructuredOutput `json:"review"`
	ReviewedAt       string                     `json:"reviewed_at"`
}

type LoopActivityItem struct {
	ID        string             `json:"id"`
	Name      string             `json:"name"`
	LastRunAt string             `json:"last_run_at"`
	Status    LoopActivityStatus `json:"status"`
	Summary   string             `json:"summary"`
}

type AutomationTriggerType string

const (
	AutomationTriggerTypeScheduled   AutomationTriggerType = "scheduled"
	AutomationTriggerTypeEventDriven AutomationTriggerType = "event_driven"
)

type AutomationOutcomeItem struct {
	Summary    string `json:"summary"`
	OccurredAt string `json:"occurred_at"`
}

type OrganizationAutomationItem struct {
	ID             string                  `json:"id"`
	Name           string                  `json:"name"`
	Purpose        string                  `json:"purpose"`
	TriggerType    AutomationTriggerType   `json:"trigger_type"`
	OwnerLabel     string                  `json:"owner_label"`
	Status         LoopActivityStatus      `json:"status"`
	Watches        string                  `json:"watches"`
	TriggerSummary string                  `json:"trigger_summary"`
	RecentOutcomes []AutomationOutcomeItem `json:"recent_outcomes,omitempty"`
}

type LearningInsightStrength string

const (
	LearningInsightStrengthEmerging   LearningInsightStrength = "emerging"
	LearningInsightStrengthConsistent LearningInsightStrength = "consistent"
	LearningInsightStrengthStrong     LearningInsightStrength = "strong"
)

type OrganizationLearningInsightItem struct {
	ID         string                  `json:"id"`
	Summary    string                  `json:"summary"`
	Source     string                  `json:"source"`
	ObservedAt string                  `json:"observed_at"`
	Strength   LearningInsightStrength `json:"strength"`
}

type LoopExecutionTracker struct {
	mu         sync.Mutex
	inProgress map[string]bool
}

type loopEventDispatchStats struct {
	executed    int
	failed      int
	skippedBusy int
}

func NewLoopExecutionTracker() *LoopExecutionTracker {
	return &LoopExecutionTracker{inProgress: make(map[string]bool)}
}

func (t *LoopExecutionTracker) TryStart(key string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.inProgress[key] {
		return false
	}
	t.inProgress[key] = true
	return true
}

func (t *LoopExecutionTracker) Finish(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()
	delete(t.inProgress, key)
}

type LoopProfileStore struct {
	mu    sync.RWMutex
	items map[string]map[string]LoopProfile
}

func NewLoopProfileStore() *LoopProfileStore {
	return &LoopProfileStore{items: make(map[string]map[string]LoopProfile)}
}

func (s *LoopProfileStore) Save(orgID string, profile LoopProfile) LoopProfile {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.items[orgID] == nil {
		s.items[orgID] = make(map[string]LoopProfile)
	}
	s.items[orgID][profile.ID] = profile
	return profile
}

func (s *LoopProfileStore) Get(orgID, loopID string) (LoopProfile, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profiles := s.items[orgID]
	if profiles == nil {
		return LoopProfile{}, false
	}
	profile, ok := profiles[loopID]
	return profile, ok
}

func (s *LoopProfileStore) ListByOrganization(orgID string) []LoopProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()

	profiles := s.items[orgID]
	if len(profiles) == 0 {
		return nil
	}

	items := make([]LoopProfile, 0, len(profiles))
	for _, profile := range profiles {
		items = append(items, profile)
	}
	sort.Slice(items, func(i, j int) bool {
		return items[i].ID < items[j].ID
	})
	return items
}

func (s *LoopProfileStore) EnsureDefaults(home OrganizationHomePayload) {
	home = normalizeOrganizationHome(home)

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.items[home.ID]) > 0 {
		return
	}

	defaults := defaultReviewLoopProfiles(home)
	if len(defaults) == 0 {
		return
	}

	s.items[home.ID] = make(map[string]LoopProfile, len(defaults))
	for _, profile := range defaults {
		s.items[home.ID][profile.ID] = profile
	}
}

type LoopResultStore struct {
	mu    sync.RWMutex
	items map[string][]ReviewLoopResult
}

func NewLoopResultStore() *LoopResultStore {
	return &LoopResultStore{items: make(map[string][]ReviewLoopResult)}
}

func (s *LoopResultStore) Add(orgID string, result ReviewLoopResult) ReviewLoopResult {
	s.mu.Lock()
	defer s.mu.Unlock()

	results := append([]ReviewLoopResult{result}, s.items[orgID]...)
	if len(results) > 20 {
		results = results[:20]
	}
	s.items[orgID] = results
	return result
}

func (s *LoopResultStore) List(orgID string) []ReviewLoopResult {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := append([]ReviewLoopResult(nil), s.items[orgID]...)
	sort.Slice(results, func(i, j int) bool {
		return results[i].ReviewedAt > results[j].ReviewedAt
	})
	return results
}

func (s *AdminServer) loopProfileStore() *LoopProfileStore {
	if s.LoopProfiles == nil {
		s.LoopProfiles = NewLoopProfileStore()
	}
	return s.LoopProfiles
}

func (s *AdminServer) loopResultStore() *LoopResultStore {
	if s.LoopResults == nil {
		s.LoopResults = NewLoopResultStore()
	}
	return s.LoopResults
}

func (s *AdminServer) loopExecutionTracker() *LoopExecutionTracker {
	if s.LoopExecution == nil {
		s.LoopExecution = NewLoopExecutionTracker()
	}
	return s.LoopExecution
}
