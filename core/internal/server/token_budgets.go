package server

import (
	"encoding/json"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

// tokenBudgetProvider is one provider row of GET /api/v1/cognitive/budgets.
type tokenBudgetProvider struct {
	ProviderID  string `json:"provider_id"`
	ModelID     string `json:"model_id"`
	BudgetClass string `json:"budget_class"`
	ClassSource string `json:"class_source"` // "explicit" | "data_boundary"
}

// tokenBudgetOverrideBody is the PUT body. Pointers tell an explicit 0 from
// an absent field; unknown fields (for example a spoofed actor) are ignored
// because the actor always comes from the authenticated request.
type tokenBudgetOverrideBody struct {
	PerExecution *int `json:"per_execution"`
	PerRun       *int `json:"per_run"`
	PerTeamDay   *int `json:"per_team_day"`
	PerAgentDay  *int `json:"per_agent_day"`
	WarnPct      *int `json:"warn_pct"`
}

// codeTokenBudgetUsageForbidden is the 403 for a usage read the caller cannot
// prove they own (B1R-B).
const codeTokenBudgetUsageForbidden = "token_budget_usage_forbidden"

var tokenBudgetUsageForbiddenCopy = roleBlockerText{
	User: blockerText{"You can only see token usage for teams you belong to.",
		"Ask an admin, or someone on that team, for its usage."},
	Admin: blockerText{Action: "Reading any team, agent or run usage needs cognitive:read or cognitive:write on your admin account."},
}

// tokenBudgetTeamTenantSQL proves team membership from persisted state only
// and returns the tenant it was proven in (” proves nothing): the team's
// active ownership binding (runtime_team_manifests owner account + owner
// group, not revoked) and an active, unexpired org membership of the caller in
// that group, with the user and account both active and the account in the
// team's tenant. A user has one account, so at most one tenant matches.
const tokenBudgetTeamTenantSQL = `SELECT COALESCE((SELECT t.tenant_id FROM runtime_team_manifests t
	JOIN accounts a ON a.id = t.owner_account_id AND a.tenant_id = t.tenant_id AND a.status = 'active'
	JOIN users u ON u.id = $2::uuid AND u.account_id = a.id AND u.status = 'active'
	JOIN org_memberships m ON m.account_id = a.id AND m.user_id = u.id AND m.group_id = t.owner_group_id
	WHERE t.team_id = $1 AND t.owner_group_id IS NOT NULL AND t.ownership_revoked_at IS NULL
		AND m.status = 'active' AND (m.expires_at IS NULL OR m.expires_at > NOW())
	LIMIT 1), '')`

// tokenBudgetUsageTenant writes the response and returns ok=false unless the
// caller may read this usage, else the tenant to read it in (B1R-C). A scoped
// root admin reads any ref in tenant "default" (identities carry no tenant
// yet); anyone else reads only a team they are proven to belong to, in the
// tenant of that membership. Agent and run usage have no ownership record, so
// only scoped root admins read them.
func (s *AdminServer) tokenBudgetUsageTenant(w http.ResponseWriter, r *http.Request, scope, ref string) (string, bool) {
	if cognitiveFullView(r) {
		return cognitive.DefaultBudgetTenant, true
	}
	if scope == protocol.TokenBudgetScopeTeamDay {
		tenant, err := s.tokenBudgetTeamTenant(r, ref)
		if err != nil {
			log.Printf("token budget usage: team membership check failed: %v", err)
			respondAPIError(w, "Team membership could not be checked; no usage was returned", http.StatusServiceUnavailable)
			return "", false
		}
		if tenant != "" {
			return tenant, true
		}
	}
	respondBlockerText(w, r, http.StatusForbidden, codeTokenBudgetUsageForbidden, tokenBudgetUsageForbiddenCopy,
		"no ownership record proves this caller may read "+scope+" usage", map[string]string{"scope": scope})
	return "", false
}

// tokenBudgetTeamTenant fails closed: no database or a non-UUID principal
// proves nothing ("").
func (s *AdminServer) tokenBudgetTeamTenant(r *http.Request, teamID string) (string, error) {
	db := s.getDB()
	userID, err := uuid.Parse(IdentityFromContext(r.Context()).UserID)
	if db == nil || err != nil {
		return "", nil
	}
	var tenant string
	err = db.QueryRowContext(r.Context(), tokenBudgetTeamTenantSQL, teamID, userID.String()).Scan(&tenant)
	return strings.TrimSpace(tenant), err
}

func (s *AdminServer) tokenBudgetGovernor(w http.ResponseWriter) (*cognitive.BudgetGovernor, bool) {
	if s.Cognitive == nil || s.Cognitive.Budgets == nil {
		respondAPIError(w, "Token budgets are unavailable: the cognitive engine is offline", http.StatusServiceUnavailable)
		return nil, false
	}
	return s.Cognitive.Budgets, true
}

// defaultBudgetClass is the class of the provider behind the chat profile.
func (s *AdminServer) budgetClassForProfile(profile string) string {
	_, cfg, _ := s.Cognitive.ProfileProviderSnapshot(profile)
	return cognitive.ResolveBudgetClass(cfg)
}

// GET /api/v1/cognitive/budgets — the effective policy for any signed-in
// user; override provenance is shown to admins only, and provider/model
// identity only to root admins with cognitive:read or cognitive:write.
func (s *AdminServer) HandleGetTokenBudgets(w http.ResponseWriter, r *http.Request) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	governor, ok := s.tokenBudgetGovernor(w)
	if !ok {
		return
	}
	policy := governor.Policy()
	fullView := cognitiveFullView(r)
	providers := []tokenBudgetProvider{}
	if config := s.Cognitive.ConfigSnapshot(); config != nil && fullView {
		for id, cfg := range config.Providers {
			source := "data_boundary"
			if protocol.IsTokenBudgetClass(strings.TrimSpace(cfg.BudgetClass)) {
				source = "explicit"
			}
			providers = append(providers, tokenBudgetProvider{ProviderID: id, ModelID: cfg.ModelID, BudgetClass: cognitive.ResolveBudgetClass(cfg), ClassSource: source})
		}
	}
	sort.Slice(providers, func(i, j int) bool { return providers[i].ProviderID < providers[j].ProviderID })
	period := protocol.TokenBudgetPeriodUTCDay
	if !governor.Durable() {
		period = protocol.TokenBudgetPeriodSinceRestart
	}
	data := map[string]any{
		"default_class": s.budgetClassForProfile(cognitive.DefaultExecutionProfileName),
		"class_order":   protocol.TokenBudgetClasses,
		"global":        policy.Global,
		"classes":       policy.Classes,
		"day_period":    period,
		"min_limit":     protocol.TokenBudgetMinLimit,
		"max_limit":     protocol.TokenBudgetMaxLimit,
		"can_edit":      identity.Role == "admin" && hasScope(identity, cognitiveWriteScope),
	}
	if fullView {
		data["providers"] = providers
	}
	if identity.Role == "admin" {
		data["overrides"] = policy.Overrides
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(data))
}

