package memory

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

// GovernedReader is the viewer a governed-memory read is filtered for
// (MEM-LIST). It is the single SQL form of the saved-memory read rule used by
// the deployment-context list and by user-facing recall; the server builds it
// from the request identity (core/internal/server/deployment_context_read.go).
//
// A viewer reads an entry when any of these hold:
//   - the entry is org-wide: knowledge_class company_knowledge or
//     soma_operating_context, or visibility other than private/team (the M2
//     isOrgWideMemory rule);
//   - the viewer saved it: owner_user_id equals UserID, or, for rows saved
//     before owner ids were recorded, loaded_by equals Label (the M2
//     isMemoryOwner rule);
//   - it is a team entry and TeamKeys holds GovernedTeamKey(tenant, team_id),
//     meaning the viewer's membership of that team was proven from persisted
//     ownership state.
//
// The zero value (no identity) reads org-wide entries only.
type GovernedReader struct {
	UserID   string
	Label    string
	TeamKeys []string
}

// governedTeamKeySep cannot occur in a trimmed tenant or team id from JSON
// metadata written by Mycelis, so a key names exactly one (tenant, team).
const governedTeamKeySep = "\x1f"

// GovernedTeamKey names one team a reader is a proven member of.
func GovernedTeamKey(tenantID, teamID string) string {
	tenantID = strings.TrimSpace(tenantID)
	if tenantID == "" {
		tenantID = "default"
	}
	return tenantID + governedTeamKeySep + strings.TrimSpace(teamID)
}

// GovernedReadClause returns the read predicate over the governed artifact
// metadata expression meta (for example "a.metadata"), appending its
// arguments from placeholder nextArg on.
func GovernedReadClause(meta string, reader GovernedReader, args []any, nextArg int) (string, []any, int) {
	field := func(key string) string { return fmt.Sprintf("NULLIF(btrim(%s->>'%s'), '')", meta, key) }
	userID := reader.UserID
	if strings.TrimSpace(userID) == "" {
		userID = ""
	}
	keys := make([]string, 0, len(reader.TeamKeys))
	for _, key := range reader.TeamKeys {
		if key != "" {
			keys = append(keys, key)
		}
	}
	clause := fmt.Sprintf(`(COALESCE(%[1]s, 'customer_context') IN ('company_knowledge', 'soma_operating_context')
		OR COALESCE(%[2]s, 'global') NOT IN ('private', 'team')
		OR (%[3]s IS NOT NULL AND $%[6]d <> '' AND %[3]s = $%[6]d)
		OR (%[3]s IS NULL AND $%[7]d <> '' AND %[4]s = $%[7]d)
		OR (%[2]s = 'team' AND %[5]s IS NOT NULL
			AND COALESCE(%[8]s, 'default') || chr(31) || %[5]s = ANY($%[9]d::text[])))`,
		field("knowledge_class"), field("visibility"), field("owner_user_id"), field("loaded_by"), field("team_id"),
		nextArg, nextArg+1, field("tenant_id"), nextArg+2)
	args = append(args, userID, strings.TrimSpace(reader.Label), textArrayLiteral(keys))
	return clause, args, nextArg + 3
}

// GovernedEntryReadable is the per-entry SQL form of GovernedReadClause for
// callers without a request (the promote tool, SRU). An unknown id, an
// artifact outside the governed store and an unreadable entry all report
// false, so the answer is no existence oracle.
func GovernedEntryReadable(ctx context.Context, db *sql.DB, artifactID string, reader GovernedReader) (bool, error) {
	read, args, _ := GovernedReadClause("a.metadata", reader, []any{strings.TrimSpace(artifactID)}, 2)
	var readable bool
	err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM artifacts a WHERE a.id::text = $1
		AND a.metadata->>'knowledge_store' = 'governed_context_store' AND `+read+`)`, args...).Scan(&readable)
	return readable, err
}

// governedRecallClause keeps recall rows the reader may read: rows outside
// the governed store are untouched; governed chunks are admitted only when
// their saved entry passes GovernedReadClause.
func governedRecallClause(reader *GovernedReader, args []any, nextArg int) (string, []any, int) {
	read, args, nextArg := GovernedReadClause("ga.metadata", *reader, args, nextArg)
	return `(COALESCE(context_vectors.metadata->>'knowledge_store', '') <> 'governed_context_store'
		OR EXISTS (SELECT 1 FROM artifacts ga WHERE ga.id::text = context_vectors.metadata->>'artifact_id'
			AND ga.metadata->>'knowledge_store' = 'governed_context_store' AND ` + read + `))`, args, nextArg
}
