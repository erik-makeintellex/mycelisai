package server

import (
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
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

func defaultReviewLoopProfiles(home OrganizationHomePayload) []LoopProfile {
	home = normalizeOrganizationHome(home)
	profiles := make([]LoopProfile, 0, 2)

	if len(home.Departments) > 0 {
		firstDepartment := home.Departments[0]
		profiles = append(profiles, LoopProfile{
			ID:          DefaultDepartmentReviewLoopID,
			Name:        "Department readiness review",
			Type:        LoopProfileTypeReview,
			Description: "Reviews the current Department structure and operating readiness without taking action.",
			Owner: LoopOwnerRef{
				Type: LoopOwnerTypeTeam,
				ID:   firstDepartment.ID,
			},
			IntervalSeconds: 60,
			EventTriggers: []ReviewLoopEventKind{
				ReviewLoopEventOrganizationCreated,
				ReviewLoopEventTeamLeadActionCompleted,
				ReviewLoopEventOrganizationAIEngineChanged,
				ReviewLoopEventResponseContractChanged,
			},
		})

		if len(firstDepartment.AgentTypeProfiles) > 0 {
			firstProfile := firstDepartment.AgentTypeProfiles[0]
			profiles = append(profiles, LoopProfile{
				ID:          DefaultAgentTypeReviewLoopID,
				Name:        "Agent type readiness review",
				Type:        LoopProfileTypeReview,
				Description: "Reviews a specialist profile and its inherited defaults without taking action.",
				Owner: LoopOwnerRef{
					Type: LoopOwnerTypeAgentType,
					ID:   firstProfile.ID,
				},
				EventTriggers: []ReviewLoopEventKind{
					ReviewLoopEventOrganizationCreated,
					ReviewLoopEventOrganizationAIEngineChanged,
					ReviewLoopEventResponseContractChanged,
				},
			})
		}
	}

	return profiles
}

func resolveLoopOwner(home OrganizationHomePayload, profile LoopProfile) (LoopOwnerResolution, error) {
	switch profile.Owner.Type {
	case LoopOwnerTypeTeam:
		for _, department := range home.Departments {
			if department.ID == profile.Owner.ID {
				return LoopOwnerResolution{
					Type:      LoopOwnerTypeTeam,
					ID:        department.ID,
					Name:      department.Name,
					HelpsWith: fmt.Sprintf("%s organizes %d Specialist%s for this AI Organization.", department.Name, department.SpecialistCount, pluralSuffix(department.SpecialistCount)),
				}, nil
			}
		}
		return LoopOwnerResolution{}, fmt.Errorf("loop owner team not found")
	case LoopOwnerTypeAgentType:
		for _, department := range home.Departments {
			for _, agentType := range department.AgentTypeProfiles {
				if agentType.ID == profile.Owner.ID {
					return LoopOwnerResolution{
						Type:      LoopOwnerTypeAgentType,
						ID:        agentType.ID,
						Name:      agentType.Name,
						HelpsWith: agentType.HelpsWith,
					}, nil
				}
			}
		}
		return LoopOwnerResolution{}, fmt.Errorf("loop owner agent type not found")
	default:
		return LoopOwnerResolution{}, fmt.Errorf("loop owner type must be team or agent_type")
	}
}

func buildReviewLoopOutput(home OrganizationHomePayload, owner LoopOwnerResolution) ReviewLoopStructuredOutput {
	findings := []string{
		fmt.Sprintf("%s is currently operating with %d Department%s, %d Specialist%s, and %d Advisor%s.", safeOrganizationName(home.Name), home.DepartmentCount, pluralSuffix(home.DepartmentCount), home.SpecialistCount, pluralSuffix(home.SpecialistCount), home.AdvisorCount, pluralSuffix(home.AdvisorCount)),
		fmt.Sprintf("%s is the active review owner for this loop.", owner.Name),
	}
	suggestions := []string{
		"Keep this loop read-only until Automations visibility and policy surfaces are in place.",
	}

	status := "healthy"
	flaggedItems := 0
	if home.DepartmentCount == 0 {
		status = "attention_needed"
		flaggedItems++
		findings = append(findings, "No Departments are configured yet, so the Team Lead does not have a clear execution lane.")
		suggestions = append(suggestions, "Add at least one Department before broadening beyond Team Lead guidance.")
	} else {
		findings = append(findings, fmt.Sprintf("The active AI Engine default is %s.", strings.TrimSpace(home.AIEngineSettingsSummary)))
		suggestions = append(suggestions, "Use the review findings to confirm whether the current Department structure still matches the organization's first priorities.")
	}

	if home.SpecialistCount == 0 {
		status = "attention_needed"
		flaggedItems++
		findings = append(findings, "No Specialists are currently attached, so follow-through still depends on Team Lead planning alone.")
		suggestions = append(suggestions, "Add Specialists only after the Team Lead confirms the first execution lane.")
	}

	if strings.TrimSpace(home.ResponseContractSummary) == "" {
		status = "attention_needed"
		flaggedItems++
		findings = append(findings, "No Response Style is visible yet for this AI Organization.")
		suggestions = append(suggestions, "Set a Response Style so future guidance stays consistent and reviewable.")
	} else {
		findings = append(findings, fmt.Sprintf("The current Response Style is %s.", strings.TrimSpace(home.ResponseContractSummary)))
	}

	if owner.Type == LoopOwnerTypeAgentType {
		findings = append(findings, fmt.Sprintf("%s currently helps with: %s.", owner.Name, owner.HelpsWith))
		suggestions = append(suggestions, "Review whether this specialist profile should keep inheriting defaults or eventually receive a bounded override.")
	} else {
		suggestions = append(suggestions, "Review the Department summary before widening into live loop execution or external actions.")
	}

	return ReviewLoopStructuredOutput{
		Status:       status,
		FlaggedItems: flaggedItems,
		Findings:     findings,
		Suggestions:  suggestions,
	}
}

func (s *AdminServer) executeReviewLoop(home OrganizationHomePayload, profile LoopProfile, trigger string) (ReviewLoopResult, error) {
	if err := validateLoopProfile(profile); err != nil {
		return ReviewLoopResult{}, err
	}

	home = normalizeOrganizationHome(home)
	owner, err := resolveLoopOwner(home, profile)
	if err != nil {
		return ReviewLoopResult{}, err
	}

	result := ReviewLoopResult{
		ID:               uuid.NewString(),
		LoopID:           profile.ID,
		LoopName:         profile.Name,
		LoopType:         profile.Type,
		OrganizationID:   home.ID,
		OrganizationName: safeOrganizationName(home.Name),
		Trigger:          trigger,
		Owner:            owner,
		Review:           buildReviewLoopOutput(home, owner),
		ReviewedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	result.ActivityStatus, result.ActivitySummary = summarizeActivityFromReview(result.Review)

	s.loopResultStore().Add(home.ID, result)
	log.Printf("[review-loop] organization=%s loop=%s owner_type=%s owner_id=%s status=%s findings=%d suggestions=%d", home.ID, profile.ID, owner.Type, owner.ID, result.Review.Status, len(result.Review.Findings), len(result.Review.Suggestions))
	return result, nil
}

func summarizeActivityFromReview(review ReviewLoopStructuredOutput) (LoopActivityStatus, string) {
	if review.Status == "attention_needed" {
		flagged := review.FlaggedItems
		if flagged <= 0 {
			flagged = 1
		}
		return LoopActivityStatusWarning, fmt.Sprintf("%d item%s flagged", flagged, pluralSuffix(flagged))
	}
	return LoopActivityStatusSuccess, "No issues detected"
}

func summarizeActivityFromError(err error) (LoopActivityStatus, string) {
	if err == nil {
		return LoopActivityStatusFailed, "Review unavailable"
	}
	message := strings.TrimSpace(err.Error())
	switch {
	case strings.Contains(message, "owner"):
		return LoopActivityStatusFailed, "Review owner unavailable"
	case strings.Contains(message, "interval_seconds"):
		return LoopActivityStatusFailed, "Review timing needs attention"
	default:
		return LoopActivityStatusFailed, "Review unavailable"
	}
}

func safeActivityName(loopName string) string {
	name := strings.TrimSpace(loopName)
	if name == "" {
		return "Organization review"
	}
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "department"):
		return "Department check"
	case strings.Contains(lower, "agent type"):
		return "Specialist review"
	default:
		return name
	}
}

