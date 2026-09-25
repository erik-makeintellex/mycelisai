package server

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/governance"
)

const clonePath = "/api/v1/config/templates/" + applyFixtureSrc + "/clone"

func TestCloneTemplate_PreviewIsGovernedAndDoesNotStore(t *testing.T) {
	s, mock := newApplyServer(t)
	expectOperatorTemplate(mock, applyCustomID, nil) // new id is free
	expectPreviewMint(mock, auditSourceTemplateClonePreview)

	env := postApply(t, s, clonePath, map[string]any{"name": "Acme Studio"}, nil, http.StatusAccepted)
	intent := previewIntent(t, env)
	if !strings.HasPrefix(intent, "template.clone.template."+applyCustomID+"@sha256:") {
		t.Fatalf("intent = %q", intent)
	}
	preview := env.Data["preview"].(map[string]any)
	if preview["id"] != applyCustomID || preview["source_template_id"] != applyFixtureSrc || preview["valid"] != true {
		t.Fatalf("preview = %+v", preview)
	}
	if doc, _ := preview["document"].(string); !strings.Contains(doc, "label: Acme Studio") || !strings.Contains(doc, "id: "+applyCustomID) {
		t.Fatalf("preview document:\n%s", doc)
	}
	if !strings.HasSuffix(intent, preview["document_sha256"].(string)) {
		t.Fatalf("intent %q not bound to document hash %v", intent, preview["document_sha256"])
	}
	mockMet(t, mock) // no operator_templates insert happened
}

func TestCloneTemplate_ConfirmStoresCloneAndEmitsEvents(t *testing.T) {
	s, mock := newApplyServer(t)
	expectOperatorTemplate(mock, applyCustomID, nil)
	expectPreviewMint(mock, auditSourceTemplateClonePreview)
	intent := previewIntent(t, postApply(t, s, clonePath, map[string]any{"name": "Acme Studio"}, nil, http.StatusAccepted))

	expectOperatorTemplate(mock, applyCustomID, nil)
	expectConsume(mock, intent, false, time.Now().Add(time.Minute))
	expectCommitTrail(mock, "template.cloned", "INSERT INTO operator_templates")
	env := postApply(t, s, clonePath, map[string]any{"name": "Acme Studio", "confirm_token": applyToken}, nil, http.StatusCreated)
	tpl := env.Data["template"].(map[string]any)
	if tpl["id"] != applyCustomID || tpl["source"] != "operator" || env.Data["intent_proof_id"] != applyProofID || env.Data["run_id"] == "" {
		t.Fatalf("clone result = %+v", env.Data)
	}
	mockMet(t, mock)
}

func TestCloneTemplate_ValidationFailures(t *testing.T) {
	cases := map[string]struct {
		overrides map[string]any
		want      string
	}{
		"unknown work mode": {map[string]any{"delivery_defaults": map[string]any{"work_mode": "wild"}}, "unknown work_mode"},
		"unresolved policy": {map[string]any{"delivery_defaults": map[string]any{"approval_policy_ref": "nope"}}, "not a group in policy.yaml"},
		"empty teams":       {map[string]any{"teams": []any{}}, "teams must not be empty"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, mock := newApplyServer(t)
			expectOperatorTemplate(mock, applyCustomID, nil)
			env := postApply(t, s, clonePath, map[string]any{"name": "Acme Studio", "overrides": tc.overrides}, nil, http.StatusBadRequest)
			if !strings.Contains(env.Error, tc.want) || env.Data["validation_errors"] == nil {
				t.Fatalf("error = %q data=%+v, want %q", env.Error, env.Data, tc.want)
			}
			mockMet(t, mock)
		})
	}
}

func TestCloneTemplate_OverridesMergeAndCannotChangeID(t *testing.T) {
	doc, err := buildCloneDocument(map[string]any{"id": "src", "name": "Src", "delivery_defaults": map[string]any{"work_mode": "propose_only", "audit": "standard"}},
		"custom-x", "X", map[string]any{"delivery_defaults": map[string]any{"audit": "full"}})
	if err != nil {
		t.Fatal(err)
	}
	dd := doc["delivery_defaults"].(map[string]any)
	if dd["work_mode"] != "propose_only" || dd["audit"] != "full" || doc["id"] != "custom-x" || doc["starter"].(map[string]any)["label"] != "X" {
		t.Fatalf("merged doc = %+v", doc)
	}
	if _, err := buildCloneDocument(map[string]any{"id": "src"}, "custom-x", "X", map[string]any{"id": "other"}); err == nil {
		t.Fatal("overrides changing id should fail")
	}
}

