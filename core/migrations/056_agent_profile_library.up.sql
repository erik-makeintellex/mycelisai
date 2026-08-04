-- 056: reusable worker profiles with governed capability and context bindings.

ALTER TABLE agent_catalogue
    ADD COLUMN IF NOT EXISTS profile_key TEXT,
    ADD COLUMN IF NOT EXISTS description TEXT,
    ADD COLUMN IF NOT EXISTS source TEXT NOT NULL DEFAULT 'user',
    ADD COLUMN IF NOT EXISTS locked BOOLEAN NOT NULL DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS capability_refs JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS context_bindings JSONB NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS usage_policy JSONB NOT NULL DEFAULT '{}'::jsonb;

CREATE UNIQUE INDEX IF NOT EXISTS idx_agent_catalogue_profile_key
    ON agent_catalogue(profile_key) WHERE profile_key IS NOT NULL;

UPDATE agent_catalogue
SET capability_refs = tools
WHERE capability_refs = '[]'::jsonb AND tools <> '[]'::jsonb;

INSERT INTO agent_catalogue (
    id, profile_key, name, description, role, system_prompt, model, tools,
    capability_refs, context_bindings, usage_policy, inputs, outputs,
    verification_strategy, verification_rubric, source, locked
) VALUES
(
    'a1000000-0000-4000-8000-000000000001', 'default.researcher', 'Research Specialist',
    'Finds and summarizes current or retained information with source-aware evidence.', 'cognitive',
    'Research the assigned question using only approved sources. Separate sourced findings from inference and return concise evidence links or references.', NULL,
    '["web_search","search_memory"]'::jsonb, '["web_search","search_memory"]'::jsonb,
    '[{"kind":"public_web","access":"search"},{"kind":"approved_local_data","access":"search"}]'::jsonb,
    '{"selection":"automatic","scope":"workspace"}'::jsonb, '[]'::jsonb, '["research_summary","source_refs"]'::jsonb,
    'semantic', '["Sources are attributable","Claims distinguish evidence from inference"]'::jsonb, 'built_in', TRUE
),
(
    'a1000000-0000-4000-8000-000000000002', 'default.context-analyst', 'Context Analyst',
    'Finds relevant workspace, mounted, and long-lived organizational context.', 'cognitive',
    'Locate only context relevant to the assigned work. Respect source boundaries, identify the source used, and avoid treating retained context as autonomous authority.', NULL,
    '["search_memory","read_file","read_text_file"]'::jsonb, '["search_memory","read_file","read_text_file"]'::jsonb,
    '[{"kind":"deployment_context","access":"search"},{"kind":"approved_mount","access":"read"}]'::jsonb,
    '{"selection":"automatic","scope":"workspace"}'::jsonb, '[]'::jsonb, '["context_brief","source_refs"]'::jsonb,
    'semantic', '["Context is relevant","Source boundaries are disclosed"]'::jsonb, 'built_in', TRUE
),
(
    'a1000000-0000-4000-8000-000000000003', 'default.builder', 'Output Builder',
    'Creates retained files or packages inside the assigned team workspace.', 'actuation',
    'Build the assigned deliverable inside the team workspace. Keep working material separate from user-facing output and report retained paths plus validation evidence.', NULL,
    '["write_file","read_file","store_artifact"]'::jsonb, '["write_file","read_file","store_artifact"]'::jsonb,
    '[{"kind":"team_workspace","access":"write"},{"kind":"outcome_sources","access":"read"}]'::jsonb,
    '{"selection":"suggested","scope":"team"}'::jsonb, '[]'::jsonb, '["retained_output","proof_refs"]'::jsonb,
    'empirical', '["Output exists in the assigned workspace","Requested behavior is validated"]'::jsonb, 'built_in', TRUE
),
(
    'a1000000-0000-4000-8000-000000000004', 'default.media-creator', 'Media Creator',
    'Creates image or media outputs when an approved media capability is available.', 'actuation',
    'Create media that matches the assigned brief using approved media capabilities. Retain the output, disclose the capability used, and report blockers without inventing media.', NULL,
    '["generate_image","save_cached_image"]'::jsonb, '["generate_image","save_cached_image"]'::jsonb,
    '[{"kind":"outcome_sources","access":"read"},{"kind":"team_workspace","access":"write"}]'::jsonb,
    '{"selection":"suggested","scope":"team"}'::jsonb, '[]'::jsonb, '["media_output","proof_refs"]'::jsonb,
    'empirical', '["Media is retained","Output matches the approved brief"]'::jsonb, 'built_in', TRUE
),
(
    'a1000000-0000-4000-8000-000000000005', 'default.reviewer', 'Quality Reviewer',
    'Reviews retained work against requested acceptance and proof requirements.', 'ledger',
    'Review the assigned output independently. State what passed, what remains uncertain, and what must be repaired before the result can be trusted.', NULL,
    '["read_file","read_text_file","search_memory"]'::jsonb, '["read_file","read_text_file","search_memory"]'::jsonb,
    '[{"kind":"outcome_outputs","access":"read"},{"kind":"run_proof","access":"read"}]'::jsonb,
    '{"selection":"suggested","scope":"outcome"}'::jsonb, '[]'::jsonb, '["review_result","proof_assessment"]'::jsonb,
    'semantic', '["Acceptance criteria are addressed","Uncertainty and failed checks are explicit"]'::jsonb, 'built_in', TRUE
)
ON CONFLICT DO NOTHING;
