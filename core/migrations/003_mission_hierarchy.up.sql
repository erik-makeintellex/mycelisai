-- 1. Create the Missions Table (The "Constitution")
CREATE TABLE missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    owner_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    directive TEXT NOT NULL, -- The "Prime Directive" (e.g., "Symbiotic Swarm")
    created_at TIMESTAMPTZ DEFAULT NOW(),
    updated_at TIMESTAMPTZ DEFAULT NOW()
);

-- 2. Update Teams for Hierarchy
ALTER TABLE teams 
    ADD COLUMN mission_id UUID REFERENCES missions(id) ON DELETE CASCADE,
    ADD COLUMN parent_id UUID REFERENCES teams(id) ON DELETE SET NULL,
    ADD COLUMN path TEXT DEFAULT ''; -- Materialized Path (e.g. "root_id.sub_id")

-- 3. Indexes for Tree Traversal Performance
CREATE INDEX idx_teams_mission_id ON teams(mission_id);
CREATE INDEX idx_teams_parent_id ON teams(parent_id);
CREATE INDEX idx_teams_path ON teams(path); -- Enables fast "Get all descendants" queries