func TestCloneTemplate_RejectsWriteBackAndBadInput(t *testing.T) {
	s, mock := newApplyServer(t)
	expectOperatorTemplate(mock, applyCustomID, nil)
	env := postApply(t, s, clonePath, map[string]any{"name": "Acme Studio",
		"overrides": map[string]any{"provider_policy": map[string]any{"api_key": "<redacted>"}}}, nil, http.StatusBadRequest)
	if env.Data["write_back_violations"] == nil {
		t.Fatalf("expected write_back_violations, got %+v", env)
	}
	postApply(t, s, clonePath, map[string]any{"name": ""}, nil, http.StatusBadRequest)
	postApply(t, s, clonePath, map[string]any{"name": "X", "new_id": "Bad ID!"}, nil, http.StatusBadRequest)
	mockMet(t, mock)
}

func TestCloneTemplate_NotFoundAndConflicts(t *testing.T) {
	s, mock := newApplyServer(t)
	// Unknown source: not a file, not an operator template.
	expectOperatorTemplate(mock, "missing-template", nil)
	postApply(t, s, "/api/v1/config/templates/missing-template/clone", map[string]any{"name": "X"}, nil, http.StatusNotFound)
	// new_id collides with a file starter: 409 before any DB lookup.
	postApply(t, s, clonePath, map[string]any{"name": "X", "new_id": applyFixtureSrc}, nil, http.StatusConflict)
	// new_id collides with an existing operator template.
	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
	postApply(t, s, clonePath, map[string]any{"name": "Acme Studio"}, nil, http.StatusConflict)
	mockMet(t, mock)
}

func TestCloneTemplate_InsertRaceIsConflictAndRollsBack(t *testing.T) {
	s, mock := newApplyServer(t)
	expectOperatorTemplate(mock, applyCustomID, nil)
	expectPreviewMint(mock, auditSourceTemplateClonePreview)
	intent := previewIntent(t, postApply(t, s, clonePath, map[string]any{"name": "Acme Studio"}, nil, http.StatusAccepted))

	expectOperatorTemplate(mock, applyCustomID, nil)
	expectConsume(mock, intent, false, time.Now().Add(time.Minute))
	mock.ExpectExec("INSERT INTO log_entries").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO operator_templates").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()
	postApply(t, s, clonePath, map[string]any{"name": "Acme Studio", "confirm_token": applyToken}, nil, http.StatusConflict)
	mockMet(t, mock)
}

func TestCloneTemplate_Authorization(t *testing.T) {
	s, mock := newApplyServer(t)
	rr := doRequest(t, applyMux(s), "POST", clonePath, `{"name":"X"}`)
	assertStatus(t, rr, http.StatusUnauthorized)
	operator := &RequestIdentity{UserID: "u2", Role: "operator", Scopes: []string{"*"}}
	postApply(t, s, clonePath, map[string]any{"name": "X"}, operator, http.StatusForbidden)
	readOnlyAdmin := &RequestIdentity{UserID: "u3", Role: "admin", Scopes: []string{"groups:read", "config:read"}}
	postApply(t, s, clonePath, map[string]any{"name": "X"}, readOnlyAdmin, http.StatusForbidden)
	mockMet(t, mock)
}

func TestCloneTemplate_NoDatabaseIs503(t *testing.T) {
	s := newTestServer(withGuard(fixtureStarterPolicy(applyPolicyRef)))
	s.TemplateBundlesPath = writeStarterCatalog(t, applyPolicyRef)
	postApply(t, s, clonePath, map[string]any{"name": "X"}, nil, http.StatusServiceUnavailable)
}

// TestCloneTemplate_RealStartersPreview proves every shipped starter clones
// cleanly: no write-back violation, and the clone validates against the
// real policy.yaml.
func TestCloneTemplate_RealStartersPreview(t *testing.T) {
	engine, err := governance.NewEngine("../../config/policy.yaml")
	if err != nil {
		t.Fatalf("load real policy.yaml: %v", err)
	}
	for _, id := range []string{"starter-client-delivery-studio", "starter-product-delivery-team",
		"starter-operations-desk", "starter-governed-enterprise-program"} {
		dbOpt, mock := withDB(t)
		s := newTestServer(dbOpt, withGuard(engine.Config), withTemplateBundlesPath("../../config/templates"))
		expectOperatorTemplate(mock, "custom-copy", nil)
		expectPreviewMint(mock, auditSourceTemplateClonePreview)
		postApply(t, s, "/api/v1/config/templates/"+id+"/clone", map[string]any{"name": "Copy"}, nil, http.StatusAccepted)
		mockMet(t, mock)
	}
}
