package server

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

func TestEnsureDefaults_UsesBundleLoopProfilesAndSupportsLearningLoops(t *testing.T) {
	s := newTestServer()
	home := testReviewLoopHome()
	home.LoopProfiles = []bootstrap.TemplateLoopProfile{
		{
			ID:              "bundle-team-review",
			Name:            "Bundle team review",
			Type:            string(LoopProfileTypeReview),
			Description:     "Reviews the platform department readiness from bundle truth.",
			OwnerType:       string(LoopOwnerTypeTeam),
			OwnerID:         "platform",
			IntervalSeconds: 300,
			EventTriggers:   []string{string(ReviewLoopEventOrganizationCreated)},
		},
		{
			ID:              "bundle-planner-learning",
			Name:            "Bundle planner learning loop",
			Type:            string(LoopProfileTypeLearning),
			Description:     "Captures recurring planning insights from bundle truth.",
			OwnerType:       string(LoopOwnerTypeAgentType),
			OwnerID:         "planner",
			IntervalSeconds: 900,
			EventTriggers:   []string{string(ReviewLoopEventOrganizationCreated), string(ReviewLoopEventTeamLeadActionCompleted)},
		},
	}

	home = s.organizationStore().Save(home)
	s.loopProfileStore().EnsureDefaults(home)

	profiles := s.loopProfileStore().ListByOrganization(home.ID)
	if len(profiles) != 2 {
		t.Fatalf("expected 2 bundle-seeded loop profiles, got %+v", profiles)
	}

	learningProfile, ok := s.loopProfileStore().Get(home.ID, "bundle-planner-learning")
	if !ok {
		t.Fatalf("expected bundle-planner-learning profile to be registered, got %+v", profiles)
	}
	if learningProfile.Type != LoopProfileTypeLearning {
		t.Fatalf("expected learning loop type, got %+v", learningProfile)
	}

	result, err := s.executeReviewLoop(home, learningProfile, "manual")
	if err != nil {
		t.Fatalf("executeReviewLoop(learningProfile) error = %v", err)
	}
	if result.LoopType != LoopProfileTypeLearning {
		t.Fatalf("expected result LoopType=%q, got %+v", LoopProfileTypeLearning, result)
	}
	if result.Owner.Type != LoopOwnerTypeAgentType || result.Owner.ID != "planner" {
		t.Fatalf("expected resolved planner owner on learning loop result, got %+v", result.Owner)
	}
}

func TestHandleCreateOrganization_SeedsBundleLoopsFromRealStarters(t *testing.T) {
	engine, err := governance.NewEngine("../../config/policy.yaml")
	if err != nil {
		t.Fatalf("load real policy.yaml: %v", err)
	}

	s := newTestServer(
		withTemplateBundlesPath("../../config/templates"),
		withGuard(engine.Config),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/organizations", s.handleCreateOrganization)
	mux.HandleFunc("GET /api/v1/organizations/{id}/automations", s.handleListAutomations)

	starterExpectations := map[string]struct {
		reviewLoopID   string
		learningLoopID string
	}{
		"starter-product-delivery-team": {
			reviewLoopID:   "product-delivery-review",
			learningLoopID: "verifier-proof-learning",
		},
		"starter-client-delivery-studio": {
			reviewLoopID:   "studio-delivery-review",
			learningLoopID: "client-review-learning",
		},
		"starter-operations-desk": {
			reviewLoopID:   "ops-desk-readiness-review",
			learningLoopID: "monitor-signal-learning",
		},
		"starter-governed-enterprise-program": {
			reviewLoopID:   "governed-program-review",
			learningLoopID: "compliance-evidence-learning",
		},
	}

	for starterID, want := range starterExpectations {
		t.Run(starterID, func(t *testing.T) {
			body := fmt.Sprintf(`{"name":"Org %s","purpose":"Verify bundle loop seeding","start_mode":"template","template_id":%q}`, starterID, starterID)
			rr := doRequest(t, mux, "POST", "/api/v1/organizations", body)
			assertStatus(t, rr, http.StatusCreated)

			var createResp protocol.APIResponse
			assertJSON(t, rr, &createResp)
			data, ok := createResp.Data.(map[string]any)
			if !ok {
				t.Fatalf("expected object response data, got %T", createResp.Data)
			}
			orgID, _ := data["id"].(string)
			if orgID == "" {
				t.Fatalf("expected organization id, got %+v", data)
			}

			reviewLoop, ok := s.loopProfileStore().Get(orgID, want.reviewLoopID)
			if !ok || reviewLoop.Type != LoopProfileTypeReview {
				t.Fatalf("starter %s: expected review loop %q, got %+v (ok=%v)", starterID, want.reviewLoopID, reviewLoop, ok)
			}
			learningLoop, ok := s.loopProfileStore().Get(orgID, want.learningLoopID)
			if !ok || learningLoop.Type != LoopProfileTypeLearning {
				t.Fatalf("starter %s: expected learning loop %q, got %+v (ok=%v)", starterID, want.learningLoopID, learningLoop, ok)
			}

			results := s.loopResultStore().List(orgID)
			if len(results) != 2 {
				t.Fatalf("starter %s: expected 2 initial organization_created loop results, got %+v", starterID, results)
			}
		})
	}
}
