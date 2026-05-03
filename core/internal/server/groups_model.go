package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	groupStatusActive   = "active"
	groupStatusPaused   = "paused"
	groupStatusArchived = "archived"
)

var validGroupWorkModes = map[string]struct{}{
	"read_only":             {},
	"propose_only":          {},
	"execute_with_approval": {},
	"execute_bounded":       {},
	"resume_continuity":     {},
}

var validGroupStatuses = map[string]struct{}{
	groupStatusActive:   {},
	groupStatusPaused:   {},
	groupStatusArchived: {},
}

// CollaborationGroup is a root-admin defined execution scope for multi-user coordination.
type CollaborationGroup struct {
	ID                  string     `json:"group_id"`
	TenantID            string     `json:"tenant_id"`
	Name                string     `json:"name"`
	GoalStatement       string     `json:"goal_statement"`
	WorkMode            string     `json:"work_mode"`
	AllowedCapabilities []string   `json:"allowed_capabilities"`
	MemberUserIDs       []string   `json:"member_user_ids"`
	TeamIDs             []string   `json:"team_ids"`
	CoordinatorProfile  string     `json:"coordinator_profile"`
	ApprovalPolicyRef   string     `json:"approval_policy_ref"`
	Status              string     `json:"status"`
	Expiry              *time.Time `json:"expiry,omitempty"`
	CreatedBy           string     `json:"created_by"`
	CreatedAuditEventID string     `json:"created_audit_event_id,omitempty"`
	UpdatedAuditEventID string     `json:"updated_audit_event_id,omitempty"`
	CreatedAt           time.Time  `json:"created_at"`
	UpdatedAt           time.Time  `json:"updated_at"`
}

type createGroupRequest struct {
	Name                string     `json:"name"`
	GoalStatement       string     `json:"goal_statement"`
	WorkMode            string     `json:"work_mode"`
	AllowedCapabilities []string   `json:"allowed_capabilities"`
	MemberUserIDs       []string   `json:"member_user_ids"`
	TeamIDs             []string   `json:"team_ids"`
	CoordinatorProfile  string     `json:"coordinator_profile"`
	ApprovalPolicyRef   string     `json:"approval_policy_ref"`
	Expiry              *time.Time `json:"expiry,omitempty"`
	ConfirmToken        string     `json:"confirm_token,omitempty"`
}

type groupBroadcastRequest struct {
	Message string `json:"message"`
}

type updateGroupStatusRequest struct {
	Status string `json:"status"`
}

func normalizeStringSlice(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func validateGroupReq(req createGroupRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if strings.TrimSpace(req.GoalStatement) == "" {
		return fmt.Errorf("goal_statement is required")
	}
	mode := strings.TrimSpace(req.WorkMode)
	if mode == "" {
		return fmt.Errorf("work_mode is required")
	}
	if _, ok := validGroupWorkModes[mode]; !ok {
		return fmt.Errorf("invalid work_mode")
	}
	return nil
}

func validateGroupStatus(status string) error {
	status = strings.TrimSpace(status)
	if status == "" {
		return fmt.Errorf("status is required")
	}
	if _, ok := validGroupStatuses[status]; !ok {
		return fmt.Errorf("invalid status")
	}
	return nil
}

func isHighImpactGroupMutation(req createGroupRequest) bool {
	mode := strings.TrimSpace(req.WorkMode)
	if mode == "execute_with_approval" || mode == "execute_bounded" {
		return true
	}
	for _, capability := range req.AllowedCapabilities {
		c := strings.ToLower(strings.TrimSpace(capability))
		if c == "" {
			continue
		}
		if strings.HasPrefix(c, "host.") ||
			strings.HasPrefix(c, "comms.") ||
			strings.HasPrefix(c, "mcp.") ||
			strings.Contains(c, "execute") ||
			strings.Contains(c, "write") {
			return true
		}
	}
	return false
}

func parseAuditUUID(id string) *uuid.UUID {
	if strings.TrimSpace(id) == "" {
		return nil
	}
	parsed, err := uuid.Parse(id)
	if err != nil {
		return nil
	}
	return &parsed
}

func (s *AdminServer) maybeRequireGroupApproval(w http.ResponseWriter, req createGroupRequest, op string, groupID string) bool {
	if !isHighImpactGroupMutation(req) {
		return false
	}

	if token := strings.TrimSpace(req.ConfirmToken); token != "" {
		if _, err := s.validateConfirmToken(token); err != nil {
			respondAPIError(w, "invalid confirm_token: "+err.Error(), http.StatusBadRequest)
			return true
		}
		return false
	}

	if s.getDB() == nil {
		respondAPIError(w, "high-impact group mutations require database-backed approval flow", http.StatusServiceUnavailable)
		return true
	}

	scope := &protocol.ScopeValidation{
		Tools:             normalizeStringSlice(req.AllowedCapabilities),
		AffectedResources: []string{"collaboration_groups", "team_channels"},
		RiskLevel:         "high",
	}
	intent := "groups." + op
	if groupID != "" {
		intent += "." + groupID
	}
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal,
		"groups."+op,
		"High-impact group mutation requested",
		map[string]any{
			"group_id":             groupID,
			"name":                 strings.TrimSpace(req.Name),
			"work_mode":            strings.TrimSpace(req.WorkMode),
			"allowed_capabilities": normalizeStringSlice(req.AllowedCapabilities),
		},
	)
	proof, err := s.createIntentProof(protocol.TemplateChatToProposal, intent, scope, auditEventID)
	if err != nil || proof == nil {
		respondAPIError(w, "failed to create approval proof", http.StatusServiceUnavailable)
		return true
	}
	confirmToken, err := s.generateConfirmToken(proof.ID, protocol.TemplateChatToProposal)
	if err != nil || confirmToken == nil {
		respondAPIError(w, "failed to create confirm token", http.StatusServiceUnavailable)
		return true
	}

	respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(map[string]any{
		"requires_approval": true,
		"group_id":          groupID,
		"intent_proof":      proof,
		"confirm_token":     confirmToken,
		"audit_event_id":    auditEventID,
	}))
	return true
}

func marshalStringList(items []string) []byte {
	items = normalizeStringSlice(items)
	raw, _ := json.Marshal(items)
	return raw
}

func unmarshalStringList(raw []byte) ([]string, error) {
	if len(raw) == 0 {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, err
	}
	return normalizeStringSlice(out), nil
}
