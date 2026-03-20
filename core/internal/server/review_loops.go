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

const (
	LoopProfileTypeReview LoopProfileType = "review"

	LoopOwnerTypeTeam      LoopOwnerType = "team"
	LoopOwnerTypeAgentType LoopOwnerType = "agent_type"

	DefaultDepartmentReviewLoopID = "department-readiness-review"
	DefaultAgentTypeReviewLoopID  = "agent-type-readiness-review"
)

type LoopOwnerRef struct {
	Type LoopOwnerType `json:"type"`
	ID   string        `json:"id"`
}

type LoopProfile struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Type            LoopProfileType `json:"type"`
	Description     string          `json:"description,omitempty"`
	Owner           LoopOwnerRef    `json:"owner"`
	IntervalSeconds int             `json:"interval_seconds,omitempty"`
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
