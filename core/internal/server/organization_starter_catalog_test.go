package server

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

func fixtureStarterPolicy(names ...string) *governance.PolicyConfig {
	cfg := &governance.PolicyConfig{Defaults: governance.DefaultConfig{DefaultAction: governance.ActionAllow}}
	for _, name := range names {
		cfg.Groups = append(cfg.Groups, governance.PolicyGroup{
			Name:    name,
			Targets: []string{"*"},
			Rules:   []governance.PolicyRule{{Intent: "*", Action: governance.ActionRequireApproval}},
		})
	}
	return cfg
}

// writeStarterCatalog writes the existing top-level test bundle plus the
// RC-004A fixture starter under starters/.
func writeStarterCatalog(t *testing.T, policyRef string) string {
	t.Helper()
	dir := writeStarterBundle(t)
	data, err := os.ReadFile(filepath.Join("..", "..", "config", "templates", "starters", "testdata", "minimal-starter.yaml"))
	if err != nil {
		t.Fatalf("read fixture starter: %v", err)
	}
	body := strings.Replace(string(data), "fixture-approve-external", policyRef, 1)
	startersDir := bootstrap.StartersPath(dir)
	if err := os.MkdirAll(startersDir, 0o755); err != nil {
		t.Fatalf("mkdir starters: %v", err)
	}
	if err := os.WriteFile(filepath.Join(startersDir, "fixture.yaml"), []byte(body), 0o644); err != nil {
		t.Fatalf("write fixture starter: %v", err)
	}
	return dir
}

func listStarterTemplates(t *testing.T, s *AdminServer) []map[string]any {
	t.Helper()
	rr := doRequest(t, http.HandlerFunc(s.handleListTemplatesAPI), "GET", "/api/v1/templates?view=organization-starters", "")
	assertStatus(t, rr, http.StatusOK)
	var resp protocol.APIResponse
	assertJSON(t, rr, &resp)
	items, ok := resp.Data.([]any)
	if !resp.OK || !ok {
		t.Fatalf("unexpected response: ok=%v data=%T", resp.OK, resp.Data)
	}
	out := make([]map[string]any, 0, len(items))
	for _, item := range items {
		out = append(out, item.(map[string]any))
	}
	return out
}

func TestOrganizationStartersProjectStarterAndDeliveryDefaults(t *testing.T) {
	s := newTestServer(
		withTemplateBundlesPath(writeStarterCatalog(t, "fixture-approve-external")),
		withGuard(fixtureStarterPolicy("fixture-approve-external")),
	)
	items := listStarterTemplates(t, s)
	if len(items) != 2 {
		t.Fatalf("expected bundle plus starter, got %d", len(items))
	}

	starter, bundle := items[0], items[1]
	if starter["id"] != "fixture-minimal-starter" || starter["source"] != "starter" || starter["available"] != true {
		t.Fatalf("labelled starter should sort first and be available: %+v", starter)
	}
	// Existing summary fields keep their names.
	for _, key := range []string{"name", "organization_type", "department_count", "specialist_count", "departments", "ai_engine_settings_summary"} {
		if _, ok := starter[key]; !ok {
			t.Errorf("starter missing existing field %q", key)
		}
	}
	labels := starter["starter"].(map[string]any)
	if labels["label"] != "Fixture studio" || labels["persona"] != "studio" || labels["sample_name"] != "Fixture Studio" || labels["order"] != float64(10) {
		t.Fatalf("unexpected starter labels: %+v", labels)
	}
	defaults := starter["delivery_defaults"].(map[string]any)
	if defaults["work_mode"] != "propose_only" || defaults["approval_policy_ref"] != "fixture-approve-external" ||
		defaults["approval_posture"] != "approve_external_sends" || defaults["audit"] != "standard" || defaults["max_team_members"] != float64(3) {
		t.Fatalf("unexpected delivery_defaults: %+v", defaults)
	}
	outputs := defaults["outputs"].(map[string]any)
	if outputs["retain"] != true || outputs["proof_required"] != true {
		t.Fatalf("unexpected outputs: %+v", outputs)
	}
	summary := starter["delivery_summary"].(map[string]any)
	if summary["work_mode_label"] != "Propose only" || !strings.Contains(summary["approval_posture"].(string), "until you approve") ||
		summary["output_proof"] != "Outputs are retained with proof." || summary["team_shape"] != "Up to 3 members per team, plus Soma." {
		t.Fatalf("unexpected delivery_summary: %+v", summary)
	}

	if bundle["id"] != "engineering-starter" || bundle["source"] != "bundle" || bundle["available"] != true {
		t.Fatalf("unexpected top-level bundle item: %+v", bundle)
	}
	if _, ok := bundle["starter"]; ok {
		t.Fatalf("bundle without starter block must omit starter: %+v", bundle)
	}
	if _, ok := bundle["delivery_defaults"]; ok {
		t.Fatalf("bundle without delivery_defaults must omit it: %+v", bundle)
	}

	resolved := mustResolveStarterTemplate(t, s, "fixture-minimal-starter")
	if resolved.DepartmentCount != 1 || resolved.SpecialistCount != 3 {
		t.Fatalf("unexpected resolved starter summary: %+v", resolved)
	}
	entry, err := s.resolveStarterCatalogEntry("fixture-minimal-starter")
	if err != nil || entry.DeliveryDefaults == nil || entry.Starter.SomaIntro == "" {
		t.Fatalf("resolveStarterCatalogEntry() = %+v, %v", entry, err)
	}
}

func TestOrganizationStartersUnknownPolicyMarksStarterUnavailable(t *testing.T) {
	s := newTestServer(
		withTemplateBundlesPath(writeStarterCatalog(t, "missing-policy")),
		withGuard(fixtureStarterPolicy("fixture-approve-external")),
	)
	items := listStarterTemplates(t, s)
	if len(items) != 2 {
		t.Fatalf("unavailable starter should still be listed, got %d items", len(items))
	}
	starter := items[0]
	if starter["available"] != false || !strings.Contains(starter["unavailable_reason"].(string), `approval policy "missing-policy" is not a group in policy.yaml`) {
		t.Fatalf("expected unavailable starter with reason, got %+v", starter)
	}
	if items[1]["available"] != true {
		t.Fatalf("top-level bundle must stay available: %+v", items[1])
	}

	if _, err := s.resolveStarterTemplate("fixture-minimal-starter"); err == nil {
		t.Fatal("an unavailable starter must not resolve for organization creation")
	}
	templates, err := s.loadOrganizationStarterTemplates()
	if err != nil || len(templates) != 1 || templates[0].ID != "engineering-starter" {
		t.Fatalf("loadOrganizationStarterTemplates() = %+v, %v", templates, err)
	}
}

func TestOrganizationStartersWithoutPolicyMarkStarterUnavailable(t *testing.T) {
	s := newTestServer(withTemplateBundlesPath(writeStarterCatalog(t, "fixture-approve-external")))
	items := listStarterTemplates(t, s)
	if items[0]["available"] != false || !strings.Contains(items[0]["unavailable_reason"].(string), "governance policy is not loaded") {
		t.Fatalf("expected starter unavailable without policy, got %+v", items[0])
	}
}
