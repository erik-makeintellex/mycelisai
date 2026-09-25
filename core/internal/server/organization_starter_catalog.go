package server

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/governance"
	"github.com/mycelis/core/pkg/protocol"
)

const operatorCatalogTimeout = 3 * time.Second

const (
	starterSourceBundle  = "bundle"
	starterSourceStarter = "starter"
	// starterSourceOperator marks an operator template (RC-004G clone),
	// labelled "Custom" in the UI.
	starterSourceOperator = "operator"
)

// OrganizationStarterTemplate is one item of
// GET /api/v1/templates?view=organization-starters. It embeds the existing
// OrganizationTemplateSummary so every existing JSON field keeps its name,
// and adds the RC-004A starter projection.
type OrganizationStarterTemplate struct {
	OrganizationTemplateSummary
	Source            string                      `json:"source"`
	SourceTemplateID  string                      `json:"source_template_id,omitempty"`
	Available         bool                        `json:"available"`
	UnavailableReason string                      `json:"unavailable_reason,omitempty"`
	Starter           *bootstrap.StarterLabels    `json:"starter,omitempty"`
	DeliveryDefaults  *bootstrap.DeliveryDefaults `json:"delivery_defaults,omitempty"`
	DeliverySummary   *StarterDeliverySummary     `json:"delivery_summary,omitempty"`
}

// StarterDeliverySummary is the plain-language delivery posture for chips.
type StarterDeliverySummary struct {
	WorkModeLabel   string `json:"work_mode_label"`
	ApprovalPosture string `json:"approval_posture"`
	OutputProof     string `json:"output_proof"`
	TeamShape       string `json:"team_shape"`
}

// loadOrganizationStarterCatalog lists top-level bundles plus starters from
// <templates>/starters/. Boot selection is unaffected: boot only reads the
// top level. Invalid starters stay in the list as unavailable.
func (s *AdminServer) loadOrganizationStarterCatalog() ([]OrganizationStarterTemplate, error) {
	templatesPath := s.templateBundlesPath()
	bundles, err := bootstrap.NewTemplateLoader(templatesPath).LoadBundles()
	if err != nil {
		return nil, err
	}
	policyGroups := s.starterPolicyGroups()

	reserved := make(map[string]struct{}, len(bundles))
	catalog := make([]OrganizationStarterTemplate, 0, len(bundles))
	for _, bundle := range bundles {
		reserved[bundle.ID] = struct{}{}
		item := projectStarterBundle(bundle, starterSourceBundle)
		if bundle.DeliveryDefaults != nil {
			if err := bundle.DeliveryDefaults.ResolveApprovalPolicy(policyGroups); err != nil {
				item.Available = false
				item.UnavailableReason = "delivery_defaults: " + err.Error()
			}
		}
		catalog = append(catalog, item)
	}

	entries, err := bootstrap.LoadStarterEntries(bootstrap.StartersPath(templatesPath), reserved, policyGroups)
	if err != nil {
		return nil, err
	}
	for _, entry := range entries {
		reserved[entry.ID] = struct{}{}
		catalog = append(catalog, projectStarterEntry(entry))
	}
	catalog = append(catalog, s.operatorStarterCatalog(reserved, policyGroups)...)

	sortStarterCatalog(catalog)
	return catalog, nil
}

// operatorStarterCatalog projects operator templates (RC-004G clones) with
// the starter availability rules. Storage problems degrade to no operator
// entries rather than failing the file catalog.
func (s *AdminServer) operatorStarterCatalog(reserved, policyGroups map[string]struct{}) []OrganizationStarterTemplate {
	if s.getDB() == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), operatorCatalogTimeout)
	defer cancel()
	rows, err := s.listOperatorTemplates(ctx)
	if err != nil {
		log.Printf("[starter-catalog] operator templates unavailable: %v", err)
		return nil
	}
	docs := make([]bootstrap.OperatorTemplateDocument, 0, len(rows))
	sources := make(map[string]string, len(rows))
	for _, row := range rows {
		docs = append(docs, bootstrap.OperatorTemplateDocument{ID: row.ID, Document: row.Document})
		sources[row.ID] = row.SourceTemplateID
	}
	entries := bootstrap.LoadOperatorStarterEntries(docs, reserved, policyGroups)
	out := make([]OrganizationStarterTemplate, 0, len(entries))
	for _, entry := range entries {
		item := projectStarterEntry(entry)
		item.Source = starterSourceOperator
		item.SourceTemplateID = sources[entry.ID]
		out = append(out, item)
	}
	return out
}

