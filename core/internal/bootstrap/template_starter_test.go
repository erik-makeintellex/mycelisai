package bootstrap

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

var fixtureStarterPath = filepath.Join("..", "..", "config", "templates", "starters", "testdata", "minimal-starter.yaml")

func validDefaults() DeliveryDefaults {
	return DeliveryDefaults{
		WorkMode:          protocol.WorkModeProposeOnly,
		ApprovalPolicyRef: "delivery-approve-external",
		ApprovalPosture:   PostureApproveExternalSends,
		Outputs:           DeliveryOutputs{Retain: true, ProofRequired: true},
		MaxTeamMembers:    3,
	}
}

func TestDeliveryDefaultsExpressGraduatedPostures(t *testing.T) {
	cases := map[string]DeliveryDefaults{
		"agency":     {WorkMode: protocol.WorkModeProposeOnly, ApprovalPosture: PostureApproveExternalSends},
		"saas":       {WorkMode: protocol.WorkModeExecuteWithApproval, ApprovalPosture: PostureApproveChanges},
		"operations": {WorkMode: protocol.WorkModeExecuteBounded, ApprovalPosture: PostureApproveHighRisk},
		"regulated":  {WorkMode: protocol.WorkModeProposeOnly, ApprovalPosture: PostureApproveAllMutations, Audit: AuditFull},
	}
	for name, d := range cases {
		d.ApprovalPolicyRef = "delivery-" + name
		d.Outputs = DeliveryOutputs{Retain: true, ProofRequired: true}
		d.MaxTeamMembers = 3
		if err := d.Validate(); err != nil {
			t.Errorf("%s posture should validate: %v", name, err)
		}
	}
}

func TestDeliveryDefaultsValidationErrors(t *testing.T) {
	cases := []struct {
		name   string
		mutate func(*DeliveryDefaults)
		want   string
	}{
		{"unknown work mode", func(d *DeliveryDefaults) { d.WorkMode = "yolo" }, `unknown work_mode "yolo"`},
		{"missing policy", func(d *DeliveryDefaults) { d.ApprovalPolicyRef = "  " }, "approval_policy_ref is required"},
		{"unknown posture", func(d *DeliveryDefaults) { d.ApprovalPosture = "trust_me" }, `unknown approval_posture "trust_me"`},
		{"unknown audit", func(d *DeliveryDefaults) { d.Audit = "partial" }, `unknown audit "partial"`},
		{"team size zero", func(d *DeliveryDefaults) { d.MaxTeamMembers = 0 }, "max_team_members must be between 1 and 5, got 0"},
		{"team size six", func(d *DeliveryDefaults) { d.MaxTeamMembers = 6 }, "max_team_members must be between 1 and 5, got 6"},
		{"proof without retain", func(d *DeliveryDefaults) { d.Outputs.Retain = false }, "outputs.proof_required needs outputs.retain: true"},
		{"regulated bounded", func(d *DeliveryDefaults) {
			d.ApprovalPosture, d.Audit, d.WorkMode = PostureApproveAllMutations, AuditFull, protocol.WorkModeExecuteBounded
		}, "cannot use work_mode execute_bounded"},
		{"regulated without full audit", func(d *DeliveryDefaults) { d.ApprovalPosture = PostureApproveAllMutations }, "requires audit: full"},
		{"external sends bounded", func(d *DeliveryDefaults) { d.WorkMode = protocol.WorkModeExecuteBounded }, "use approve_high_risk"},
	}
	for _, tc := range cases {
		d := validDefaults()
		tc.mutate(&d)
		err := d.Validate()
		if err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("%s: want error containing %q, got %v", tc.name, tc.want, err)
		}
	}
}

func TestDeliveryDefaultsDefaultAuditIsStandard(t *testing.T) {
	d := validDefaults()
	if err := d.Validate(); err != nil {
		t.Fatalf("Validate() failed: %v", err)
	}
	if d.Audit != AuditStandard {
		t.Fatalf("audit = %q, want standard", d.Audit)
	}
}

