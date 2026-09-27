package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// Approved-plan tool scope (S7b item 1). A call the model chose (origin
// "agent") may only name a tool the originating agent declares; it is checked
// before a proposal is minted and again at execution, where it runs through
// that agent's ScopedToolExecutor so the S7 create_team child bound applies.
// A call Core inferred from the operator's own request (origin "core", or a
// legacy call with no origin) is limited to corePlanTools and never MCP.

const (
	codePlannedToolNotDeclared      = "planned_tool_not_declared"
	codePlannedToolScopeUnavailable = "planned_tool_scope_unavailable"
	codeCancellerNotProposer        = "canceller_not_proposer"

	somaOriginTeamID  = "admin-core"
	somaOriginAgentID = "admin"
)

// corePlanTools are the only tools Core's deterministic planners emit.
var corePlanTools = map[string]struct{}{
	"create_team": {}, "delegate_task": {}, "write_file": {}, "generate_image": {},
	"save_cached_image": {}, "store_config_document": {}, "activate_config_document": {},
}

func isAgentPlannedCall(call protocol.PlannedToolCall) bool {
	return call.Origin == protocol.PlannedCallOriginAgent
}

func corePlannedCallAllowed(call protocol.PlannedToolCall) bool {
	_, ok := corePlanTools[strings.TrimSpace(call.Name)]
	return ok && strings.TrimSpace(call.ToolRef) == ""
}

// plannedCallLabel names a call for denials, bounded like swarm denials.
func plannedCallLabel(call protocol.PlannedToolCall) string {
	label := strings.TrimSpace(call.ToolRef)
	if label == "" {
		label = strings.TrimSpace(call.Name)
	}
	if runes := []rune(label); len(runes) > 128 {
		label = string(runes[:128]) + "…"
	}
	return label
}

// originToolScope returns the originating agent's live declared-tool scope,
// or ok=false when the runtime cannot resolve that agent.
func (s *AdminServer) originToolScope(inner *swarm.CompositeToolExecutor, teamID, agentID string, names map[uuid.UUID]string) (*swarm.ScopedToolExecutor, bool) {
	if s == nil || s.Soma == nil {
		return nil, false
	}
	declared, ok := s.Soma.AgentDeclaredTools(teamID, agentID)
	if !ok {
		return nil, false
	}
	return swarm.NewScopedToolExecutor(inner, declared, names), true
}

// scopePlannedCallsOrRespond tags every call with its origin and refuses the
// whole plan, before any audit/proof/token, when a call falls outside scope.
func (s *AdminServer) scopePlannedCallsOrRespond(w http.ResponseWriter, r *http.Request, teamID, agentID string, planned []protocol.PlannedToolCall) ([]protocol.PlannedToolCall, bool) {
	var scoped *swarm.ScopedToolExecutor
	var denied []string
	for i, call := range planned {
		if !isAgentPlannedCall(call) {
			planned[i].Origin = protocol.PlannedCallOriginCore
			if !corePlannedCallAllowed(call) {
				denied = append(denied, plannedCallLabel(call))
			}
			continue
		}
		if scoped == nil {
			var ok bool
			if scoped, ok = s.originToolScope(swarm.NewCompositeToolExecutor(nil, s.plannedMCPToolExecutor()), teamID, agentID, nil); !ok {
				respondPlanScopeBlocker(w, http.StatusServiceUnavailable, cognitive.ExecutionAvailability{
					Code:              codePlannedToolScopeUnavailable,
					Summary:           "Soma couldn't check which tools this agent may use, so nothing was proposed.",
					RecommendedAction: "Try again in a moment. If it keeps happening, ask an admin to check the agent runtime.",
					AdminAction:       fmt.Sprintf("Agent %s/%s is not running, so its declared tools cannot be verified.", teamID, agentID),
				})
				return nil, false
			}
		}
		if !scoped.PermitsPlannedCall(r.Context(), call.Name, call.ToolRef) {
			denied = append(denied, plannedCallLabel(call))
		}
	}
	if len(denied) == 0 {
		return planned, true
	}
	log.Printf("S7b: plan from %s/%s refused; undeclared tools: %s", teamID, agentID, strings.Join(denied, ", "))
	_, _ = s.createAuditEvent(protocol.TemplateChatToProposal, agentID, "Planned tool not declared",
		attachActorIdentity(map[string]any{"action": "proposal_blocked", "result_status": "blocked", "agent_id": agentID,
			"team_id": teamID, "tools": denied, "reason": "not_declared"}, r))
	respondPlanScopeBlocker(w, http.StatusUnprocessableEntity, cognitive.ExecutionAvailability{
		Code:              codePlannedToolNotDeclared,
		Summary:           "Soma planned a tool this agent isn't allowed to use, so nothing was proposed and nothing ran.",
		RecommendedAction: "Ask again in different words, or ask an admin to give the agent that tool.",
		AdminAction:       "Undeclared tools: " + strings.Join(denied, ", ") + ". Add them to the agent's declared tools only if it should hold them.",
	})
	return nil, false
}

func respondPlanScopeBlocker(w http.ResponseWriter, status int, availability cognitive.ExecutionAvailability) {
	availability.Available = false
	respondAPIJSON(w, status, protocol.APIResponse{OK: false, Error: availability.Summary, Data: availability})
}

// approvedPlanToolGuard is the execution-time re-check for one approved plan.
type approvedPlanToolGuard struct {
	composite *swarm.CompositeToolExecutor
	scoped    *swarm.ScopedToolExecutor
	names     map[uuid.UUID]string // MCP server names the scoped executor trusts
}

// newApprovedPlanToolGuard checks every call before any call runs, so a plan
// holding one out-of-scope call executes nothing. The origin's declared tools
// are re-read now, so a revoked declaration blocks an older proposal.
func (s *AdminServer) newApprovedPlanToolGuard(ctx context.Context, composite *swarm.CompositeToolExecutor, scope *protocol.ScopeValidation) (*approvedPlanToolGuard, error) {
	guard := &approvedPlanToolGuard{composite: composite, names: map[uuid.UUID]string{}}
	for _, call := range scope.PlannedToolCalls {
		call = normalizePlannedToolCall(call)
		if !isAgentPlannedCall(call) {
			if !corePlannedCallAllowed(call) {
				return nil, &swarm.ToolNotPermittedError{Tool: plannedCallLabel(call)}
			}
			continue
		}
		if guard.scoped == nil {
			if s != nil && s.Soma != nil {
				for id, name := range s.Soma.MCPServerNames() {
					guard.names[id] = name
				}
			}
			scoped, ok := s.originToolScope(composite, scope.OriginTeamID, scope.OriginAgentID, guard.names)
			if !ok {
				return nil, &swarm.ToolNotPermittedError{Tool: plannedCallLabel(call)}
			}
			guard.scoped = scoped
		}
		if !guard.scoped.PermitsPlannedCall(ctx, call.Name, call.ToolRef) {
			return nil, &swarm.ToolNotPermittedError{Tool: plannedCallLabel(call)}
		}
	}
	return guard, nil
}

// executorFor returns the executor a call must run through: the originating
// agent's scope for agent calls (CallTool re-checks and hands create_team the
// caller scope), the composite for bounded Core calls.
func (g *approvedPlanToolGuard) executorFor(call protocol.PlannedToolCall) swarm.MCPToolExecutor {
	if isAgentPlannedCall(call) {
		return g.scoped
	}
	return g.composite
}
