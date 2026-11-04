package bootstrap

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestTemplateBundle_ValidateLoops_AcceptsReviewAndLearningLoops(t *testing.T) {
	raw := `
id: loop-starter
name: Loop Starter
teams:
  - id: product-delivery
    name: Product Delivery
    inputs:
      - swarm.team.product-delivery.internal.command
    deliveries:
      - swarm.team.product-delivery.signal.status
    members:
      - id: product-lead
        role: product_lead
      - id: verifier
        role: verifier
loops:
  - id: delivery-review
    name: Product delivery readiness review
    type: review
    description: Reviews team readiness without taking action.
    owner_type: team
    owner_id: product-delivery
    interval_seconds: 300
    event_triggers:
      - organization_created
      - team_lead_action_completed
  - id: verifier-learning
    name: Verification proof learning loop
    type: learning
    description: Captures recurring verification patterns.
    owner_type: agent_type
    owner_id: verifier
    interval_seconds: 900
    event_triggers:
      - organization_created
`
	bundle, err := ParseTemplateBundle([]byte(raw))
	if err != nil {
		t.Fatalf("ParseTemplateBundle() error = %v", err)
	}
	if err := bundle.Validate(); err != nil {
		t.Fatalf("Validate() error = %v", err)
	}
	if len(bundle.Loops) != 2 {
		t.Fatalf("expected 2 loops, got %d", len(bundle.Loops))
	}
	if bundle.Loops[0].Type != TemplateLoopTypeReview || bundle.Loops[1].Type != TemplateLoopTypeLearning {
		t.Fatalf("unexpected loop types: %+v", bundle.Loops)
	}
}

func TestTemplateBundle_ValidateLoops_RejectsInvalidDefinitions(t *testing.T) {
	base := `
id: loop-starter
name: Loop Starter
teams:
  - id: product-delivery
    name: Product Delivery
    inputs:
      - swarm.team.product-delivery.internal.command
    deliveries:
      - swarm.team.product-delivery.signal.status
    members:
      - id: product-lead
        role: product_lead
`
	cases := []struct {
		name    string
		loops   string
		wantSub string
	}{
		{
			name: "invalid loop type",
			loops: `
loops:
  - id: bad-type
    name: Bad type
    type: custom
    owner_type: team
    owner_id: product-delivery
`,
			wantSub: `type must be "review" or "learning"`,
		},
		{
			name: "unknown team owner_id",
			loops: `
loops:
  - id: bad-team
    name: Bad team owner
    type: review
    owner_type: team
    owner_id: unknown-team
`,
			wantSub: `owner_id "unknown-team" not found in bundle teams`,
		},
		{
			name: "unknown member owner_id",
			loops: `
loops:
  - id: bad-member
    name: Bad member owner
    type: learning
    owner_type: agent_type
    owner_id: unknown-member
`,
			wantSub: `owner_id "unknown-member" not found in bundle team members`,
		},
		{
			name: "negative interval_seconds",
			loops: `
loops:
  - id: bad-interval
    name: Bad interval
    type: review
    owner_type: team
    owner_id: product-delivery
    interval_seconds: -5
`,
			wantSub: "interval_seconds must be >= 0",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			bundle, err := ParseTemplateBundle([]byte(base + tc.loops))
			if err != nil {
				t.Fatalf("ParseTemplateBundle() error = %v", err)
			}
			err = bundle.Validate()
			if err == nil || !strings.Contains(err.Error(), tc.wantSub) {
				t.Fatalf("Validate() error = %v, want substring %q", err, tc.wantSub)
			}
		})
	}
}

func TestRealStarterBundles_ValidateLoops(t *testing.T) {
	startersDir := filepath.Join("..", "..", "config", "templates", "starters")
	loader := NewTemplateLoader(startersDir)
	bundles, err := loader.LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles(%s) error = %v", startersDir, err)
	}
	if len(bundles) != 4 {
		t.Fatalf("expected 4 starter bundles, got %d", len(bundles))
	}
	for _, bundle := range bundles {
		if len(bundle.Loops) != 2 {
			t.Fatalf("starter %s: expected 2 bundle loops, got %d", bundle.ID, len(bundle.Loops))
		}
		hasReview := false
		hasLearning := false
		for _, loop := range bundle.Loops {
			if loop.Type == TemplateLoopTypeReview {
				hasReview = true
			}
			if loop.Type == TemplateLoopTypeLearning {
				hasLearning = true
			}
		}
		if !hasReview || !hasLearning {
			t.Fatalf("starter %s: expected both review and learning loops, got %+v", bundle.ID, bundle.Loops)
		}
	}
}
