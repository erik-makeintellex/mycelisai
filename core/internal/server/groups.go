package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"sync"
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

type GroupBusSnapshot struct {
	Status          string   `json:"status"`
	PublishedCount  int64    `json:"published_count"`
	LastGroupID     string   `json:"last_group_id,omitempty"`
	LastActorID     string   `json:"last_actor_id,omitempty"`
	LastMessage     string   `json:"last_message,omitempty"`
	LastSubjects    []string `json:"last_subjects,omitempty"`
	LastPublishedAt string   `json:"last_published_at,omitempty"`
	LastError       string   `json:"last_error,omitempty"`
}

// GroupBusMonitor tracks recent group-bus fanout activity for status surfaces.
type GroupBusMonitor struct {
	mu             sync.RWMutex
	publishedCount int64
	lastGroupID    string
	lastActorID    string
	lastMessage    string
	lastSubjects   []string
	lastPublished  time.Time
	lastError      string
}

func NewGroupBusMonitor() *GroupBusMonitor {
	return &GroupBusMonitor{}
}

func (m *GroupBusMonitor) RecordSuccess(groupID, actorID, message string, subjects []string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.publishedCount++
	m.lastGroupID = groupID
	m.lastActorID = actorID
	m.lastMessage = message
	m.lastSubjects = append([]string(nil), subjects...)
	m.lastPublished = time.Now().UTC()
	m.lastError = ""
}

func (m *GroupBusMonitor) RecordFailure(err error) {
	if err == nil {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastError = err.Error()
}

func (m *GroupBusMonitor) Snapshot(natsOnline bool) GroupBusSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	s := GroupBusSnapshot{
		Status:         "offline",
		PublishedCount: m.publishedCount,
		LastGroupID:    m.lastGroupID,
		LastActorID:    m.lastActorID,
		LastMessage:    m.lastMessage,
		LastSubjects:   append([]string(nil), m.lastSubjects...),
		LastError:      m.lastError,
	}
	if !m.lastPublished.IsZero() {
		s.LastPublishedAt = m.lastPublished.Format(time.RFC3339Nano)
	}
	if natsOnline {
		s.Status = "online"
	}
	if natsOnline && m.lastError != "" {
		s.Status = "degraded"
	}
	return s
}

func hasScope(identity *RequestIdentity, required string) bool {
	if identity == nil || required == "" {
		return false
	}
	for _, scope := range identity.Scopes {
		scope = strings.TrimSpace(scope)
		if scope == "*" || scope == required {
			return true
		}
		if strings.HasSuffix(scope, ":*") {
			prefix := strings.TrimSuffix(scope, "*")
			if strings.HasPrefix(required, prefix) {
				return true
			}
		}
	}
	return false
}

func requireRootAdminScope(w http.ResponseWriter, r *http.Request, requiredScope string) (*RequestIdentity, bool) {
	identity := IdentityFromContext(r.Context())
	if identity == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return nil, false
	}
	if identity.Role != "admin" {
		respondAPIError(w, "Root admin role required", http.StatusForbidden)
		return nil, false
	}
	if !hasScope(identity, requiredScope) {
		respondAPIError(w, "Missing required scope: "+requiredScope, http.StatusForbidden)
		return nil, false
	}
	return identity, true
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

