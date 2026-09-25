package server

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/bootstrap"
)

const (
	configAPIVersion       = "mycelis/v1alpha1"
	configKindGroupDocName = "CollaborationGroup"
)

// groupSpecDocument is the spec block apply accepts for a group: exactly
// the fields View config shows. Unknown fields are rejected, not dropped.
type groupSpecDocument struct {
	Name                string     `json:"name"`
	GoalStatement       string     `json:"goal_statement"`
	WorkMode            string     `json:"work_mode"`
	AllowedCapabilities []string   `json:"allowed_capabilities"`
	MemberUserIDs       []string   `json:"member_user_ids"`
	TeamIDs             []string   `json:"team_ids"`
	CoordinatorProfile  string     `json:"coordinator_profile"`
	ApprovalPolicyRef   string     `json:"approval_policy_ref"`
	Expiry              *time.Time `json:"expiry,omitempty"`
}

// groupApplyTarget loads a group and validates a CollaborationGroup
// document for it. Only spec is applied; metadata is informational except
// metadata.id, which must match the path id.
func (s *AdminServer) groupApplyTarget(ctx context.Context, id string, doc map[string]any) (*configApplyTarget, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, configNotFound("group", id)
	}
	if s.getDB() == nil {
		return nil, configUnavailable("database not configured; group config is unavailable")
	}
	group, err := s.getGroupDB(ctx, id)
	if err != nil {
		log.Printf("[config-apply] group query: %v", err)
		return nil, configUnavailable("group query failed")
	}
	if group == nil {
		return nil, configNotFound("group", id)
	}
	current, err := normalizeConfigObject(groupConfigSpec(group))
	if err != nil {
		return nil, err
	}

	var problems []string
	if kind := fmt.Sprint(doc["kind"]); kind != configKindGroupDocName {
		problems = append(problems, "kind must be "+configKindGroupDocName)
	}
	if v, ok := doc["apiVersion"]; ok && fmt.Sprint(v) != configAPIVersion {
		problems = append(problems, "apiVersion must be "+configAPIVersion)
	}
	if meta, ok := doc["metadata"].(map[string]any); ok {
		if mid, ok := meta["id"]; ok && fmt.Sprint(mid) != id {
			problems = append(problems, "metadata.id must be "+id)
		}
	}
	spec, _ := doc["spec"].(map[string]any)
	req, err := decodeGroupSpec(spec)
	if err != nil {
		problems = append(problems, err.Error())
	} else if err := validateGroupReq(req); err != nil {
		problems = append(problems, "spec: "+err.Error())
	}

	return &configApplyTarget{
		current:    current,
		diffAfter:  spec,
		proposed:   doc,
		validation: problems,
		auditDetails: map[string]any{
			"name":                 req.Name,
			"work_mode":            req.WorkMode,
			"approval_policy_ref":  req.ApprovalPolicyRef,
			"allowed_capabilities": normalizeStringSlice(req.AllowedCapabilities),
			"team_ids":             normalizeStringSlice(req.TeamIDs),
		},
		apply: func(ctx context.Context, tx *sql.Tx, auditID, _ string) error {
			return execUpdateGroup(ctx, tx, id, req, auditID)
		},
		result: func(ctx context.Context) (any, error) { return s.getGroupDB(ctx, id) },
	}, nil
}

func decodeGroupSpec(spec map[string]any) (createGroupRequest, error) {
	if spec == nil {
		return createGroupRequest{}, fmt.Errorf("spec is required")
	}
	if expiry, ok := spec["expiry"].(string); ok && expiry == "" {
		delete(spec, "expiry")
	}
	raw, _ := json.Marshal(spec)
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	var in groupSpecDocument
	if err := dec.Decode(&in); err != nil {
		return createGroupRequest{}, fmt.Errorf("spec: %v", err)
	}
	return createGroupRequest{
		Name: in.Name, GoalStatement: in.GoalStatement, WorkMode: in.WorkMode,
		AllowedCapabilities: in.AllowedCapabilities, MemberUserIDs: in.MemberUserIDs, TeamIDs: in.TeamIDs,
		CoordinatorProfile: in.CoordinatorProfile, ApprovalPolicyRef: in.ApprovalPolicyRef, Expiry: in.Expiry,
	}, nil
}

// templateApplyTarget loads an operator template and validates a
// replacement bundle with the starter rules. File templates are read-only.
func (s *AdminServer) templateApplyTarget(ctx context.Context, id string, doc map[string]any) (*configApplyTarget, error) {
	if s.getDB() == nil {
		return nil, configUnavailable("database not configured; operator templates are unavailable")
	}
	var row *operatorTemplateRow
	if operatorTemplateIDPattern.MatchString(id) {
		var err error
		if row, err = s.getOperatorTemplate(ctx, id); err != nil {
			return nil, operatorTemplateStoreError(err)
		}
	}
	if row == nil {
		fileDoc, ferr := s.fileTemplateDocument(id)
		if ferr != nil {
			return nil, ferr
		}
		if fileDoc != nil {
			return nil, configConflict("template " + id + " is a read-only file template; clone it to create an editable Custom template")
		}
		return nil, configNotFound("template", id)
	}
	var current map[string]any
	if err := json.Unmarshal(row.Document, &current); err != nil {
		current = map[string]any{}
	}
	if _, ok := doc["id"]; !ok {
		doc["id"] = id
	}
	var problems []string
	if docID := fmt.Sprint(doc["id"]); docID != id {
		problems = append(problems, "id must be "+id)
	}
	canonical, sha, err := canonicalConfigJSON(doc)
	if err != nil {
		return nil, err
	}
	name := row.Name
	if bundle, verr := bootstrap.ValidateStarterDocument(canonical, s.starterPolicyGroups()); verr != nil {
		problems = append(problems, verr.Error())
	} else {
		name = bundle.Name
	}
	return &configApplyTarget{
		current:      current,
		proposed:     doc,
		validation:   problems,
		auditDetails: map[string]any{"name": name, "source_template_id": row.SourceTemplateID},
		apply: func(ctx context.Context, tx *sql.Tx, auditID, proofID string) error {
			return updateOperatorTemplateTx(ctx, tx, operatorTemplateWrite{
				ID: id, Name: name, Document: canonical, DocumentSHA256: sha,
				AuditEventID: auditID, IntentProofID: proofID,
			})
		},
		result: func(ctx context.Context) (any, error) {
			updated, err := s.getOperatorTemplate(ctx, id)
			if err != nil || updated == nil {
				return nil, err
			}
			return operatorTemplateSummary(updated), nil
		},
	}, nil
}

// operatorTemplateSummary is the API view of an operator template row.
func operatorTemplateSummary(row *operatorTemplateRow) map[string]any {
	return map[string]any{
		"id":                 row.ID,
		"name":               row.Name,
		"source":             starterSourceOperator,
		"source_template_id": row.SourceTemplateID,
		"document_sha256":    row.DocumentSHA256,
		"created_by":         row.CreatedBy,
		"created_at":         configTime(row.CreatedAt),
		"updated_at":         configTime(row.UpdatedAt),
	}
}
