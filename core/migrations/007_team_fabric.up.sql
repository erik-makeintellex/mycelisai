-- Migration 007: The Fractal Fabric (Data Seeds & Node Assignment)
-- 1. Link Nodes to Teams (Ownership)
ALTER TABLE nodes
ADD COLUMN team_id UUID REFERENCES teams(id);
-- 2. Seed Root Supervisor (The Architect)
INSERT INTO users (id, username, role, settings)
VALUES (
        '00000000-0000-0000-0000-000000000000',
        -- Fixed ID for Root
        'architect',
        'admin',
        '{"theme": "cyber-minimal"}'
    ) ON CONFLICT (id) DO NOTHING;
-- 3. Seed Root Mission (The Prime Directive)
INSERT INTO missions (id, owner_id, name, directive)
VALUES (
        '11111111-1111-1111-1111-111111111111',
        '00000000-0000-0000-0000-000000000000',
        'Symbiotic Swarm',
        'Establish a self-governing, fractal intelligence network that serves humanity.'
    ) ON CONFLICT (id) DO NOTHING;
-- 4. Seed Root Team (Mycelis Core)
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
-- 5. Create "Directives" Table (if separate entity needed)
-- For now we stick to Missions having a Directive.
-- But if we want sub-directives per team that are distinct from Mission...
-- Let's add a `directives` JSONB column to teams? 
-- Or rely on Mission inheritance.
-- Start simple: Teams inherit Mission.