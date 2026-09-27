package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	tokenBudgetAuditSource = "cognitive-token-budget"
	tokenBudgetTenant      = "default"
	// EventTokenBudgetWarning is the mission event emitted once per scope
	// when usage crosses warn_pct.
	EventTokenBudgetWarning protocol.EventType = "token_budget.warning"
)

// tokenBudgetWriteMu serializes override PUT/DELETE from read through apply,
// so two writers never merge against the same stale policy.
var tokenBudgetWriteMu sync.Mutex

var tokenBudgetOverridesScope = protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeOperator, Ref: tokenBudgetTenant}

// Blocker copy for token_budget_exhausted: the user variant never names an
// API path; the admin variant names the override route.
const (
	tokenBudgetUserMessage = "This work stopped because it reached its token budget."
	tokenBudgetUserAction  = "Ask for smaller work, or ask an admin to raise the token budget, then try again."
	tokenBudgetAdminAction = "Raise the limit with PUT /api/v1/cognitive/budgets/overrides/{agent|team|profile|class}/{ref} " +
		"(root admin + cognitive:write), then retry. Nothing was marked complete."
)

// respondTokenBudgetBlocker writes the honest 429 budget stop with the scope,
// usage, limit and reset time (day scopes only).
func respondTokenBudgetBlocker(w http.ResponseWriter, r *http.Request, stop *cognitive.TokenBudgetExhaustedError, summary string) {
	if summary == "" {
		summary = tokenBudgetUserMessage
	}
	extra := map[string]string{"scope": stop.Scope, "used": strconv.Itoa(stop.Used), "limit": strconv.Itoa(stop.Limit)}
	if !stop.ResetsAt.IsZero() {
		extra["resets_at"] = stop.ResetsAt.UTC().Format(time.RFC3339)
	}
	text := roleBlockerText{User: blockerText{summary, tokenBudgetUserAction}, Admin: blockerText{Action: tokenBudgetAdminAction}}
	respondBlockerText(w, r, http.StatusTooManyRequests, cognitive.TokenBudgetExhaustedCode, text, "", extra)
}

// loadTokenBudgetPolicy returns the effective policy: the active built-in
// defaults revision (the compiled D3 table when none is seeded) plus the
// overrides of the active operator revision, if any.
func loadTokenBudgetPolicy(ctx context.Context, db *sql.DB) (protocol.TokenBudgetPolicySpec, error) {
	spec := protocol.DefaultTokenBudgetPolicySpec()
	if db == nil {
		return spec, nil
	}
	store := configdocuments.NewStore(db)
	builtIn, err := store.GetActiveRevision(ctx, tokenBudgetTenant, protocol.ConfigDocumentKindTokenBudgetPolicy,
		protocol.TokenBudgetDefaultsDocumentID, protocol.ConfigDocumentScope{Kind: protocol.ConfigDocumentScopeBuiltIn})
	switch {
	case err == nil:
		defaults, decodeErr := protocol.DecodeTokenBudgetPolicySpec(builtIn.Document.Spec)
		if decodeErr != nil {
			return spec, decodeErr
		}
		spec.Global, spec.Classes = defaults.Global, defaults.Classes
	case !errors.Is(err, configdocuments.ErrRevisionNotFound):
		return spec, err
	}
	active, err := store.GetActiveRevision(ctx, tokenBudgetTenant, protocol.ConfigDocumentKindTokenBudgetPolicy,
		protocol.TokenBudgetOverridesDocumentID, tokenBudgetOverridesScope)
	if errors.Is(err, configdocuments.ErrRevisionNotFound) {
		return spec, nil
	}
	if err != nil {
		return spec, err
	}
	stored, err := protocol.DecodeTokenBudgetPolicySpec(active.Document.Spec)
	if err != nil {
		return spec, err
	}
	spec.Overrides = stored.Overrides
	return spec, nil
}