func isAllowedReviewLoopEventKind(kind ReviewLoopEventKind) bool {
	switch kind {
	case ReviewLoopEventOrganizationCreated,
		ReviewLoopEventTeamLeadActionCompleted,
		ReviewLoopEventOrganizationAIEngineChanged,
		ReviewLoopEventResponseContractChanged:
		return true
	default:
		return false
	}
}

func matchesReviewLoopEvent(profile LoopProfile, eventKind ReviewLoopEventKind) bool {
	for _, trigger := range profile.EventTriggers {
		if trigger == eventKind {
			return true
		}
	}
	return false
}

func loopEventTriggerLabel(eventKind ReviewLoopEventKind) string {
	return "event:" + string(eventKind)
}

func (s *AdminServer) triggerReviewLoopsForEvent(orgID string, eventKind ReviewLoopEventKind) (loopEventDispatchStats, error) {
	if !isAllowedReviewLoopEventKind(eventKind) {
		return loopEventDispatchStats{}, fmt.Errorf("review event must be one of the allowed internal review events")
	}

	home, ok := s.organizationStore().Get(orgID)
	if !ok {
		return loopEventDispatchStats{}, fmt.Errorf("organization not found")
	}

	home = normalizeOrganizationHome(home)
	s.loopProfileStore().EnsureDefaults(home)

	stats := loopEventDispatchStats{}
	triggerLabel := loopEventTriggerLabel(eventKind)
	for _, profile := range s.loopProfileStore().ListByOrganization(home.ID) {
		if !matchesReviewLoopEvent(profile, eventKind) {
			continue
		}

		key := scheduledLoopExecutionKey(home.ID, profile.ID)
		if !s.loopExecutionTracker().TryStart(key) {
			stats.skippedBusy++
			log.Printf("[review-loop-event] organization=%s loop=%s event=%s skipped_previous_execution_still_running=true", home.ID, profile.ID, eventKind)
			continue
		}

		func() {
			defer s.loopExecutionTracker().Finish(key)

			result, err := s.executeReviewLoop(home, profile, triggerLabel)
			if err != nil {
				s.recordFailedLoopResult(home, profile, triggerLabel, err)
				stats.failed++
				log.Printf("[review-loop-event] organization=%s loop=%s event=%s failed error=%v", home.ID, profile.ID, eventKind, err)
				return
			}

			stats.executed++
			log.Printf("[review-loop-event] organization=%s loop=%s event=%s completed result_id=%s", home.ID, profile.ID, eventKind, result.ID)
		}()
	}

	return stats, nil
}