// GET /api/v1/cognitive/budgets/usage?team_id=|agent_id=|run_id= — exactly one,
// scoped and tenant-keyed by tokenBudgetUsageTenant.
func (s *AdminServer) HandleGetTokenBudgetUsage(w http.ResponseWriter, r *http.Request) {
	if IdentityFromContext(r.Context()) == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return
	}
	governor, ok := s.tokenBudgetGovernor(w)
	if !ok {
		return
	}
	query := r.URL.Query()
	scopes := map[string]string{"team_id": protocol.TokenBudgetScopeTeamDay, "agent_id": protocol.TokenBudgetScopeAgentDay, "run_id": protocol.TokenBudgetScopeRun}
	var scope, ref string
	for key := range query {
		if _, known := scopes[key]; !known || scope != "" {
			respondAPIError(w, "Give exactly one of team_id, agent_id or run_id", http.StatusBadRequest)
			return
		}
		scope, ref = scopes[key], strings.TrimSpace(query.Get(key))
	}
	if scope == "" || !protocol.ValidTokenBudgetRef(ref) {
		respondAPIError(w, "Give exactly one valid team_id, agent_id or run_id", http.StatusBadRequest)
		return
	}
	tenant, allowed := s.tokenBudgetUsageTenant(w, r, scope, ref)
	if !allowed {
		return
	}
	subject := cognitive.BudgetSubject{Class: s.budgetClassForProfile(cognitive.DefaultExecutionProfileName)}
	switch scope {
	case protocol.TokenBudgetScopeTeamDay:
		subject.TeamID = ref
	case protocol.TokenBudgetScopeAgentDay:
		subject.AgentID = ref
	}
	usage := governor.Usage(r.Context(), tenant, scope, ref, governor.Limits(subject))
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(usage))
}

