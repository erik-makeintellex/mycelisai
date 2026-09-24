"""Canonical current-schema installer and compatibility contract."""

CANONICAL_SCHEMA_NAME = "001_current_schema.sql"
PUBLIC_SCHEMA_NONEMPTY_SQL = (
    "SELECT 1 WHERE EXISTS (SELECT 1 FROM information_schema.tables "
    "WHERE table_schema = 'public' AND table_type = 'BASE TABLE');"
)

SCHEMA_COMPATIBILITY_CHECKS = (
    ("pgvector extension", "SELECT 1 FROM pg_extension WHERE extname = 'vector';"),
    ("semantic context vectors", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'context_vectors';"),
    ("durable agent memory", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'agent_memories';"),
    ("retained artifacts", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'artifacts';"),
    ("temporary continuity", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'temp_memory_channels';"),
    ("managed exchange channels", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'exchange_channels';"),
    ("managed exchange items", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'exchange_items';"),
    ("conversation templates", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'conversation_templates';"),
    ("nodes.type column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'nodes' AND column_name = 'type';"),
    ("nodes.specs column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'nodes' AND column_name = 'specs';"),
    ("intent_proofs table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'intent_proofs';"),
    ("confirm_tokens table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'confirm_tokens';"),
    ("conversation_turns table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'conversation_turns';"),
    ("collaboration_groups table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'collaboration_groups';"),
    ("collaboration_groups workspace_folder column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'collaboration_groups' AND column_name = 'workspace_folder';"),
    ("capability_manifests table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'capability_manifests';"),
    ("capability_manifests health column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'capability_manifests' AND column_name = 'health';"),
    ("execution_contracts table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'execution_contracts';"),
    ("proof_artifacts table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'proof_artifacts';"),
    ("team_work_items table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'team_work_items';"),
    ("team_work_items work_intent column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'team_work_items' AND column_name = 'work_intent' AND data_type = 'jsonb';"),
    ("team_work_items execution_mode column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'team_work_items' AND column_name = 'execution_mode' AND data_type = 'text';"),
    ("team_work_items recovery_deadline_at column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'team_work_items' AND column_name = 'recovery_deadline_at' AND data_type = 'timestamp with time zone';"),
    ("team_interactions table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'team_interactions';"),
    ("team_status_events table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'team_status_events';"),
    ("team_status_events work_intent column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'team_status_events' AND column_name = 'work_intent' AND data_type = 'jsonb';"),
    ("team_status_events execution_mode column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'team_status_events' AND column_name = 'execution_mode' AND data_type = 'text';"),
    ("outcome_projects table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'outcome_projects';"),
    ("outcome_projects run_id text column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'outcome_projects' AND column_name = 'run_id' AND data_type = 'text';"),
    ("outcome_projects intent_proof_id text column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'outcome_projects' AND column_name = 'intent_proof_id' AND data_type = 'text';"),
    ("team_registry_entries table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'team_registry_entries';"),
    ("trigger_rules schedule columns", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'trigger_rules' AND column_name = 'trigger_kind';"),
    ("search_sources table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'search_sources';"),
    ("input_sources table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'input_sources';"),
    ("execution_dispatch_outbox table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'execution_dispatch_outbox';"),
    ("operator_sse_events table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'operator_sse_events';"),
    ("team_signal_receipts table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'team_signal_receipts';"),
    ("agent_catalogue profile_key column", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'agent_catalogue' AND column_name = 'profile_key';"),
    ("qa_fixture_scopes table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'qa_fixture_scopes';"),
    ("qa_fixture_resources table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'qa_fixture_resources';"),
    ("qa_fixture resource ownership index", "SELECT 1 FROM pg_indexes WHERE schemaname = 'public' AND indexname = 'uq_qa_fixture_resource_claim';"),
    ("purged QA fixture claims released", "SELECT 1 WHERE NOT EXISTS (SELECT 1 FROM qa_fixture_resources r JOIN qa_fixture_scopes s ON s.id=r.scope_id WHERE s.status='purged');"),
    ("config_documents table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'config_documents';"),
    ("config_document_activations table", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'config_document_activations' AND column_name = 'kind';"),
    ("config_document_activation_history table", "SELECT 1 FROM information_schema.columns WHERE table_schema = 'public' AND table_name = 'config_document_activation_history' AND column_name = 'kind';"),
    ("config_document fixture ownership", "SELECT 1 FROM pg_constraint WHERE conname = 'chk_qa_fixture_resource_kind' AND pg_get_constraintdef(oid) LIKE '%config_document%';"),
    ("runtime_team_manifests table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'runtime_team_manifests';"),
    ("code_context_sources table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'code_context_sources';"),
    ("code_context_snapshots table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'code_context_snapshots';"),
    ("code_context_files table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'code_context_files';"),
    ("code_context_symbols table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'code_context_symbols';"),
    ("code_context_edges table", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'code_context_edges';"),
)

# Preserve the accepted pre-G4 compatibility gate for the bounded additive upgrade.
BASE_SCHEMA_COMPATIBILITY_CHECKS = SCHEMA_COMPATIBILITY_CHECKS
SCHEMA_COMPATIBILITY_CHECKS += (
    ("effect grants", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'execution_effect_grants';"),
    ("durable invocations", "SELECT 1 FROM information_schema.tables WHERE table_schema = 'public' AND table_name = 'execution_invocations';"),
)

SCHEMA_COMPATIBILITY_CHECKS += (
    ("effect grant state", "SELECT 1 FROM information_schema.tables WHERE table_schema='public' AND table_name='execution_effect_grant_state';"),
    ("invocation idempotency", "SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('public.execution_invocations') AND conname='uq_effect_invocation_key' AND contype='u';"),
    ("invocation recovery index", "SELECT 1 FROM pg_indexes WHERE schemaname='public' AND tablename='execution_invocations' AND indexname='idx_effect_invocations_recovery';"),
)
for _table, _columns in (
    ("execution_effect_grants", ("id", "intent_proof_id", "execution_contract_id", "account_id", "user_id", "group_id", "membership_id", "capability_id", "authority_snapshot", "authority_digest", "binding_snapshot", "binding_digest", "input_snapshot", "input_digest", "unit_budget", "expires_at", "digest", "created_at")),
    ("execution_effect_grant_state", ("grant_id", "reserved_units", "revoked_at", "revoked_by", "revoke_reason")),
    ("execution_invocations", ("id", "grant_id", "account_id", "user_id", "group_id", "capability_id", "grant_digest", "authority_snapshot", "authority_digest", "binding_snapshot", "binding_digest", "input_snapshot", "input_digest", "idempotency_key", "reservation_units", "owner_token", "generation", "attempt", "state", "lease_until", "executing_at", "observed_at", "result", "reconciliation", "created_at", "updated_at")),
):
    _names = ",".join("'" + name + "'" for name in _columns)
    SCHEMA_COMPATIBILITY_CHECKS += ((f"{_table} receipt columns", f"SELECT 1 WHERE (SELECT count(*) FROM information_schema.columns WHERE table_schema='public' AND table_name='{_table}' AND column_name IN ({_names}))={len(_columns)};"),)

SCHEMA_COMPATIBILITY_CHECKS += (("immutable effect grant trigger", "SELECT 1 FROM pg_trigger WHERE tgrelid=to_regclass('public.execution_effect_grants') AND tgname='trg_execution_effect_grants_immutable' AND tgenabled='O';"),)

SCHEMA_COMPATIBILITY_CHECKS += (("immutable invocation identity trigger", "SELECT 1 FROM pg_trigger WHERE tgrelid=to_regclass('public.execution_invocations') AND tgname='trg_execution_invocations_identity_immutable' AND tgenabled='O';"),)

# G4/E10 is a supported retained baseline for the bounded C2a ALTER transaction.
G4_SCHEMA_COMPATIBILITY_CHECKS = SCHEMA_COMPATIBILITY_CHECKS
TEAM_OWNERSHIP_COLUMNS = (
    ("owner_account_id", "uuid"), ("owner_group_id", "uuid"),
    ("ownership_provisioned_by", "uuid"),
    ("ownership_provisioned_at", "timestamp with time zone"),
    ("ownership_revoked_by", "uuid"),
    ("ownership_revoked_at", "timestamp with time zone"),
)
for _column, _type in TEAM_OWNERSHIP_COLUMNS:
    SCHEMA_COMPATIBILITY_CHECKS += ((f"team ownership {_column}",
        "SELECT 1 FROM information_schema.columns WHERE table_schema='public' "
        "AND table_name='runtime_team_manifests' "
        f"AND column_name='{_column}' AND data_type='{_type}' AND is_nullable='YES';"),)
SCHEMA_COMPATIBILITY_CHECKS += (
    ("team owner account/group FK", "SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('public.runtime_team_manifests') AND confrelid=to_regclass('public.groups') AND conname='fk_runtime_team_owner' AND contype='f' AND confmatchtype='f' AND confdeltype='r' AND convalidated AND pg_get_constraintdef(oid) LIKE 'FOREIGN KEY (owner_account_id, owner_group_id) REFERENCES groups(account_id, id)%';"),
)
# Exact pg16 constraint definitions fail closed on same-name weakened constraints.
_TEAM_EVIDENCE_CONSTRAINTS = (
    ("chk_runtime_team_provisioned", "CHECK ((((owner_account_id IS NULL) AND (owner_group_id IS NULL) AND (ownership_provisioned_by IS NULL) AND (ownership_provisioned_at IS NULL) AND (ownership_revoked_by IS NULL) AND (ownership_revoked_at IS NULL)) OR ((owner_account_id IS NOT NULL) AND (owner_group_id IS NOT NULL) AND (ownership_provisioned_by IS NOT NULL) AND (ownership_provisioned_at IS NOT NULL))))"),
    ("chk_runtime_team_revoked", "CHECK ((((ownership_revoked_by IS NULL) AND (ownership_revoked_at IS NULL)) OR ((ownership_revoked_by IS NOT NULL) AND (ownership_revoked_at IS NOT NULL) AND (ownership_revoked_at >= ownership_provisioned_at))))"),
    ("runtime_team_manifests_ownership_provisioned_by_fkey", "FOREIGN KEY (ownership_provisioned_by) REFERENCES users(id) ON DELETE RESTRICT"),
    ("runtime_team_manifests_ownership_revoked_by_fkey", "FOREIGN KEY (ownership_revoked_by) REFERENCES users(id) ON DELETE RESTRICT"),
)
for _name, _definition in _TEAM_EVIDENCE_CONSTRAINTS:
    SCHEMA_COMPATIBILITY_CHECKS += ((_name,
        "SELECT 1 FROM pg_constraint WHERE conrelid=to_regclass('public.runtime_team_manifests') "
        f"AND conname='{_name}' AND convalidated AND NOT condeferrable "
        f"AND pg_get_constraintdef(oid)='{_definition}';"),)
