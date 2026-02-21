-- Migration 022: Intent Proofs & Confirm Tokens (CE-1: Orchestration Templates)
--
-- Stores intent proof bundles for Chat-to-Proposal executions.
-- Confirm tokens gate state mutation: no token = no commit.
-- Audit events use the existing log_entries table (level='audit').

CREATE TABLE IF NOT EXISTS intent_proofs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    template_id TEXT NOT NULL,                           -- 'chat-to-answer' or 'chat-to-proposal'
    resolved_intent TEXT NOT NULL,                        -- classified intent text
    user_confirmation_token TEXT,                         -- UUID set when user confirms
    permission_check TEXT NOT NULL DEFAULT 'pass',        -- 'pass' or 'fail'
    policy_decision TEXT NOT NULL DEFAULT 'allow',        -- 'allow', 'deny', 'require_approval'
    scope_validation JSONB,                              -- {tools, affected_resources, risk_level}
    audit_event_id UUID,                                 -- references log_entries(id)
    mission_id UUID REFERENCES missions(id) ON DELETE SET NULL,
    status TEXT NOT NULL DEFAULT 'pending',               -- pending, confirmed, denied
    created_at TIMESTAMPTZ DEFAULT NOW(),
    confirmed_at TIMESTAMPTZ,
    expires_at TIMESTAMPTZ                               -- proof expiry (matches token TTL)
);

CREATE INDEX IF NOT EXISTS idx_intent_proofs_status ON intent_proofs(status);
CREATE INDEX IF NOT EXISTS idx_intent_proofs_mission ON intent_proofs(mission_id);
CREATE INDEX IF NOT EXISTS idx_intent_proofs_token ON intent_proofs(user_confirmation_token)
    WHERE user_confirmation_token IS NOT NULL;

-- Confirm tokens: single-use, TTL-bound tokens that gate state mutation.
-- Generated during negotiate, consumed during commit.
CREATE TABLE IF NOT EXISTS confirm_tokens (
    token UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    intent_proof_id UUID NOT NULL REFERENCES intent_proofs(id) ON DELETE CASCADE,
    template_id TEXT NOT NULL,
    consumed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    expires_at TIMESTAMPTZ NOT NULL,                     -- 15 min from creation
    consumed_at TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_confirm_tokens_proof ON confirm_tokens(intent_proof_id);
CREATE INDEX IF NOT EXISTS idx_confirm_tokens_expires ON confirm_tokens(expires_at)
    WHERE consumed = FALSE;
