package memory

import (
	"fmt"
	"strings"
)

// Owner-attributed memory lanes (MEM-LANES). Remembered facts (context_vectors
// type agent_memory), conversation summaries (type conversation) and temp
// memory channels record whose they are when written:
//
//   - a write during a signed-in user's turn stores metadata owner_user_id,
//     the verified user id Core resolved for that turn; model or request
//     arguments never supply it (OwnerLaneMetadata drops them);
//   - a write with no user (background, scheduler and team-internal turns)
//     stores owner_class "system".
//
// Reads are filtered per viewer below. Rows saved before owners were recorded
// carry neither key: they are never user-readable unless they are already
// org-wide (visibility global for the vector lanes, a signal checkpoint
// channel for temp memory).
const (
	OwnerUserIDKey   = "owner_user_id"
	OwnerClassKey    = "owner_class"
	OwnerClassSystem = "system"
	// SignalCheckpointChannelPrefix names the temp channels that mirror the
	// latest bus signal on a subject. Their owner-less rows are bus state, an
	// org-wide class, not saved memory.
	SignalCheckpointChannelPrefix = "signal.latest."
)

// ownerLaneTypes are the context_vectors types written by the owner lanes.
var ownerLaneTypes = []string{"agent_memory", "conversation"}

// OwnerLaneMetadata returns a copy of metadata stamped with the owner of one
// lane write: ownerUserID, or the system class when the write has no user.
// Any caller-supplied owner key is dropped first, so it cannot be forged.
func OwnerLaneMetadata(metadata map[string]any, ownerUserID string) map[string]any {
	out := make(map[string]any, len(metadata)+1)
	for key, value := range metadata {
		if key == OwnerUserIDKey || key == OwnerClassKey {
			continue
		}
		out[key] = value
	}
	if owner := strings.TrimSpace(ownerUserID); owner != "" {
		out[OwnerUserIDKey] = owner
	} else {
		out[OwnerClassKey] = OwnerClassSystem
	}
	return out
}

// ownerLaneClause keeps owner-lane vector rows the reader may read; rows of
// other types are untouched. A row is readable when it is org-wide
// (visibility global), when the reader owns it, or when its owner shared it
// with a team the reader is a proven member of (reader.TeamKeys, the MEM-LIST
// membership proof). An owner-less team or private row is readable by no one:
// that is every legacy row and every no-user write that was not org-wide.
func ownerLaneClause(meta string, reader GovernedReader, args []any, nextArg int) (string, []any, int) {
	owner := fmt.Sprintf("NULLIF(btrim(%s->>'%s'), '')", meta, OwnerUserIDKey)
	visibility := fmt.Sprintf("COALESCE(%s->>'visibility', '')", meta)
	team := fmt.Sprintf("NULLIF(btrim(%s->>'team_id'), '')", meta)
	tenant := fmt.Sprintf("COALESCE(NULLIF(btrim(%s->>'tenant_id'), ''), 'default')", meta)
	clause := fmt.Sprintf(`(COALESCE(%[1]s->>'type', '') <> ALL($%[2]d::text[])
		OR %[3]s = 'global'
		OR (%[4]s IS NOT NULL AND $%[5]d <> '' AND %[4]s = $%[5]d)
		OR (%[4]s IS NOT NULL AND %[3]s = 'team' AND %[6]s IS NOT NULL
			AND %[7]s || chr(31) || %[6]s = ANY($%[8]d::text[])))`,
		meta, nextArg, visibility, owner, nextArg+1, team, tenant, nextArg+2)
	args = append(args, textArrayLiteral(ownerLaneTypes), strings.TrimSpace(reader.UserID), textArrayLiteral(nonEmpty(reader.TeamKeys)))
	return clause, args, nextArg + 3
}

// tempReadClause keeps temp-memory rows the reader may read. A signed-in
// reader sees their own rows plus owner-less signal checkpoints; a reader
// with no user sees system rows plus owner-less signal checkpoints. A user's
// row never reaches another user or a no-user turn, and a legacy owner-less
// row outside a signal channel reaches no one.
func tempReadClause(reader GovernedReader, args []any, nextArg int) (string, []any, int) {
	owner := fmt.Sprintf("NULLIF(btrim(metadata->>'%s'), '')", OwnerUserIDKey)
	signal := fmt.Sprintf("(%s IS NULL AND starts_with(channel_key, $%d))", owner, nextArg)
	args = append(args, SignalCheckpointChannelPrefix)
	nextArg++
	if user := strings.TrimSpace(reader.UserID); user != "" {
		args = append(args, user)
		return fmt.Sprintf("(%s = $%d OR %s)", owner, nextArg, signal), args, nextArg + 1
	}
	return fmt.Sprintf("((%s IS NULL AND metadata->>'%s' = '%s') OR %s)", owner, OwnerClassKey, OwnerClassSystem, signal), args, nextArg
}

// tempClearClause limits a clear to the reader's own rows: a signed-in
// reader's rows, or owner-less rows for a reader with no user.
func tempClearClause(reader GovernedReader, args []any, nextArg int) (string, []any, int) {
	owner := fmt.Sprintf("NULLIF(btrim(metadata->>'%s'), '')", OwnerUserIDKey)
	if user := strings.TrimSpace(reader.UserID); user != "" {
		return fmt.Sprintf("%s = $%d", owner, nextArg), append(args, user), nextArg + 1
	}
	return owner + " IS NULL", args, nextArg
}
