package catalogue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

// ResolveActive selects the most specific active ConfigDocument revision for a
// custom profile, then falls back to an immutable built-in catalogue profile.
func (s *Service) ResolveActive(ctx context.Context, ref string, scope ProfileResolutionScope) (*AgentTemplate, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("profile reference is required")
	}
	tenantID := strings.TrimSpace(scope.TenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	store := configdocuments.NewStore(s.DB)
	for _, candidate := range profileScopeCandidates(scope) {
		record, err := store.GetActiveRevision(ctx, tenantID, protocol.ConfigDocumentKindWorkerProfile, ref, candidate)
		if errors.Is(err, configdocuments.ErrRevisionNotFound) {
			continue
		}
		if err != nil {
			return nil, fmt.Errorf("resolve active profile %s at %s scope: %w", ref, candidate.Kind, err)
		}
		return activeRevisionTemplate(*record)
	}
	return s.resolveLockedBuiltin(ctx, ref)
}

func profileScopeCandidates(scope ProfileResolutionScope) []protocol.ConfigDocumentScope {
	candidates := make([]protocol.ConfigDocumentScope, 0, 4)
	appendScope := func(kind protocol.ConfigDocumentScopeKind, ref string) {
		if ref = strings.TrimSpace(ref); ref != "" {
			candidates = append(candidates, protocol.ConfigDocumentScope{Kind: kind, Ref: ref})
		}
	}
	appendScope(protocol.ConfigDocumentScopeOperator, scope.OperatorRef)
	appendScope(protocol.ConfigDocumentScopeWorkspace, scope.WorkspaceRef)
	appendScope(protocol.ConfigDocumentScopeOrganization, scope.OrganizationRef)
	candidates = append(candidates, protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeBuiltIn})
	return candidates
}

func activeRevisionTemplate(record configdocuments.RevisionRecord) (*AgentTemplate, error) {
	compiled, err := configdocuments.CompileWorkerProfileDocument(record.Document)
	if err != nil {
		return nil, fmt.Errorf("compile active profile %s: %w", record.Document.Metadata.ID, err)
	}
	recordID, err := uuid.Parse(record.RecordID)
	if err != nil {
		return nil, fmt.Errorf("active profile %s has invalid record id: %w", record.Document.Metadata.ID, err)
	}
	compiled.Snapshot.RecordID = record.RecordID
	compiled.Snapshot.TenantID = record.TenantID
	compiled.Snapshot.Scope = record.Document.Metadata.Scope
	profile := compiled.Profile
	return &AgentTemplate{
		ID:                   recordID,
		ProfileKey:           record.Document.Metadata.ID,
		Name:                 record.Document.Metadata.Name,
		Role:                 profile.Role,
		Source:               "config_document",
		Locked:               true,
		SystemPrompt:         profile.SystemPrompt,
		Model:                profile.Model,
		Tools:                append([]string(nil), profile.CapabilityRefs...),
		CapabilityRefs:       append([]string(nil), profile.CapabilityRefs...),
		ContextBindings:      append([]protocol.AgentContextBinding(nil), profile.ContextBindings...),
		UsagePolicy:          profile.UsagePolicy,
		Inputs:               append([]string(nil), profile.Inputs...),
		Outputs:              append([]string(nil), profile.Outputs...),
		VerificationStrategy: profile.VerificationStrategy,
		VerificationRubric:   append([]string(nil), profile.VerificationRubric...),
		ValidationCommand:    profile.ValidationCommand,
		ProfileSnapshot:      &compiled.Snapshot,
		CreatedAt:            record.CreatedAt,
		UpdatedAt:            record.CreatedAt,
	}, nil
}

func (s *Service) resolveLockedBuiltin(ctx context.Context, ref string) (*AgentTemplate, error) {
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, role, system_prompt, model, tools, inputs, outputs,
		       verification_strategy, verification_rubric, validation_command,
		       profile_key, description, source, locked, capability_refs, context_bindings, usage_policy,
		       created_at, updated_at
		FROM agent_catalogue
		WHERE (profile_key = $1 OR id::text = $1)
		  AND locked = TRUE
		  AND source = 'built_in'
	`, ref)
	profile, err := scanAgent(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %s has no active scoped revision or locked built-in", ErrProfileNotFound, ref)
	}
	if err != nil {
		return nil, fmt.Errorf("resolve locked built-in profile %s: %w", ref, err)
	}
	return profile, nil
}
