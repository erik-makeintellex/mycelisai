-- Migration 002: V6 Core Schema
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    role TEXT DEFAULT 'operator',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES users(id),
    name TEXT NOT NULL,
    role TEXT DEFAULT 'admin',
    settings JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS service_manifests (
    id UUID PRIMARY KEY,
    team_id UUID REFERENCES teams(id),
    name TEXT NOT NULL,
    manifest JSONB NOT NULL,
    status TEXT DEFAULT 'draft',
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
-- Migration 003: Mission Hierarchy
CREATE TABLE IF NOT EXISTS missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    directive TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
ALTER TABLE teams
ADD COLUMN IF NOT EXISTS mission_id UUID REFERENCES missions(id) ON DELETE CASCADE;
ALTER TABLE teams
ADD COLUMN IF NOT EXISTS parent_id UUID REFERENCES teams(id) ON DELETE
SET NULL;
ALTER TABLE teams
ADD COLUMN IF NOT EXISTS path TEXT DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_teams_mission_id ON teams(mission_id);
CREATE INDEX IF NOT EXISTS idx_teams_parent_id ON teams(parent_id);
CREATE INDEX IF NOT EXISTS idx_teams_path ON teams(path);
-- Migration 004: Registry
CREATE TABLE IF NOT EXISTS connector_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    image TEXT NOT NULL,
    config_schema JSONB NOT NULL,
    topic_template TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS active_connectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID REFERENCES teams(id),
    template_id UUID REFERENCES connector_templates(id),
    name TEXT NOT NULL,
    config JSONB NOT NULL,
    status TEXT DEFAULT 'provisioning'
);
CREATE TABLE IF NOT EXISTS blueprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    cognitive_profile TEXT NOT NULL,
    tools JSONB DEFAULT '[]'
);
-- Migration 005: Nodes (If Not Exists)
CREATE TABLE IF NOT EXISTS nodes (
    id TEXT PRIMARY KEY,
    status TEXT DEFAULT 'pending',
    capabilities JSONB DEFAULT '[]',
    last_seen TIMESTAMPTZ DEFAULT NOW(),
    created_at TIMESTAMPTZ DEFAULT NOW()
);
-- Migration 006: Cognitive Registry
CREATE TABLE IF NOT EXISTS llm_providers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    driver TEXT NOT NULL,
    base_url TEXT NOT NULL,
    api_key_env_var TEXT,
    config JSONB DEFAULT '{}',
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE TABLE IF NOT EXISTS system_config (
    key TEXT PRIMARY KEY,
    value TEXT NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW()
);
INSERT INTO llm_providers (id, name, driver, base_url, is_default)
VALUES (
        'local-sovereign',
        'Local Sovereign (Ollama)',
        'ollama',
        'http://ollama-service:11434',
        TRUE
    ) ON CONFLICT (id) DO NOTHING;
INSERT INTO system_config (key, value)
VALUES ('role.architect', 'local-sovereign'),
    ('role.coder', 'local-sovereign'),
    ('role.sentry', 'local-sovereign') ON CONFLICT (key) DO NOTHING;
-- Migration 007: Team Fabric
ALTER TABLE nodes
ADD COLUMN IF NOT EXISTS team_id UUID REFERENCES teams(id);
INSERT INTO users (id, username, role, settings)
VALUES (
        '00000000-0000-0000-0000-000000000000',
        'architect',
        'admin',
        '{"theme": "cyber-minimal"}'
    ) ON CONFLICT (id) DO NOTHING;
INSERT INTO missions (id, owner_id, name, directive)
VALUES (
        '11111111-1111-1111-1111-111111111111',
        '00000000-0000-0000-0000-000000000000',
        'Symbiotic Swarm',
        'Establish a self-governing, fractal intelligence network that serves humanity.'
    ) ON CONFLICT (id) DO NOTHING;
INSERT INTO teams (
        id,
        owner_id,
        name,
        role,
        mission_id,
        parent_id,
        path
    )
VALUES (
        '22222222-2222-2222-2222-222222222222',
        '00000000-0000-0000-0000-000000000000',
        'Mycelis Core',
        'admin',
        '11111111-1111-1111-1111-111111111111',
        NULL,
        '22222222-2222-2222-2222-222222222222'
    ) ON CONFLICT (id) DO NOTHING;
-- Migration 008: Context Engine
CREATE EXTENSION IF NOT EXISTS vector;
CREATE TABLE IF NOT EXISTS context_vectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    entity_id TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    content TEXT NOT NULL,
    embedding vector(768),
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_context_entity ON context_vectors(entity_id);
CREATE INDEX IF NOT EXISTS idx_context_embedding ON context_vectors USING hnsw (embedding vector_cosine_ops);