func (s *AdminServer) recordFailedLoopResult(home OrganizationHomePayload, profile LoopProfile, trigger string, err error) ReviewLoopResult {
	status, summary := summarizeActivityFromError(err)
	result := ReviewLoopResult{
		ID:               uuid.NewString(),
		LoopID:           profile.ID,
		LoopName:         profile.Name,
		LoopType:         profile.Type,
		OrganizationID:   home.ID,
		OrganizationName: safeOrganizationName(home.Name),
		Trigger:          trigger,
		ActivityStatus:   status,
		ActivitySummary:  summary,
		ReviewedAt:       time.Now().UTC().Format(time.RFC3339),
	}
	s.loopResultStore().Add(home.ID, result)
	return result
}

func recentLoopActivity(results []ReviewLoopResult, limit int) []LoopActivityItem {
	if len(results) == 0 {
		return nil
	}
	if limit <= 0 {
		limit = 5
	}
	if len(results) > limit {
		results = results[:limit]
	}

	items := make([]LoopActivityItem, 0, len(results))
	for _, result := range results {
		items = append(items, LoopActivityItem{
			ID:        result.ID,
			Name:      safeActivityName(result.LoopName),
			LastRunAt: result.ReviewedAt,
			Status:    result.ActivityStatus,
			Summary:   result.ActivitySummary,
		})
	}
	return items
}

func safeAutomationName(profile LoopProfile) string {
	name := strings.TrimSpace(profile.Name)
	if name == "" {
		return "Organization review"
	}
	return name
}

func safeAutomationPurpose(profile LoopProfile, owner LoopOwnerResolution) string {
	description := strings.TrimSpace(profile.Description)
	if description != "" {
		return description
	}
	switch owner.Type {
	case LoopOwnerTypeTeam:
		return fmt.Sprintf("Keeps %s under review so the Team Lead can see whether the current structure is still ready.", owner.Name)
	case LoopOwnerTypeAgentType:
		return fmt.Sprintf("Keeps the %s specialist role under review so inherited defaults stay safe and understandable.", owner.Name)
	default:
		return "Keeps part of the AI Organization under review without taking action."
	}
}

