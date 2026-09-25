package bootstrap

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

// StarterLabels are beginner-facing labels for an organization starter.
// They never shape runtime behavior.
type StarterLabels struct {
	Label         string `yaml:"label" json:"label"`
	Persona       string `yaml:"persona,omitempty" json:"persona,omitempty"`
	Order         int    `yaml:"order,omitempty" json:"order"`
	SampleName    string `yaml:"sample_name,omitempty" json:"sample_name,omitempty"`
	SamplePurpose string `yaml:"sample_purpose,omitempty" json:"sample_purpose,omitempty"`
	SomaIntro     string `yaml:"soma_intro,omitempty" json:"soma_intro,omitempty"`
}

// DeliveryDefaults shape how an organization created from the bundle
// delivers work: work mode, approval policy, approval posture, audit depth,
// output retention and proof, and the maximum team size.
type DeliveryDefaults struct {
	WorkMode          string          `yaml:"work_mode" json:"work_mode"`
	ApprovalPolicyRef string          `yaml:"approval_policy_ref" json:"approval_policy_ref"`
	ApprovalPosture   string          `yaml:"approval_posture" json:"approval_posture"`
	Audit             string          `yaml:"audit,omitempty" json:"audit"`
	Outputs           DeliveryOutputs `yaml:"outputs" json:"outputs"`
	MaxTeamMembers    int             `yaml:"max_team_members" json:"max_team_members"`
}

// DeliveryOutputs controls retention and proof for delivered outputs.
type DeliveryOutputs struct {
	Retain        bool `yaml:"retain" json:"retain"`
	ProofRequired bool `yaml:"proof_required" json:"proof_required"`
}

// Starter personas (company types).
const (
	PersonaStudio              = "studio"
	PersonaSaaSProductTeam     = "saas_product_team"
	PersonaOperations          = "operations"
	PersonaRegulatedEnterprise = "regulated_enterprise"
)

// Graduated approval postures (RC-004 owner decision 4).
const (
	// PostureApproveExternalSends: agency. Nothing leaves the org unapproved.
	PostureApproveExternalSends = "approve_external_sends"
	// PostureApproveChanges: SaaS. Deploy and write actions need approval.
	PostureApproveChanges = "approve_changes"
	// PostureApproveHighRisk: operations. Bounded execution; high-risk
	// actions need approval.
	PostureApproveHighRisk = "approve_high_risk"
	// PostureApproveAllMutations: regulated. Every mutation needs approval
	// and audit must be full.
	PostureApproveAllMutations = "approve_all_mutations"
)

// Audit depths.
const (
	AuditStandard = "standard"
	AuditFull     = "full"
)

const (
	MinTeamMembers     = 1
	MaxTeamMembersCap  = 5
	maxStarterLabelLen = 80
)

var starterPersonas = []string{PersonaStudio, PersonaSaaSProductTeam, PersonaOperations, PersonaRegulatedEnterprise}

// ApprovalPostures lists the known approval postures.
var ApprovalPostures = []string{
	PostureApproveExternalSends,
	PostureApproveChanges,
	PostureApproveHighRisk,
	PostureApproveAllMutations,
}

func (b *TemplateBundle) validateStarterBlocks() error {
	if b.Starter != nil {
		if err := b.Starter.Validate(); err != nil {
			return fmt.Errorf("starter: %w", err)
		}
	}
	if b.DeliveryDefaults == nil {
		return nil
	}
	if err := b.DeliveryDefaults.Validate(); err != nil {
		return fmt.Errorf("delivery_defaults: %w", err)
	}
	limit := b.DeliveryDefaults.MaxTeamMembers
	for _, team := range b.Teams {
		if team != nil && len(team.Members) > limit {
			return fmt.Errorf("delivery_defaults: team %q has %d members, above max_team_members %d", team.ID, len(team.Members), limit)
		}
	}
	return nil
}