// resolveStarterCatalogEntry returns the available catalog item for id, with
// its starter labels and delivery defaults. Organization create/update code
// can use it to copy delivery_defaults onto the organization home.
func (s *AdminServer) resolveStarterCatalogEntry(id string) (*OrganizationStarterTemplate, error) {
	catalog, err := s.loadOrganizationStarterCatalog()
	if err != nil {
		return nil, err
	}
	for i := range catalog {
		if catalog[i].ID == id && catalog[i].Available {
			item := catalog[i]
			return &item, nil
		}
	}
	return nil, fmt.Errorf("starter template %q not found or unavailable", id)
}

// starterPolicyGroups returns the policy.yaml group names. It prefers the
// live governance guard and falls back to reading the policy file. nil means
// no policy could be loaded.
func (s *AdminServer) starterPolicyGroups() map[string]struct{} {
	var cfg *governance.PolicyConfig
	if s.Guard != nil && s.Guard.Engine != nil {
		cfg = s.Guard.GetPolicyConfig()
	} else if engine, err := governance.NewEngine(defaultPolicyPath); err == nil {
		cfg = engine.Config
	}
	if cfg == nil {
		return nil
	}
	groups := make(map[string]struct{}, len(cfg.Groups))
	for _, group := range cfg.Groups {
		if name := strings.TrimSpace(group.Name); name != "" {
			groups[name] = struct{}{}
		}
	}
	return groups
}

func projectStarterBundle(bundle *bootstrap.TemplateBundle, source string) OrganizationStarterTemplate {
	return OrganizationStarterTemplate{
		OrganizationTemplateSummary: summarizeStarterBundle(bundle),
		Source:                      source,
		Available:                   true,
		Starter:                     bundle.Starter,
		DeliveryDefaults:            bundle.DeliveryDefaults,
		DeliverySummary:             summarizeDelivery(bundle.DeliveryDefaults),
	}
}

func projectStarterEntry(entry *bootstrap.StarterEntry) OrganizationStarterTemplate {
	var item OrganizationStarterTemplate
	if entry.Bundle != nil {
		item = projectStarterBundle(entry.Bundle, starterSourceStarter)
	} else {
		item = OrganizationStarterTemplate{
			OrganizationTemplateSummary: OrganizationTemplateSummary{
				ID:               entry.ID,
				Name:             entry.ID,
				OrganizationType: "AI Organization starter",
				TeamLeadLabel:    "Team Lead",
			},
			Source: starterSourceStarter,
		}
	}
	item.Available = entry.Available
	item.UnavailableReason = entry.UnavailableReason
	return item
}

// sortStarterCatalog puts labelled starters first by order then name, and
// unlabelled bundles after them by name.
func sortStarterCatalog(catalog []OrganizationStarterTemplate) {
	sort.SliceStable(catalog, func(i, j int) bool {
		a, b := catalog[i], catalog[j]
		if (a.Starter != nil) != (b.Starter != nil) {
			return a.Starter != nil
		}
		if a.Starter != nil && a.Starter.Order != b.Starter.Order {
			return a.Starter.Order < b.Starter.Order
		}
		return a.Name < b.Name
	})
}

var workModeLabels = map[string]string{
	protocol.WorkModeReadOnly:            "Read only",
	protocol.WorkModeProposeOnly:         "Propose only",
	protocol.WorkModeExecuteWithApproval: "Execute with approval",
	protocol.WorkModeExecuteBounded:      "Bounded execution",
	protocol.WorkModeResumeContinuity:    "Resume continuity",
}

var approvalPostureSentences = map[string]string{
	bootstrap.PostureApproveExternalSends: "Nothing leaves the organization until you approve it.",
	bootstrap.PostureApproveChanges:       "Deploy and write actions wait for your approval.",
	bootstrap.PostureApproveHighRisk:      "Routine checks run within bounds; high-risk actions wait for your approval.",
	bootstrap.PostureApproveAllMutations:  "Every change waits for your approval and is fully audited.",
}

func summarizeDelivery(defaults *bootstrap.DeliveryDefaults) *StarterDeliverySummary {
	if defaults == nil {
		return nil
	}
	output := "Outputs are not retained by default."
	switch {
	case defaults.Outputs.Retain && defaults.Outputs.ProofRequired:
		output = "Outputs are retained with proof."
	case defaults.Outputs.Retain:
		output = "Outputs are retained without a proof requirement."
	}
	return &StarterDeliverySummary{
		WorkModeLabel:   workModeLabels[defaults.WorkMode],
		ApprovalPosture: approvalPostureSentences[defaults.ApprovalPosture],
		OutputProof:     output,
		TeamShape:       fmt.Sprintf("Up to %d members per team, plus Soma.", defaults.MaxTeamMembers),
	}
}