// PUT /api/v1/cognitive/budgets/overrides/{level}/{ref}
func (s *AdminServer) HandlePutTokenBudgetOverride(w http.ResponseWriter, r *http.Request) {
	s.changeTokenBudgetOverride(w, r, true)
}

// DELETE /api/v1/cognitive/budgets/overrides/{level}/{ref} — idempotent.
func (s *AdminServer) HandleDeleteTokenBudgetOverride(w http.ResponseWriter, r *http.Request) {
	s.changeTokenBudgetOverride(w, r, false)
}

func (s *AdminServer) changeTokenBudgetOverride(w http.ResponseWriter, r *http.Request, put bool) {
	if _, ok := requireRootAdminScope(w, r, cognitiveWriteScope); !ok {
		return
	}
	governor, ok := s.tokenBudgetGovernor(w)
	if !ok {
		return
	}
	level, ref := r.PathValue("level"), r.PathValue("ref")
	if !validTokenBudgetOverrideTarget(level, ref) {
		respondAPIError(w, "Unknown override level or invalid ref (level is agent, team, profile or class)", http.StatusBadRequest)
		return
	}
	var limits protocol.TokenBudgetLimits
	if put {
		var body tokenBudgetOverrideBody
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			respondAPIError(w, "Invalid token budget override body", http.StatusBadRequest)
			return
		}
		var message string
		if limits, message = body.limits(); message != "" {
			respondAPIError(w, message, http.StatusBadRequest)
			return
		}
	}

	tokenBudgetWriteMu.Lock()
	defer tokenBudgetWriteMu.Unlock()
	current := governor.Policy()
	next := governor.Policy()
	entries := overrideEntries(&next, level)
	before, existed := (*entries)[ref]
	if put {
		if *entries == nil {
			*entries = map[string]protocol.TokenBudgetLimits{}
		}
		(*entries)[ref] = limits
		if message := s.tokenBudgetOverrideIssue(next, level, ref); message != "" {
			respondAPIError(w, message, http.StatusBadRequest)
			return
		}
	} else {
		delete(*entries, ref)
	}
	changed := put && (!existed || before != limits) || !put && existed
	db := s.getDB()
	if db == nil {
		respondAPIError(w, "Database unavailable: token budget overrides are stored as audited config revisions", http.StatusServiceUnavailable)
		return
	}
	status := "requested"
	if !changed {
		status = "noop"
	}
	var beforeValue, afterValue any
	if existed {
		beforeValue = before
	}
	if put {
		afterValue = limits
	}
	auditID, err := s.createAuditEvent(protocol.TemplateChatToProposal, tokenBudgetAuditSource, "Token budget override changed", attachActorIdentity(map[string]any{
		"actor": "operator", "user": auditUserLabelFromRequest(r), "action": "token_budget_changed",
		"level": level, "ref": ref, "before": beforeValue, "after": afterValue, "result_status": status,
	}, r))
	if err != nil || strings.TrimSpace(auditID) == "" {
		respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "token budget audit event could not be recorded", nil)
		return
	}
	result := map[string]any{"level": level, "ref": ref, "changed": changed, "before": beforeValue, "after": afterValue, "audit_event_id": auditID}
	if changed {
		recordID, err := persistTokenBudgetPolicy(r.Context(), db, auditActorIDFromRequest(r), auditID, next)
		if err != nil {
			log.Printf("token budget override not stored: %v", err)
			respondBlocker(w, r, http.StatusServiceUnavailable, codeServiceUnavailable, "token budget revision could not be stored", nil)
			return
		}
		governor.SetPolicy(next)
		result["record_id"] = recordID
	} else {
		next = current
	}
	result["effective"] = s.tokenBudgetEffective(next, level, ref)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

