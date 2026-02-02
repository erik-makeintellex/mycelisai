-- 1. Identity
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    role TEXT NOT NULL DEFAULT 'operator', -- 'admin', 'operator', 'observer'
    created_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS teams (
    id UUID PRIMARY KEY,
    name TEXT NOT NULL,
    owner_id UUID REFERENCES users(id)
);

CREATE TABLE IF NOT EXISTS team_members (
    team_id UUID REFERENCES teams(id),
    user_id UUID REFERENCES users(id),
    role TEXT NOT NULL, -- 'lead', 'member'
    PRIMARY KEY (team_id, user_id)
);

-- 2. Configuration (Stable Persistence)
CREATE TABLE IF NOT EXISTS user_settings (
    user_id UUID REFERENCES users(id) PRIMARY KEY,
    preferences JSONB NOT NULL DEFAULT '{}' 
    -- e.g. {"theme": "aero-light", "notifications": true}
);

CREATE TABLE IF NOT EXISTS team_settings (
    team_id UUID REFERENCES teams(id) PRIMARY KEY,
    config JSONB NOT NULL DEFAULT '{}'
    -- e.g. {"model_matrix": {...}, "api_limits": {...}}
);