// Validate checks starter labels. Label is required when the block exists.
func (s *StarterLabels) Validate() error {
	s.Label = strings.TrimSpace(s.Label)
	s.Persona = strings.TrimSpace(s.Persona)
	if s.Label == "" {
		return fmt.Errorf("label is required")
	}
	if len(s.Label) > maxStarterLabelLen {
		return fmt.Errorf("label must be at most %d characters", maxStarterLabelLen)
	}
	if s.Persona != "" && !contains(starterPersonas, s.Persona) {
		return fmt.Errorf("unknown persona %q (want one of %s)", s.Persona, strings.Join(starterPersonas, ", "))
	}
	if s.Order < 0 {
		return fmt.Errorf("order must not be negative")
	}
	return nil
}

// Validate checks the structural delivery defaults. It does not resolve the
// approval policy reference; see ResolveApprovalPolicy.
func (d *DeliveryDefaults) Validate() error {
	d.WorkMode = strings.TrimSpace(d.WorkMode)
	d.ApprovalPolicyRef = strings.TrimSpace(d.ApprovalPolicyRef)
	d.ApprovalPosture = strings.TrimSpace(d.ApprovalPosture)
	d.Audit = strings.TrimSpace(d.Audit)
	if d.Audit == "" {
		d.Audit = AuditStandard
	}

	if !protocol.IsKnownWorkMode(d.WorkMode) {
		return fmt.Errorf("unknown work_mode %q (want one of %s)", d.WorkMode, strings.Join(protocol.WorkModes, ", "))
	}
	if d.ApprovalPolicyRef == "" {
		return fmt.Errorf("approval_policy_ref is required")
	}
	if !contains(ApprovalPostures, d.ApprovalPosture) {
		return fmt.Errorf("unknown approval_posture %q (want one of %s)", d.ApprovalPosture, strings.Join(ApprovalPostures, ", "))
	}
	if d.Audit != AuditStandard && d.Audit != AuditFull {
		return fmt.Errorf("unknown audit %q (want standard or full)", d.Audit)
	}
	if d.MaxTeamMembers < MinTeamMembers || d.MaxTeamMembers > MaxTeamMembersCap {
		return fmt.Errorf("max_team_members must be between %d and %d, got %d", MinTeamMembers, MaxTeamMembersCap, d.MaxTeamMembers)
	}
	if d.Outputs.ProofRequired && !d.Outputs.Retain {
		return fmt.Errorf("outputs.proof_required needs outputs.retain: true")
	}
	return d.validatePosture()
}

func (d *DeliveryDefaults) validatePosture() error {
	switch d.ApprovalPosture {
	case PostureApproveAllMutations:
		if d.WorkMode == protocol.WorkModeExecuteBounded || d.WorkMode == protocol.WorkModeResumeContinuity {
			return fmt.Errorf("approval_posture %s cannot use work_mode %s", d.ApprovalPosture, d.WorkMode)
		}
		if d.Audit != AuditFull {
			return fmt.Errorf("approval_posture %s requires audit: full", d.ApprovalPosture)
		}
	case PostureApproveExternalSends:
		if d.WorkMode == protocol.WorkModeExecuteBounded {
			return fmt.Errorf("approval_posture %s cannot use work_mode %s; use approve_high_risk", d.ApprovalPosture, d.WorkMode)
		}
	}
	return nil
}

// ResolveApprovalPolicy checks that approval_policy_ref names a group in the
// governance policy catalog (policy.yaml group names). A nil catalog means
// the policy could not be loaded.
func (d *DeliveryDefaults) ResolveApprovalPolicy(policyGroups map[string]struct{}) error {
	if policyGroups == nil {
		return fmt.Errorf("approval policy %q cannot be checked: governance policy is not loaded", d.ApprovalPolicyRef)
	}
	if _, ok := policyGroups[d.ApprovalPolicyRef]; !ok {
		return fmt.Errorf("approval policy %q is not a group in policy.yaml", d.ApprovalPolicyRef)
	}
	return nil
}

func contains(values []string, value string) bool {
	for _, v := range values {
		if v == value {
			return true
		}
	}
	return false
}
