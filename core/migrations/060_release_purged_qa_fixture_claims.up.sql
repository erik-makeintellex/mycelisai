-- QA ownership rows are active cleanup leases, not permanent resource locks.
DELETE FROM qa_fixture_resources AS resource
USING qa_fixture_scopes AS scope
WHERE resource.scope_id = scope.id
  AND scope.status = 'purged';
