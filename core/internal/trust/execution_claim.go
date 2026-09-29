package trust

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

// ExecutionClaim is the correlation a team trigger presents to ask for
// execution posture (F16b). Every field comes from the untrusted trigger except
// TeamID, which the receiving agent sets to its own team.
type ExecutionClaim struct {
	IntentProofID string
	ContractID    string
	RunID         string
	WorkItemID    string
	TeamID        string
	// IdempotencyKey must equal the delivery key of this planned call:
	// ConfirmedCallDeliveryKey(outbox key, WorkItemID) (TPD). The team's
	// durable command receipt accepts it once.
	IdempotencyKey string
}

// ConfirmedCallDeliveryKey is the delivery key of one planned call in a
// confirmed plan (TPD): the plan's dispatch outbox key (confirm-action:<proof>)
// bound to the call's work item. One plan may delegate several calls to one
// team, so a plan-wide key would let the team's receipt accept only the first.
// verifyExecutionClaimSQL derives the same value in SQL.
func ConfirmedCallDeliveryKey(outboxKey, workItemID string) string {
	return strings.TrimSpace(outboxKey) + ":" + strings.TrimSpace(workItemID)
}

// ErrExecutionClaimRejected means the claim does not name a confirmed,
// dispatched proof for this team; the caller must stay planning-only.
var ErrExecutionClaimRejected = errors.New("trust: execution claim not backed by a confirmed proof")

// QueryRower is the read seam VerifyExecutionClaim needs (*sql.DB, *sql.Tx).
type QueryRower interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

// verifyExecutionClaimSQL admits a claim only when the confirm-action
// transaction committed all of: the intent proof in status 'confirmed'; the
// execution contract for that proof and run; the confirmed-action dispatch
// outbox row for the same proof, contract and run; and, in the proof's
// correlated scope, a planned call whose context names this team, this proof
// and this work item.
//
// Freshness (F16c): the run is still live (pending or running; completed,
// degraded, failed and cancelled are terminal); the outbox row was released
// for dispatch and not dead-lettered (pending, executing or completed;
// staged, awaiting_handler and failed are not); and the claim's
// idempotency_key is exactly this planned call's delivery key (TPD: the outbox
// key, ':', the claimed work item; see ConfirmedCallDeliveryKey), which the
// planned call itself also carries and the team's durable command receipt
// accepts once. The plan-wide outbox key or another call's key never matches.
//
// intent_proofs.expires_at is deliberately NOT checked: it is the confirm-token
// TTL (15 minutes from proposal creation) and governs only whether a proposal
// may be confirmed. Checking it here would downgrade live confirmed runs
// (late confirm, retry, dispatch after restart) to planning-only. Replay is
// closed by single use (outbox key + receipt) and run state instead.
const verifyExecutionClaimSQL = `
SELECT EXISTS (
	SELECT 1
	FROM intent_proofs ip
	JOIN execution_contracts ec ON ec.intent_proof_id = ip.id
	JOIN mission_runs r ON r.id = ec.run_id
	JOIN execution_dispatch_outbox o ON o.intent_proof_id = ip.id
	WHERE ip.id = $1::uuid
	  AND ip.status = 'confirmed'
	  AND ec.id = $2::uuid
	  AND ec.run_id = $3::uuid
	  AND r.status IN ('pending', 'running') -- runs.StatusPending, runs.StatusRunning
	  AND o.dispatch_kind = 'confirmed_action_team_plan' -- server.confirmedActionDispatchKind
	  AND o.status IN ('pending', 'executing', 'completed') -- dispatchoutbox.Status*
	  AND $6::text = o.idempotency_key || ':' || $5::text -- trust.ConfirmedCallDeliveryKey
	  AND o.contract_id = ec.id
	  AND o.run_id = ec.run_id::text
	  AND EXISTS (
		SELECT 1
		FROM jsonb_array_elements(COALESCE(ip.scope_validation->'planned_tool_calls', '[]'::jsonb)) AS planned
		WHERE planned->'arguments'->'context'->>'team_id' = $4
		  AND planned->'arguments'->'context'->>'intent_proof_id' = ip.id::text
		  AND planned->'arguments'->'context'->>'work_item_id' = $5::text
		  AND planned->'arguments'->'context'->>'idempotency_key' = $6::text
	  )
)`

// VerifyExecutionClaim checks claim against the proof store. It returns nil
// only for a fully bound, fresh claim; a malformed claim, a missing store, a
// query error or no matching record all return an error (fail closed).
func VerifyExecutionClaim(ctx context.Context, db QueryRower, claim ExecutionClaim) error {
	for label, value := range map[string]string{"intent_proof_id": claim.IntentProofID, "contract_id": claim.ContractID, "run_id": claim.RunID} {
		if _, err := uuid.Parse(strings.TrimSpace(value)); err != nil {
			return fmt.Errorf("%w: %s is not a UUID", ErrExecutionClaimRejected, label)
		}
	}
	if strings.TrimSpace(claim.TeamID) == "" || strings.TrimSpace(claim.WorkItemID) == "" || strings.TrimSpace(claim.IdempotencyKey) == "" {
		return fmt.Errorf("%w: team_id, work_item_id and idempotency_key are required", ErrExecutionClaimRejected)
	}
	if db == nil {
		return fmt.Errorf("%w: proof store unavailable", ErrExecutionClaimRejected)
	}
	var ok bool
	err := db.QueryRowContext(ctx, verifyExecutionClaimSQL,
		strings.TrimSpace(claim.IntentProofID), strings.TrimSpace(claim.ContractID), strings.TrimSpace(claim.RunID),
		strings.TrimSpace(claim.TeamID), strings.TrimSpace(claim.WorkItemID), strings.TrimSpace(claim.IdempotencyKey)).Scan(&ok)
	if err != nil {
		return fmt.Errorf("%w: proof lookup failed: %v", ErrExecutionClaimRejected, err)
	}
	if !ok {
		return ErrExecutionClaimRejected
	}
	return nil
}