func (b tokenBudgetOverrideBody) limits() (protocol.TokenBudgetLimits, string) {
	var limits protocol.TokenBudgetLimits
	fields := []struct {
		value *int
		set   *int
	}{{b.PerExecution, &limits.PerExecution}, {b.PerRun, &limits.PerRun}, {b.PerTeamDay, &limits.PerTeamDay}, {b.PerAgentDay, &limits.PerAgentDay}, {b.WarnPct, &limits.WarnPct}}
	for _, field := range fields {
		if field.value == nil {
			continue
		}
		if *field.value <= 0 {
			return limits, "Token budget values must be positive integers; there is no unlimited value"
		}
		*field.set = *field.value
	}
	if limits == (protocol.TokenBudgetLimits{}) {
		return limits, "An override must set at least one of per_execution, per_run, per_team_day, per_agent_day or warn_pct"
	}
	if issues := (protocol.TokenBudgetPolicySpec{Global: protocol.DefaultTokenBudgetPolicySpec().Global,
		Overrides: protocol.TokenBudgetOverrides{Team: map[string]protocol.TokenBudgetLimits{"check": limits}}}).Issues(); len(issues) > 0 {
		return limits, issues[0].Message
	}
	return limits, ""
}

func validTokenBudgetOverrideTarget(level, ref string) bool {
	switch level {
	case cognitive.BudgetLevelAgent, cognitive.BudgetLevelTeam, cognitive.BudgetLevelProfile:
		return protocol.ValidTokenBudgetRef(ref)
	case cognitive.BudgetLevelClass:
		return protocol.IsTokenBudgetClass(ref)
	}
	return false
}

func overrideEntries(spec *protocol.TokenBudgetPolicySpec, level string) *map[string]protocol.TokenBudgetLimits {
	switch level {
	case cognitive.BudgetLevelAgent:
		return &spec.Overrides.Agent
	case cognitive.BudgetLevelTeam:
		return &spec.Overrides.Team
	case cognitive.BudgetLevelProfile:
		return &spec.Overrides.Profile
	}
	return &spec.Overrides.Class
}

// tokenBudgetSubject is the representative subject an override applies to.
func (s *AdminServer) tokenBudgetSubject(level, ref string) cognitive.BudgetSubject {
	subject := cognitive.BudgetSubject{Class: s.budgetClassForProfile(cognitive.DefaultExecutionProfileName)}
	switch level {
	case cognitive.BudgetLevelAgent:
		subject.AgentID = ref
	case cognitive.BudgetLevelTeam:
		subject.TeamID = ref
	case cognitive.BudgetLevelProfile:
		subject.Profile, subject.Class = ref, s.budgetClassForProfile(ref)
	case cognitive.BudgetLevelClass:
		subject.Class = ref
	}
	return subject
}

// tokenBudgetOverrideIssue validates the merged policy: structure, bounds,
// and per_execution <= per_run <= per_team_day after resolution.
func (s *AdminServer) tokenBudgetOverrideIssue(spec protocol.TokenBudgetPolicySpec, level, ref string) string {
	if issues := spec.Issues(); len(issues) > 0 {
		return issues[0].Field + ": " + issues[0].Message
	}
	effective, _ := cognitive.ResolveBudgetLimits(spec, s.tokenBudgetSubject(level, ref))
	if message := protocol.TokenBudgetOrderingIssue(effective); message != "" {
		return "After merging, " + message
	}
	return ""
}

func (s *AdminServer) tokenBudgetEffective(spec protocol.TokenBudgetPolicySpec, level, ref string) map[string]any {
	limits, sources := cognitive.ResolveBudgetLimits(spec, s.tokenBudgetSubject(level, ref))
	return map[string]any{"limits": limits, "sources": sources}
}