func safeAutomationOwnerLabel(owner LoopOwnerResolution) string {
	switch owner.Type {
	case LoopOwnerTypeTeam:
		return fmt.Sprintf("Team: %s", owner.Name)
	case LoopOwnerTypeAgentType:
		return fmt.Sprintf("Specialist role: %s", owner.Name)
	default:
		if strings.TrimSpace(owner.Name) != "" {
			return owner.Name
		}
		return "Owner unavailable"
	}
}

func safeAutomationWatches(home OrganizationHomePayload, owner LoopOwnerResolution) string {
	switch owner.Type {
	case LoopOwnerTypeTeam:
		return fmt.Sprintf("Watches %s structure, specialist coverage, and current organization defaults inside AI Organization %s.", owner.Name, safeOrganizationName(home.Name))
	case LoopOwnerTypeAgentType:
		return fmt.Sprintf("Watches the %s specialist role, its working focus, and the defaults it inherits inside AI Organization %s.", owner.Name, safeOrganizationName(home.Name))
	default:
		return fmt.Sprintf("Watches the current organization setup and guided defaults inside AI Organization %s.", safeOrganizationName(home.Name))
	}
}

func automationTriggerType(profile LoopProfile) AutomationTriggerType {
	if profile.IntervalSeconds > 0 {
		return AutomationTriggerTypeScheduled
	}
	return AutomationTriggerTypeEventDriven
}

func humanizeLoopInterval(seconds int) string {
	if seconds <= 0 {
		return ""
	}
	if seconds < 60 {
		if seconds == 1 {
			return "every second"
		}
		return fmt.Sprintf("every %d seconds", seconds)
	}
	if seconds%3600 == 0 {
		hours := seconds / 3600
		if hours == 1 {
			return "every hour"
		}
		return fmt.Sprintf("every %d hours", hours)
	}
	if seconds%60 == 0 {
		minutes := seconds / 60
		if minutes == 1 {
			return "every minute"
		}
		return fmt.Sprintf("every %d minutes", minutes)
	}
	return fmt.Sprintf("every %d seconds", seconds)
}

func safeAutomationEventLabel(eventKind ReviewLoopEventKind) string {
	switch eventKind {
	case ReviewLoopEventOrganizationCreated:
		return "organization setup"
	case ReviewLoopEventTeamLeadActionCompleted:
		return "Team Lead guidance"
	case ReviewLoopEventOrganizationAIEngineChanged:
		return "AI Engine changes"
	case ReviewLoopEventResponseContractChanged:
		return "Response Style changes"
	default:
		return ""
	}
}

func joinWithOr(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " or " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", or " + items[len(items)-1]
	}
}

func safeAutomationTriggerSummary(profile LoopProfile) string {
	eventLabels := make([]string, 0, len(profile.EventTriggers))
	for _, trigger := range profile.EventTriggers {
		if label := safeAutomationEventLabel(trigger); label != "" {
			eventLabels = append(eventLabels, label)
		}
	}

	switch {
	case profile.IntervalSeconds > 0 && len(eventLabels) > 0:
		return fmt.Sprintf("Runs %s and also after %s.", humanizeLoopInterval(profile.IntervalSeconds), joinWithOr(eventLabels))
	case profile.IntervalSeconds > 0:
		return fmt.Sprintf("Runs %s.", humanizeLoopInterval(profile.IntervalSeconds))
	case len(eventLabels) > 0:
		return fmt.Sprintf("Runs after %s.", joinWithOr(eventLabels))
	default:
		return "Runs when this review is triggered from the organization workspace."
	}
}

func safeAutomationOwner(home OrganizationHomePayload, profile LoopProfile) LoopOwnerResolution {
	owner, err := resolveLoopOwner(home, profile)
	if err == nil {
		return owner
	}

	switch profile.Owner.Type {
	case LoopOwnerTypeTeam:
		return LoopOwnerResolution{
			Type:      LoopOwnerTypeTeam,
			ID:        profile.Owner.ID,
			Name:      "Team owner unavailable",
			HelpsWith: "This Automation is waiting for its Team owner to become available again.",
		}
	case LoopOwnerTypeAgentType:
		return LoopOwnerResolution{
			Type:      LoopOwnerTypeAgentType,
			ID:        profile.Owner.ID,
			Name:      "Specialist role unavailable",
			HelpsWith: "This Automation is waiting for its Specialist role to become available again.",
		}
	default:
		return LoopOwnerResolution{
			Type:      profile.Owner.Type,
			ID:        profile.Owner.ID,
			Name:      "Owner unavailable",
			HelpsWith: "This Automation is waiting for its owner to become available again.",
		}
	}
}

