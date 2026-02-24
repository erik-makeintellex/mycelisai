-- V7 Team B: Trigger Executions — audit log of every rule evaluation.
-- Depends on: trigger_rules (025), mission_runs (023), mission_events (024).

CREATE TABLE trigger_executions (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    rule_id         UUID NOT NULL REFERENCES trigger_rules(id) ON DELETE CASCADE,
    event_id        TEXT NOT NULL,                   -- mission_events.id that triggered evaluation
    run_id          TEXT,                             -- child run created (NULL if skipped)
    status          TEXT NOT NULL,                    -- "fired" | "skipped" | "proposed"
    skip_reason     TEXT,                             -- e.g. "cooldown", "recursion_limit", "concurrency_limit", "condition_failed"
    executed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX trigger_executions_rule_id ON trigger_executions(rule_id, executed_at DESC);
CREATE INDEX trigger_executions_event_id ON trigger_executions(event_id);
