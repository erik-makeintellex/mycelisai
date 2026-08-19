DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM runtime_team_manifests LIMIT 1) THEN
        RAISE EXCEPTION 'refusing to discard persisted runtime team manifests';
    END IF;
END $$;

DROP TABLE IF EXISTS runtime_team_manifests;
