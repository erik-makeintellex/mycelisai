package swarm

import (
	"context"
	"fmt"
	"strings"
)

// ContextClassRequiresApprovalCode is the blocker for an agent write of an
// organization-wide knowledge class without operator confirmation.
const ContextClassRequiresApprovalCode = "context_class_requires_approval"

// orgWideContextClasses shape every future Soma answer, so an agent may only
// write them inside an operator-confirmed invocation.
var orgWideContextClasses = map[string]bool{"company_knowledge": true, "soma_operating_context": true}

// requireConfirmedOrgWideWrite blocks org-wide class writes unless the tool
// runs inside a confirmed (non-planning) invocation.
func requireConfirmedOrgWideWrite(ctx context.Context, knowledgeClass string) error {
	class := strings.ToLower(strings.TrimSpace(knowledgeClass))
	if !orgWideContextClasses[class] {
		return nil
	}
	if inv, ok := ToolInvocationContextFromContext(ctx); ok && !inv.PlanningOnly {
		return nil
	}
	return fmt.Errorf("%s: writing %s needs operator confirmation; propose the save for approval instead", ContextClassRequiresApprovalCode, class)
}