func scanGroupRow(scanner interface{ Scan(dest ...any) error }) (*CollaborationGroup, error) {
	var (
		g              CollaborationGroup
		allowedRaw     []byte
		memberRaw      []byte
		teamRaw        []byte
		expiry         sql.NullTime
		createdAuditID string
		updatedAuditID string
	)
	if err := scanner.Scan(
		&g.ID,
		&g.TenantID,
		&g.Name,
		&g.GoalStatement,
		&g.WorkMode,
		&allowedRaw,
		&memberRaw,
		&teamRaw,
		&g.CoordinatorProfile,
		&g.ApprovalPolicyRef,
		&g.Status,
		&g.CreatedBy,
		&expiry,
		&createdAuditID,
		&updatedAuditID,
		&g.CreatedAt,
		&g.UpdatedAt,
	); err != nil {
		return nil, err
	}
	var err error
	g.AllowedCapabilities, err = unmarshalStringList(allowedRaw)
	if err != nil {
		return nil, fmt.Errorf("decode allowed_capabilities: %w", err)
	}
	g.MemberUserIDs, err = unmarshalStringList(memberRaw)
	if err != nil {
		return nil, fmt.Errorf("decode member_user_ids: %w", err)
	}
	g.TeamIDs, err = unmarshalStringList(teamRaw)
	if err != nil {
		return nil, fmt.Errorf("decode team_ids: %w", err)
	}
	if expiry.Valid {
		ts := expiry.Time.UTC()
		g.Expiry = &ts
	}
	g.CreatedAuditEventID = strings.TrimSpace(createdAuditID)
	g.UpdatedAuditEventID = strings.TrimSpace(updatedAuditID)
	return &g, nil
}

