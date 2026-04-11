-- 038: Conversation templates for reusable Soma/Council/team asks.
-- These are distinct from CE-1 orchestration templates, connector templates,
-- organization starter bundles, and inception/RAG recipes.

CREATE TABLE IF NOT EXISTS conversation_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL DEFAULT 'default',
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    scope TEXT NOT NULL,
    created_by TEXT NOT NULL,
    creator_kind TEXT NOT NULL,
    status TEXT NOT NULL DEFAULT 'active',
    template_body TEXT NOT NULL,
    variables JSONB NOT NULL DEFAULT '{}'::jsonb,
    output_contract JSONB NOT NULL DEFAULT '{}'::jsonb,
    recommended_team_shape JSONB NOT NULL DEFAULT '{}'::jsonb,
    model_routing_hint JSONB NOT NULL DEFAULT '{}'::jsonb,
    governance_tags JSONB NOT NULL DEFAULT '[]'::jsonb,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    last_used_at TIMESTAMPTZ NULL,
    CONSTRAINT conversation_templates_scope_check
        CHECK (scope IN ('soma', 'council', 'team', 'temporary_group')),
    CONSTRAINT conversation_templates_creator_kind_check
        CHECK (creator_kind IN ('user', 'soma', 'council', 'system')),
    CONSTRAINT conversation_templates_status_check
        CHECK (status IN ('active', 'draft', 'archived'))
);

CREATE INDEX IF NOT EXISTS idx_conversation_templates_tenant_status
    ON conversation_templates(tenant_id, status, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_templates_scope
    ON conversation_templates(scope, updated_at DESC);
CREATE INDEX IF NOT EXISTS idx_conversation_templates_creator
    ON conversation_templates(creator_kind, created_by, updated_at DESC);
