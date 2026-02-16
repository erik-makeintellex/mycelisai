-- Fix service_manifests → teams FK to cascade on delete.
-- Without this, DELETE FROM missions cascades to teams but fails
-- on the service_manifests FK constraint (orphan rows).
ALTER TABLE service_manifests
    DROP CONSTRAINT IF EXISTS service_manifests_team_id_fkey;
ALTER TABLE service_manifests
    ADD CONSTRAINT service_manifests_team_id_fkey
    FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