func recentAutomationOutcomes(results []ReviewLoopResult, loopID string, limit int) []AutomationOutcomeItem {
	if limit <= 0 {
		limit = 3
	}

	items := make([]AutomationOutcomeItem, 0, limit)
	for _, result := range results {
		if result.LoopID != loopID {
			continue
		}
		items = append(items, AutomationOutcomeItem{
			Summary:    strings.TrimSpace(result.ActivitySummary),
			OccurredAt: result.ReviewedAt,
		})
		if len(items) == limit {
			break
		}
	}
	return items
}

func automationStatusForLoop(results []ReviewLoopResult, loopID string) LoopActivityStatus {
	for _, result := range results {
		if result.LoopID == loopID {
			if result.ActivityStatus == "" {
				return LoopActivityStatusSuccess
			}
			return result.ActivityStatus
		}
	}
	return LoopActivityStatusSuccess
}

func organizationAutomations(home OrganizationHomePayload, profiles []LoopProfile, results []ReviewLoopResult) []OrganizationAutomationItem {
	if len(profiles) == 0 {
		return nil
	}

	items := make([]OrganizationAutomationItem, 0, len(profiles))
	for _, profile := range profiles {
		owner := safeAutomationOwner(home, profile)
		items = append(items, OrganizationAutomationItem{
			ID:             profile.ID,
			Name:           safeAutomationName(profile),
			Purpose:        safeAutomationPurpose(profile, owner),
			TriggerType:    automationTriggerType(profile),
			OwnerLabel:     safeAutomationOwnerLabel(owner),
			Status:         automationStatusForLoop(results, profile.ID),
			Watches:        safeAutomationWatches(home, owner),
			TriggerSummary: safeAutomationTriggerSummary(profile),
			RecentOutcomes: recentAutomationOutcomes(results, profile.ID, 3),
		})
	}

	return items
}

func validateLoopProfile(profile LoopProfile) error {
	if profile.Type != LoopProfileTypeReview {
		return fmt.Errorf("loop profile must be a review loop")
	}
	if strings.TrimSpace(profile.ID) == "" {
		return fmt.Errorf("loop profile id is required")
	}
	if profile.Owner.Type != LoopOwnerTypeTeam && profile.Owner.Type != LoopOwnerTypeAgentType {
		return fmt.Errorf("loop owner type must be team or agent_type")
	}
	if strings.TrimSpace(profile.Owner.ID) == "" {
		return fmt.Errorf("loop owner id is required")
	}
	if profile.IntervalSeconds < 0 {
		return fmt.Errorf("interval_seconds must be greater than or equal to zero")
	}
	for _, trigger := range profile.EventTriggers {
		if !isAllowedReviewLoopEventKind(trigger) {
			return fmt.Errorf("event_triggers must use allowed internal review events only")
		}
	}
	return nil
}

func (s *AdminServer) handleTriggerLoop(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	loopID := strings.TrimSpace(r.PathValue("loopId"))
	if orgID == "" || loopID == "" {
		respondAPIError(w, "organization id and loop id are required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(orgID)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	s.loopProfileStore().EnsureDefaults(home)
	profile, ok := s.loopProfileStore().Get(orgID, loopID)
	if !ok {
		respondAPIError(w, "loop profile not found", http.StatusNotFound)
		return
	}

	result, err := s.executeReviewLoop(home, profile, "manual")
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

func (s *AdminServer) handleListLoopResults(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	if orgID == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	if _, ok := s.organizationStore().Get(orgID); !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(s.loopResultStore().List(orgID)))
}

func (s *AdminServer) handleListLoopActivity(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	if orgID == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	if _, ok := s.organizationStore().Get(orgID); !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(recentLoopActivity(s.loopResultStore().List(orgID), 8)))
}

func (s *AdminServer) handleListAutomations(w http.ResponseWriter, r *http.Request) {
	orgID := strings.TrimSpace(r.PathValue("id"))
	if orgID == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(orgID)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	home = normalizeOrganizationHome(home)
	s.loopProfileStore().EnsureDefaults(home)
	items := organizationAutomations(home, s.loopProfileStore().ListByOrganization(orgID), s.loopResultStore().List(orgID))
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(items))
}
