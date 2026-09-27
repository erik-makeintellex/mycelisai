package cognitive

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

const budgetLedgerTenant = "default"

// PostgresBudgetLedger is the append-only token_usage_ledger store.
type PostgresBudgetLedger struct {
	db *sql.DB
}

func NewPostgresBudgetLedger(db *sql.DB) *PostgresBudgetLedger {
	return &PostgresBudgetLedger{db: db}
}

func (l *PostgresBudgetLedger) PeriodTotal(ctx context.Context, scope, ref string, since time.Time) (int, bool, error) {
	if l == nil || l.db == nil {
		return 0, false, fmt.Errorf("token usage ledger: database unavailable")
	}
	var column string
	switch scope {
	case protocol.TokenBudgetScopeRun:
		column = "run_id"
	case protocol.TokenBudgetScopeTeamDay:
		column = "team_id"
	case protocol.TokenBudgetScopeAgentDay:
		column = "agent_id"
	default:
		return 0, false, fmt.Errorf("token usage ledger: unknown scope %q", scope)
	}
	var total int64
	var unreported bool
	err := l.db.QueryRowContext(ctx, `
		SELECT COALESCE(SUM(total_tokens), 0), COALESCE(bool_or(NOT usage_reported), false)
		FROM token_usage_ledger
		WHERE tenant_id = $1 AND `+column+` = $2 AND occurred_at >= $3
		  AND outcome = 'charged' AND execution_kind <> 'system'`,
		budgetLedgerTenant, ref, since.UTC()).Scan(&total, &unreported)
	if err != nil {
		return 0, false, fmt.Errorf("token usage ledger: period total: %w", err)
	}
	return int(total), unreported, nil
}

func (l *PostgresBudgetLedger) Append(ctx context.Context, e TokenLedgerEntry) error {
	if l == nil || l.db == nil {
		return fmt.Errorf("token usage ledger: database unavailable")
	}
	_, err := l.db.ExecContext(ctx, `
		INSERT INTO token_usage_ledger
			(tenant_id, occurred_at, execution_id, execution_kind, run_id, team_id, agent_id,
			 provider_id, model_id, budget_class, prompt_tokens, completion_tokens, total_tokens,
			 usage_reported, outcome)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)`,
		budgetLedgerTenant, e.OccurredAt, e.ExecutionID, e.ExecutionKind, e.RunID, e.TeamID, e.AgentID,
		e.ProviderID, e.ModelID, e.BudgetClass, e.PromptTokens, e.CompletionTokens, e.TotalTokens,
		e.UsageReported, e.Outcome)
	if err != nil {
		return fmt.Errorf("token usage ledger: append: %w", err)
	}
	return nil
}
