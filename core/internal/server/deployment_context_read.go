package server

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/internal/memory"
)

// Saved-memory read authority (MEM-LIST). Reading is wider than managing
// (memoryLifecycleDenial) and built from the same M2 predicates:
//
//   - org-wide entries (isOrgWideMemory) are readable by every caller;
//   - the saver (isMemoryOwner) reads their own private and team entries;
//   - a team entry is readable by a proven member of its team, using the B1R
//     persisted membership proof (tokenBudgetTeamTenant: active ownership
//     binding plus an active, unexpired org membership in the owner group).
//
// Root admins get no extra read: they manage org-wide entries only, so they
// read what any member reads. No identity reads org-wide entries only. The
// list and user-facing search apply the same rule in SQL through
// memory.GovernedReadClause; entry routes apply canReadMemoryEntry and answer
// an unreadable entry exactly like an unknown id.

// memoryTeamMember reports whether the caller is a proven member of teamID in
// tenantID. No identity, a non-UUID principal, or no database proves nothing.
func (s *AdminServer) memoryTeamMember(r *http.Request, tenantID, teamID string) (bool, error) {
	if IdentityFromContext(r.Context()) == nil || strings.TrimSpace(teamID) == "" {
		return false, nil
	}
	tenant, err := s.tokenBudgetTeamTenant(r, teamID)
	if err != nil {
		return false, err
	}
	return tenant != "" && tenant == tenantID, nil
}

// canReadMemoryEntry is the per-entry form of memory.GovernedReadClause.
func (s *AdminServer) canReadMemoryEntry(r *http.Request, identity *RequestIdentity, record *deploymentcontext.EntryRecord) (bool, error) {
	if isOrgWideMemory(record) || isMemoryOwner(identity, record) {
		return true, nil
	}
	if record.Visibility != "team" || record.TeamID == "" {
		return false, nil
	}
	return s.memoryTeamMember(r, record.TenantID, record.TeamID)
}

// memoryReader builds the SQL read filter for the caller. Team keys are
// proven one team at a time; a failed membership check is an error, never a
// silently shorter list. Without a deployment-context store there are no
// team entries to prove, so the reader carries none.
func (s *AdminServer) memoryReader(r *http.Request) (memory.GovernedReader, error) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		return memory.GovernedReader{}, nil
	}
	reader := memory.GovernedReader{Label: memoryOwnerLabel(identity)}
	if strings.TrimSpace(identity.UserID) != "" {
		reader.UserID = identity.UserID
	}
	svc := s.deploymentContextService()
	if svc.Artifacts == nil || svc.Artifacts.DB == nil {
		return reader, nil
	}
	refs, err := svc.TeamRefs(r.Context())
	if err != nil {
		return reader, err
	}
	for _, ref := range refs {
		member, err := s.memoryTeamMember(r, ref.TenantID, ref.TeamID)
		if err != nil {
			return reader, fmt.Errorf("check team membership: %w", err)
		}
		if member {
			reader.TeamKeys = append(reader.TeamKeys, memory.GovernedTeamKey(ref.TenantID, ref.TeamID))
		}
	}
	return reader, nil
}
