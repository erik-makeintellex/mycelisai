package catalogue

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
)

// Resolve finds a reusable profile by stable profile key or UUID text.
func (s *Service) Resolve(ctx context.Context, ref string) (*AgentTemplate, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return nil, fmt.Errorf("profile reference is required")
	}
	row := s.DB.QueryRowContext(ctx, `
		SELECT id, name, role, system_prompt, model, tools, inputs, outputs,
		       verification_strategy, verification_rubric, validation_command,
		       profile_key, description, source, locked, capability_refs, context_bindings, usage_policy,
		       created_at, updated_at
		FROM agent_catalogue
		WHERE profile_key = $1 OR id::text = $1
	`, ref)
	profile, err := scanAgent(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("profile %s not found", ref)
		}
		return nil, fmt.Errorf("resolve profile %s: %w", ref, err)
	}
	return profile, nil
}
