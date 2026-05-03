package server

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

// GET /api/v1/groups
func (s *AdminServer) HandleListGroups(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:read"); !ok {
		return
	}
	groups, err := s.listGroupsDB(r.Context())
	if err != nil {
		respondAPIError(w, "Failed to list groups: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(groups))
}

// GET /api/v1/groups/monitor
func (s *AdminServer) HandleGroupMonitor(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:read"); !ok {
		return
	}
	monitor := s.GroupBus
	if monitor == nil {
		monitor = NewGroupBusMonitor()
		s.GroupBus = monitor
	}
	natsOnline := s.NC != nil && s.NC.IsConnected()
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(monitor.Snapshot(natsOnline)))
}

// POST /api/v1/groups
func (s *AdminServer) HandleCreateGroup(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, "groups:write")
	if !ok {
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := validateGroupReq(req); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.maybeRequireGroupApproval(w, req, "create", "") {
		return
	}

	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal,
		"groups.create",
		"Created collaboration group",
		map[string]any{
			"name":                 strings.TrimSpace(req.Name),
			"work_mode":            strings.TrimSpace(req.WorkMode),
			"approval_policy_ref":  strings.TrimSpace(req.ApprovalPolicyRef),
			"allowed_capabilities": normalizeStringSlice(req.AllowedCapabilities),
			"team_ids":             normalizeStringSlice(req.TeamIDs),
		},
	)

	group := CollaborationGroup{
		ID:                  uuid.NewString(),
		TenantID:            "default",
		Name:                strings.TrimSpace(req.Name),
		GoalStatement:       strings.TrimSpace(req.GoalStatement),
		WorkMode:            strings.TrimSpace(req.WorkMode),
		AllowedCapabilities: normalizeStringSlice(req.AllowedCapabilities),
		MemberUserIDs:       normalizeStringSlice(req.MemberUserIDs),
		TeamIDs:             normalizeStringSlice(req.TeamIDs),
		CoordinatorProfile:  strings.TrimSpace(req.CoordinatorProfile),
		ApprovalPolicyRef:   strings.TrimSpace(req.ApprovalPolicyRef),
		Status:              groupStatusActive,
		Expiry:              req.Expiry,
		CreatedBy:           identity.UserID,
		CreatedAuditEventID: auditEventID,
		UpdatedAuditEventID: auditEventID,
	}
	if err := s.insertGroupDB(r.Context(), &group); err != nil {
		respondAPIError(w, "Failed to create group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(group))
}

// PUT /api/v1/groups/{id}
func (s *AdminServer) HandleUpdateGroup(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:write"); !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	var req createGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := validateGroupReq(req); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if s.maybeRequireGroupApproval(w, req, "update", id) {
		return
	}

	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal,
		"groups.update",
		"Updated collaboration group",
		map[string]any{
			"group_id":             id,
			"name":                 strings.TrimSpace(req.Name),
			"work_mode":            strings.TrimSpace(req.WorkMode),
			"approval_policy_ref":  strings.TrimSpace(req.ApprovalPolicyRef),
			"allowed_capabilities": normalizeStringSlice(req.AllowedCapabilities),
			"team_ids":             normalizeStringSlice(req.TeamIDs),
		},
	)
	group, err := s.updateGroupDB(r.Context(), id, req, auditEventID)
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to update group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(group))
}

// PATCH /api/v1/groups/{id}/status
func (s *AdminServer) HandleUpdateGroupStatus(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:write"); !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	var req updateGroupStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := validateGroupStatus(req.Status); err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	status := strings.TrimSpace(req.Status)
	auditMessage := "Updated collaboration group status"
	if status == groupStatusArchived {
		auditMessage = "Archived collaboration group"
	}
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal,
		"groups.update_status",
		auditMessage,
		map[string]any{
			"group_id": id,
			"status":   status,
		},
	)
	group, err := s.updateGroupStatusDB(r.Context(), id, status, auditEventID)
	if errors.Is(err, sql.ErrNoRows) {
		respondAPIError(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to update group status: "+err.Error(), http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(group))
}

// GET /api/v1/groups/{id}/outputs
func (s *AdminServer) HandleGroupOutputs(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:read"); !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	group, err := s.getGroupDB(r.Context(), id)
	if err != nil {
		respondAPIError(w, "Failed to load group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		respondAPIError(w, "Group not found", http.StatusNotFound)
		return
	}

	outputs, err := s.listGroupOutputs(r.Context(), group, parseLimit(r.URL.Query().Get("limit"), 20))
	if err != nil {
		respondAPIError(w, "Failed to list group outputs: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	if outputs == nil {
		outputs = []artifacts.Artifact{}
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(outputs))
}
