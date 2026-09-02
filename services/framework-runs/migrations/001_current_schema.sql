CREATE TABLE framework_runs_schema_metadata (
    singleton BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (singleton),
    schema_version INTEGER NOT NULL CHECK (schema_version = 1),
    schema_contract TEXT NOT NULL CHECK (schema_contract = 'framework-runs-v1'),
    installed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE runs (
    run_id VARCHAR(128) PRIMARY KEY,
    idempotency_key VARCHAR(128) NOT NULL UNIQUE,
    request_digest CHAR(64) NOT NULL,
    request_json JSONB NOT NULL,
    snapshot_json JSONB NOT NULL,
    status TEXT NOT NULL,
    version BIGINT NOT NULL CHECK (version >= 1),
    next_sequence BIGINT NOT NULL CHECK (next_sequence >= 2),
    pending_command_id VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    CONSTRAINT runs_request_digest_hex CHECK (request_digest ~ '^[0-9a-f]{64}$'),
    CONSTRAINT runs_status_exact CHECK (status IN ('accepted','running','approval_needed','completed','failed','cancelled')),
    CONSTRAINT runs_request_identity CHECK (request_json->>'run_id' = run_id),
    CONSTRAINT runs_snapshot_identity CHECK (snapshot_json->>'run_id' = run_id),
    CONSTRAINT runs_snapshot_version CHECK ((snapshot_json->>'version')::BIGINT = version),
    CONSTRAINT runs_snapshot_status CHECK (snapshot_json->>'status' = status)
);

CREATE TABLE run_events (
    run_id VARCHAR(128) NOT NULL REFERENCES runs(run_id) ON DELETE RESTRICT,
    sequence BIGINT NOT NULL CHECK (sequence >= 1),
    event_id VARCHAR(128) NOT NULL UNIQUE,
    payload_digest CHAR(64) NOT NULL,
    payload_json JSONB NOT NULL,
    kind TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (run_id, sequence),
    CONSTRAINT run_events_digest_hex CHECK (payload_digest ~ '^[0-9a-f]{64}$'),
    CONSTRAINT run_events_identity CHECK (payload_json->>'run_id' = run_id),
    CONSTRAINT run_events_sequence_identity CHECK ((payload_json->>'sequence')::BIGINT = sequence),
    CONSTRAINT run_events_kind_status CHECK (
        (kind='accepted' AND status='accepted') OR
        (kind='progress' AND status='running') OR
        (kind='approval_needed' AND status='approval_needed') OR
        (kind='completed' AND status='completed') OR
        (kind='failed' AND status='failed') OR
        (kind='cancelled' AND status='cancelled')
    )
);

CREATE TABLE run_commands (
    command_id VARCHAR(128) PRIMARY KEY,
    run_id VARCHAR(128) NOT NULL REFERENCES runs(run_id) ON DELETE RESTRICT,
    kind TEXT NOT NULL CHECK (kind IN ('start','approve','deny','stop')),
    payload_digest CHAR(64) NOT NULL CHECK (payload_digest ~ '^[0-9a-f]{64}$'),
    payload_json JSONB NOT NULL,
    expected_version BIGINT NOT NULL CHECK (expected_version >= 0),
    approval_id VARCHAR(128),
    state TEXT NOT NULL CHECK (state IN ('pending','leased','applied','failed')),
    attempts INTEGER NOT NULL DEFAULT 0 CHECK (attempts >= 0),
    available_at TIMESTAMPTZ NOT NULL,
    lease_owner TEXT,
    lease_token VARCHAR(128),
    lease_generation BIGINT NOT NULL DEFAULT 0 CHECK (lease_generation >= 0),
    lease_until TIMESTAMPTZ,
    receipt_json JSONB NOT NULL,
    last_error TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);

CREATE UNIQUE INDEX run_commands_one_pending_per_run
    ON run_commands(run_id) WHERE state IN ('pending','leased');
CREATE INDEX run_commands_claim_order
    ON run_commands(available_at, created_at) WHERE state IN ('pending','leased');

ALTER TABLE runs ADD CONSTRAINT runs_pending_command_id_fkey
    FOREIGN KEY (pending_command_id) REFERENCES run_commands(command_id)
    DEFERRABLE INITIALLY DEFERRED;

CREATE TABLE run_approvals (
    approval_id VARCHAR(128) PRIMARY KEY,
    run_id VARCHAR(128) NOT NULL REFERENCES runs(run_id) ON DELETE RESTRICT,
    request_json JSONB NOT NULL,
    state TEXT NOT NULL CHECK (state IN ('pending','approved','denied')),
    decision_command_id VARCHAR(128) REFERENCES run_commands(command_id),
    decision_digest CHAR(64),
    actor_id VARCHAR(128),
    created_at TIMESTAMPTZ NOT NULL,
    decided_at TIMESTAMPTZ,
    CONSTRAINT run_approvals_decision_digest_check CHECK (
        decision_digest IS NULL OR decision_digest ~ '^[0-9a-f]{64}$'
    ),
    CONSTRAINT run_approvals_request_identity CHECK (request_json->>'id' = approval_id)
);

CREATE UNIQUE INDEX run_approvals_one_pending_per_run
    ON run_approvals(run_id) WHERE state = 'pending';

CREATE TABLE candidate_manifests (
    run_id VARCHAR(128) NOT NULL REFERENCES runs(run_id) ON DELETE RESTRICT,
    output_id VARCHAR(128) NOT NULL,
    candidate_uri TEXT NOT NULL,
    kind TEXT NOT NULL,
    name TEXT,
    content_type TEXT NOT NULL,
    size_bytes BIGINT NOT NULL CHECK (size_bytes >= 0),
    sha256 CHAR(64) NOT NULL CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    metadata JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL,
    PRIMARY KEY (run_id, output_id),
    UNIQUE (candidate_uri),
    CONSTRAINT candidate_uri_scoped CHECK (
        left(candidate_uri, length('candidate://' || run_id || '/')) = 'candidate://' || run_id || '/'
    )
);

CREATE FUNCTION reject_framework_runs_immutable_change() RETURNS trigger
LANGUAGE plpgsql AS $$
BEGIN
    RAISE EXCEPTION 'immutable framework Runs journal row';
END;
$$;

CREATE TRIGGER run_events_immutable
BEFORE UPDATE OR DELETE ON run_events
FOR EACH ROW EXECUTE FUNCTION reject_framework_runs_immutable_change();

CREATE TRIGGER candidate_manifests_immutable
BEFORE UPDATE OR DELETE ON candidate_manifests
FOR EACH ROW EXECUTE FUNCTION reject_framework_runs_immutable_change();

INSERT INTO framework_runs_schema_metadata (singleton, schema_version, schema_contract)
VALUES (TRUE, 1, 'framework-runs-v1');
