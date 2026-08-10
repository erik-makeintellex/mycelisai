-- Enforce one destructive owner per fixture resource and lock scopes while
-- producers are stopped and durable state is removed.

ALTER TABLE qa_fixture_scopes
    DROP CONSTRAINT IF EXISTS chk_qa_fixture_scope_status;

ALTER TABLE qa_fixture_scopes
    ADD CONSTRAINT chk_qa_fixture_scope_status
    CHECK (status IN ('open', 'purging', 'partial', 'purged'));

-- Release historical terminal leases before adding the global ownership rule.
DELETE FROM qa_fixture_resources AS resource
USING qa_fixture_scopes AS scope
WHERE resource.scope_id = scope.id
  AND scope.status = 'purged';

-- Pre-index development databases may contain ambiguous duplicate claims.
-- Fail closed: mark every affected scope partial and remove every ambiguous
-- claim rather than assigning destructive ownership by insertion order.
WITH duplicate_refs AS (
    SELECT resource_kind, resource_ref
    FROM qa_fixture_resources
    GROUP BY resource_kind, resource_ref
    HAVING COUNT(*) > 1
), affected_scopes AS (
    SELECT DISTINCT resource.scope_id
    FROM qa_fixture_resources AS resource
    JOIN duplicate_refs AS duplicate
      ON duplicate.resource_kind = resource.resource_kind
     AND duplicate.resource_ref = resource.resource_ref
)
UPDATE qa_fixture_scopes AS scope
SET status = 'partial', updated_at = NOW()
FROM affected_scopes AS affected
WHERE scope.id = affected.scope_id;

WITH duplicate_refs AS (
    SELECT resource_kind, resource_ref
    FROM qa_fixture_resources
    GROUP BY resource_kind, resource_ref
    HAVING COUNT(*) > 1
)
DELETE FROM qa_fixture_resources AS resource
USING duplicate_refs AS duplicate
WHERE duplicate.resource_kind = resource.resource_kind
  AND duplicate.resource_ref = resource.resource_ref;

CREATE UNIQUE INDEX IF NOT EXISTS uq_qa_fixture_resource_claim
    ON qa_fixture_resources (resource_kind, resource_ref);