func TestStarterLabelsValidationErrors(t *testing.T) {
	cases := []struct {
		labels StarterLabels
		want   string
	}{
		{StarterLabels{Label: " "}, "label is required"},
		{StarterLabels{Label: strings.Repeat("x", 81)}, "at most 80 characters"},
		{StarterLabels{Label: "Studio", Persona: "pirates"}, `unknown persona "pirates"`},
		{StarterLabels{Label: "Studio", Order: -1}, "order must not be negative"},
	}
	for _, tc := range cases {
		if err := tc.labels.Validate(); err == nil || !strings.Contains(err.Error(), tc.want) {
			t.Errorf("want error containing %q, got %v", tc.want, err)
		}
	}
}

func TestResolveApprovalPolicy(t *testing.T) {
	d := validDefaults()
	if err := d.ResolveApprovalPolicy(map[string]struct{}{"delivery-approve-external": {}}); err != nil {
		t.Fatalf("known policy should resolve: %v", err)
	}
	if err := d.ResolveApprovalPolicy(map[string]struct{}{"other": {}}); err == nil || !strings.Contains(err.Error(), "is not a group in policy.yaml") {
		t.Fatalf("unknown policy error = %v", err)
	}
	if err := d.ResolveApprovalPolicy(nil); err == nil || !strings.Contains(err.Error(), "governance policy is not loaded") {
		t.Fatalf("nil catalog error = %v", err)
	}
}

func TestFixtureStarterParsesBothBlocks(t *testing.T) {
	bundle, err := loadTemplateBundle(fixtureStarterPath)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	if err := bundle.Validate(); err != nil {
		t.Fatalf("fixture Validate() failed: %v", err)
	}
	if bundle.Starter == nil || bundle.Starter.Label != "Fixture studio" || bundle.Starter.Order != 10 {
		t.Fatalf("unexpected starter block: %+v", bundle.Starter)
	}
	d := bundle.DeliveryDefaults
	if d == nil || d.WorkMode != protocol.WorkModeProposeOnly || d.ApprovalPolicyRef != "fixture-approve-external" ||
		!d.Outputs.ProofRequired || d.MaxTeamMembers != 3 || d.Audit != AuditStandard {
		t.Fatalf("unexpected delivery_defaults: %+v", d)
	}
}

func TestTeamAboveMaxTeamMembersFailsValidation(t *testing.T) {
	bundle, err := loadTemplateBundle(fixtureStarterPath)
	if err != nil {
		t.Fatalf("load fixture: %v", err)
	}
	bundle.DeliveryDefaults.MaxTeamMembers = 2
	if err := bundle.Validate(); err == nil || !strings.Contains(err.Error(), `team "fixture-delivery" has 3 members, above max_team_members 2`) {
		t.Fatalf("expected team size error, got %v", err)
	}
}

func TestBridgeBundleStaysBackwardCompatible(t *testing.T) {
	bundles, err := NewTemplateLoader(filepath.Join("..", "..", "config", "templates")).LoadBundles()
	if err != nil {
		t.Fatalf("LoadBundles() failed: %v", err)
	}
	if len(bundles) != 1 || bundles[0].ID != "v8-migration-standing-team-bridge" {
		t.Fatalf("expected only the bridge bundle at the top level, got %d", len(bundles))
	}
	if bundles[0].Starter != nil || bundles[0].DeliveryDefaults != nil {
		t.Fatalf("bridge bundle should carry no starter blocks: %+v %+v", bundles[0].Starter, bundles[0].DeliveryDefaults)
	}
}

func TestTopLevelBundleWithInvalidDeliveryDefaultsFailsLoad(t *testing.T) {
	dir := t.TempDir()
	data, err := os.ReadFile(fixtureStarterPath)
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	broken := strings.Replace(string(data), "work_mode: propose_only", "work_mode: yolo", 1)
	if err := os.WriteFile(filepath.Join(dir, "broken.yaml"), []byte(broken), 0o644); err != nil {
		t.Fatalf("write bundle: %v", err)
	}
	if _, err := NewTemplateLoader(dir).LoadBundles(); err == nil || !strings.Contains(err.Error(), `delivery_defaults: unknown work_mode "yolo"`) {
		t.Fatalf("expected delivery_defaults error, got %v", err)
	}
}
