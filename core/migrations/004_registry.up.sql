-- 1. Connector Templates (The "Class")
CREATE TABLE connector_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    type TEXT NOT NULL, -- 'ingress' (Source) or 'egress' (Sink)
    image TEXT NOT NULL, -- Docker image (e.g. 'mycelis/weather-poller')
    config_schema JSONB NOT NULL, -- JSON Schema for user inputs
    topic_template TEXT NOT NULL -- e.g. "swarm.data.weather.{{city}}"
);

-- 2. Active Connectors (The "Instance")
CREATE TABLE active_connectors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    team_id UUID REFERENCES teams(id),
    template_id UUID REFERENCES connector_templates(id),
    name TEXT NOT NULL,
    config JSONB NOT NULL, -- The actual user values
    status TEXT DEFAULT 'provisioning' -- provisioning, running, error
);

-- 3. Agent Blueprints (Reusable Roles)
CREATE TABLE blueprints (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL,
    role TEXT NOT NULL,
    cognitive_profile TEXT NOT NULL, -- 'sentry', 'architect'
    tools JSONB DEFAULT '[]' -- List of MCP Tool IDs
);