func (s *AdminServer) listGroupsDB(ctx context.Context) ([]CollaborationGroup, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, tenant_id, name, goal_statement, work_mode,
		       allowed_capabilities, member_user_ids, team_ids,
		       coordinator_profile, approval_policy_ref, status, created_by,
		       expiry,
		       COALESCE(created_audit_event_id::text, ''),
		       COALESCE(updated_audit_event_id::text, ''),
		       created_at, updated_at
		FROM collaboration_groups
		WHERE tenant_id='default'
		ORDER BY created_at DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]CollaborationGroup, 0)
	for rows.Next() {
		g, scanErr := scanGroupRow(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		out = append(out, *g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func (s *AdminServer) getGroupDB(ctx context.Context, id string) (*CollaborationGroup, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	row := db.QueryRowContext(ctx, `
		SELECT id::text, tenant_id, name, goal_statement, work_mode,
		       allowed_capabilities, member_user_ids, team_ids,
		       coordinator_profile, approval_policy_ref, status, created_by,
		       expiry,
		       COALESCE(created_audit_event_id::text, ''),
		       COALESCE(updated_audit_event_id::text, ''),
		       created_at, updated_at
		FROM collaboration_groups
		WHERE id=$1 AND tenant_id='default'`, id)

	g, err := scanGroupRow(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return g, err
}

func (s *AdminServer) insertGroupDB(ctx context.Context, g *CollaborationGroup) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return db.QueryRowContext(ctx, `
		INSERT INTO collaboration_groups (
			id, tenant_id, name, goal_statement, work_mode,
			allowed_capabilities, member_user_ids, team_ids,
			coordinator_profile, approval_policy_ref, status,
			created_by, expiry, created_audit_event_id, updated_audit_event_id
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8,
			$9, $10, $11,
			$12, $13, $14, $15
		)
		RETURNING created_at, updated_at`,
		g.ID, g.TenantID, g.Name, g.GoalStatement, g.WorkMode,
		marshalStringList(g.AllowedCapabilities),
		marshalStringList(g.MemberUserIDs),
		marshalStringList(g.TeamIDs),
		g.CoordinatorProfile, g.ApprovalPolicyRef, g.Status,
		g.CreatedBy, g.Expiry,
		parseAuditUUID(g.CreatedAuditEventID),
		parseAuditUUID(g.UpdatedAuditEventID),
	).Scan(&g.CreatedAt, &g.UpdatedAt)
}

func (s *AdminServer) updateGroupDB(ctx context.Context, id string, req createGroupRequest, updatedAuditEventID string) (*CollaborationGroup, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	res, err := db.ExecContext(ctx, `
		UPDATE collaboration_groups
		SET name=$2,
		    goal_statement=$3,
		    work_mode=$4,
		    allowed_capabilities=$5,
		    member_user_ids=$6,
		    team_ids=$7,
		    coordinator_profile=$8,
		    approval_policy_ref=$9,
		    expiry=$10,
		    updated_audit_event_id=$11,
		    updated_at=NOW()
		WHERE id=$1 AND tenant_id='default'`,
		id,
		strings.TrimSpace(req.Name),
		strings.TrimSpace(req.GoalStatement),
		strings.TrimSpace(req.WorkMode),
		marshalStringList(req.AllowedCapabilities),
		marshalStringList(req.MemberUserIDs),
		marshalStringList(req.TeamIDs),
		strings.TrimSpace(req.CoordinatorProfile),
		strings.TrimSpace(req.ApprovalPolicyRef),
		req.Expiry,
		parseAuditUUID(updatedAuditEventID),
	)
	if err != nil {
		return nil, err
	}
	rowsAffected, _ := res.RowsAffected()
	if rowsAffected == 0 {
		return nil, sql.ErrNoRows
	}
	return s.getGroupDB(ctx, id)
}

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

// POST /api/v1/groups/{id}/broadcast
// Publishes in parallel to the group collab channel and each team's internal command lane.
func (s *AdminServer) HandleGroupBroadcast(w http.ResponseWriter, r *http.Request) {
	identity, ok := requireRootAdminScope(w, r, "groups:broadcast")
	if !ok {
		return
	}
	if s.NC == nil || !s.NC.IsConnected() {
		respondAPIError(w, "NATS connection offline", http.StatusServiceUnavailable)
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
	if group.Status != groupStatusActive {
		respondAPIError(w, "Group is not active", http.StatusConflict)
		return
	}

	var req groupBroadcastRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	msg := strings.TrimSpace(req.Message)
	if msg == "" {
		respondAPIError(w, "message is required", http.StatusBadRequest)
		return
	}

	subjects := []string{fmt.Sprintf(protocol.TopicGroupCollabFmt, group.ID)}
	for _, teamID := range group.TeamIDs {
		teamID = strings.TrimSpace(teamID)
		if teamID == "" {
			continue
		}
		subjects = append(subjects, fmt.Sprintf(protocol.TopicTeamInternalCommand, teamID))
	}

	payload := map[string]any{
		"group_id":   group.ID,
		"tenant_id":  group.TenantID,
		"message":    msg,
		"actor_id":   identity.UserID,
		"emitted_at": time.Now().UTC().Format(time.RFC3339Nano),
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		respondAPIError(w, "Failed to serialize payload", http.StatusInternalServerError)
		return
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(subjects))
	for _, subject := range subjects {
		subject := subject
		wg.Add(1)
		go func() {
			defer wg.Done()
			if pubErr := s.NC.Publish(subject, raw); pubErr != nil {
				errCh <- fmt.Errorf("%s: %w", subject, pubErr)
			}
		}()
	}
	wg.Wait()
	close(errCh)

	for pubErr := range errCh {
		if s.GroupBus != nil {
			s.GroupBus.RecordFailure(pubErr)
		}
		respondAPIError(w, "Failed to publish group broadcast: "+pubErr.Error(), http.StatusBadGateway)
		return
	}

	if s.GroupBus != nil {
		s.GroupBus.RecordSuccess(group.ID, identity.UserID, msg, subjects)
	}

	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal,
		"groups.broadcast",
		"Broadcast to collaboration group",
		map[string]any{
			"group_id":   group.ID,
			"subjects":   subjects,
			"message":    msg,
			"actor_id":   identity.UserID,
			"team_ids":   group.TeamIDs,
			"team_count": len(group.TeamIDs),
		},
	)

	respondAPIJSON(w, http.StatusAccepted, protocol.NewAPISuccess(map[string]any{
		"group_id":       group.ID,
		"status":         "queued",
		"subjects":       subjects,
		"team_count":     len(group.TeamIDs),
		"audit_event_id": auditEventID,
	}))
}
