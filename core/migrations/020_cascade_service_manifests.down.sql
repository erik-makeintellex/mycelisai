ALTER TABLE service_manifests
    DROP CONSTRAINT IF EXISTS service_manifests_team_id_fkey;
ALTER TABLE service_manifests
    ADD CONSTRAINT service_manifests_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams(id);