// InstallTokenBudgets enables budget enforcement on the router at startup:
// durable period totals from token_usage_ledger when a database is present,
// since-restart counters (labelled so) otherwise.
func InstallTokenBudgets(ctx context.Context, router *cognitive.Router, db *sql.DB) {
	if router == nil {
		return
	}
	spec, err := loadTokenBudgetPolicy(ctx, db)
	if err != nil {
		// Fail toward the shipped defaults, never toward "unlimited".
		log.Printf("WARN: token budget overrides unreadable; enforcing built-in defaults: %v", err)
		spec = protocol.DefaultTokenBudgetPolicySpec()
	}
	var ledger cognitive.BudgetLedger
	if db != nil {
		ledger = cognitive.NewPostgresBudgetLedger(db)
	} else {
		log.Println("WARN: token budgets enforce since-restart period counters (database unavailable).")
	}
	router.Budgets = cognitive.NewBudgetGovernor(spec, ledger)
	log.Println("Token budgets enforced at Router.InferWithContract.")
}

// WireTokenBudgetWarnings persists each warn_pct crossing: a mission event
// when the execution has a run, otherwise an audit log entry.
func (s *AdminServer) WireTokenBudgetWarnings() {
	if s == nil || s.Cognitive == nil || s.Cognitive.Budgets == nil {
		return
	}
	s.Cognitive.Budgets.SetWarningSink(func(w cognitive.BudgetWarning) {
		payload := map[string]any{"action": "token_budget_warning", "scope": w.Scope, "ref": w.Ref, "used": w.Used, "limit": w.Limit,
			"warn_pct": w.WarnPct, "execution_id": w.ExecutionID, "execution_kind": w.ExecutionKind, "budget_class": w.BudgetClass}
		if s.Events != nil && w.RunID != "" {
			if _, err := s.Events.Emit(context.Background(), w.RunID, EventTokenBudgetWarning, protocol.SeverityWarn, w.AgentID, w.TeamID, payload); err == nil {
				return
			}
		}
		payload["run_id"], payload["team_id"], payload["agent_id"] = w.RunID, w.TeamID, w.AgentID
		if _, err := s.createAuditEvent(protocol.TemplateChatToProposal, tokenBudgetAuditSource, "Token budget warning", payload); err != nil {
			log.Printf("WARN: token budget warning not recorded: %v", err)
		}
	})
}

// persistTokenBudgetPolicy writes one operator TokenBudgetPolicy revision
// (source api) and activates it in one transaction bound to auditID.
func persistTokenBudgetPolicy(ctx context.Context, db *sql.DB, actorID, auditID string, spec protocol.TokenBudgetPolicySpec) (string, error) {
	raw, err := json.Marshal(spec)
	if err != nil {
		return "", err
	}
	document := protocol.ConfigDocument{
		APIVersion: protocol.ConfigDocumentAPIVersionV1,
		Kind:       protocol.ConfigDocumentKindTokenBudgetPolicy,
		Metadata: protocol.ConfigDocumentMetadata{
			ID: protocol.TokenBudgetOverridesDocumentID, Name: "Token budget overrides",
			Version: "r" + time.Now().UTC().Format("20060102T150405.000000000"), OwnerID: actorID,
			Scope: tokenBudgetOverridesScope, Enabled: true,
			Source:     protocol.ConfigDocumentSource{Kind: protocol.ConfigDocumentSourceAPI, Ref: "api:/api/v1/cognitive/budgets/overrides"},
			Governance: protocol.ConfigDocumentGovernance{RiskLevel: protocol.ConfigDocumentRiskMedium, ApprovalPosture: protocol.ApprovalPostureRequired},
		},
		Spec: raw,
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	result, err := configdocuments.NewStore(db).StoreAndActivateTokenBudgetPolicyTx(ctx, tx, tokenBudgetTenant, actorID, auditID, document)
	if err != nil {
		_ = tx.Rollback()
		return "", err
	}
	if err := tx.Commit(); err != nil {
		return "", fmt.Errorf("commit token budget policy: %w", err)
	}
	return result.ToRecordID, nil
}
