package server

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

const applyTemplatePath = "/api/v1/config/templates/" + applyCustomID + "/apply"

// editedStarter is the stored custom template with a new description.
func editedStarter(t *testing.T) string {
	return strings.Replace(fixtureStarterDoc(t, applyCustomID),
		"description: Minimal starter bundle used by bootstrap and server tests.", "description: Edited by the operator.", 1)
}

func previewTemplateApply(t *testing.T, s *AdminServer, mock sqlmock.Sqlmock, doc string) (applyEnvelope, string) {
	t.Helper()
	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
	expectPreviewMint(mock, auditSourceConfigApplyPreview)
	env := postApply(t, s, applyTemplatePath, map[string]any{"document": doc, "format": "yaml"}, nil, http.StatusAccepted)
	return env, previewIntent(t, env)
}

func TestApplyTemplate_DryRunFirstWithDiffAndNoMutation(t *testing.T) {
	s, mock := newApplyServer(t)
	env, intent := previewTemplateApply(t, s, mock, editedStarter(t))
	preview := env.Data["preview"].(map[string]any)
	diff := preview["diff"].(map[string]any)
	if changed := diff["changed"].([]any); len(changed) != 1 || changed[0] != "description" {
		t.Fatalf("diff = %+v", diff)
	}
	if preview["valid"] != true || !strings.HasSuffix(intent, preview["document_sha256"].(string)) {
		t.Fatalf("preview = %+v intent=%s", preview, intent)
	}
	mockMet(t, mock) // no UPDATE operator_templates was expected or made
}

func TestApplyTemplate_ConfirmAppliesPreviewedDocument(t *testing.T) {
	s, mock := newApplyServer(t)
	doc := editedStarter(t)
	_, intent := previewTemplateApply(t, s, mock, doc)

	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
	expectConsume(mock, intent, false, time.Now().Add(time.Minute))
	expectCommitTrail(mock, "config.applied", "UPDATE operator_templates")
	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID)) // reload
	env := postApply(t, s, applyTemplatePath, map[string]any{"document": doc, "format": "yaml", "confirm_token": applyToken}, nil, http.StatusOK)
	if env.Data["applied"] != true || env.Data["intent_proof_id"] != applyProofID || env.Data["mission_event_id"] == "" {
		t.Fatalf("apply result = %+v", env.Data)
	}
	mockMet(t, mock)
}

func TestApplyTemplate_SameDocumentInJSONHasSameHash(t *testing.T) {
	yamlDoc, err := parseConfigDocument(editedStarter(t), "yaml")
	if err != nil {
		t.Fatal(err)
	}
	raw, yamlSHA, _ := canonicalConfigJSON(yamlDoc)
	jsonDoc, err := parseConfigDocument(string(raw), "json")
	if err != nil {
		t.Fatal(err)
	}
	if _, jsonSHA, _ := canonicalConfigJSON(jsonDoc); jsonSHA != yamlSHA {
		t.Fatalf("hash differs across formats: %s vs %s", yamlSHA, jsonSHA)
	}
}

func TestApplyTemplate_TokenBoundToDocumentHash(t *testing.T) {
	s, mock := newApplyServer(t)
	_, intent := previewTemplateApply(t, s, mock, editedStarter(t))

	changed := strings.Replace(editedStarter(t), "Edited by the operator.", "Changed after preview.", 1)
	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
	expectConsume(mock, intent, false, time.Now().Add(time.Minute))
	mock.ExpectRollback() // token spend is rolled back with the refused apply
	env := postApply(t, s, applyTemplatePath, map[string]any{"document": changed, "confirm_token": applyToken}, nil, http.StatusConflict)
	if !strings.Contains(env.Error, "different request") {
		t.Fatalf("error = %q", env.Error)
	}
	mockMet(t, mock)
}

func TestApplyTemplate_TokenSingleUseAndExpiry(t *testing.T) {
	cases := map[string]struct {
		consumed bool
		expires  time.Time
		want     string
	}{
		"consumed": {true, time.Now().Add(time.Minute), "already consumed"},
		"expired":  {false, time.Now().Add(-time.Minute), "expired"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			s, mock := newApplyServer(t)
			expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
			expectConsume(mock, "", tc.consumed, tc.expires)
			env := postApply(t, s, applyTemplatePath, map[string]any{"document": editedStarter(t), "confirm_token": applyToken}, nil, http.StatusBadRequest)
			if !strings.Contains(env.Error, tc.want) {
				t.Fatalf("error = %q, want %q", env.Error, tc.want)
			}
			mockMet(t, mock)
		})
	}
}

func TestApplyConfig_RejectsRedactedAndPlaceholderSecrets(t *testing.T) {
	s, mock := newApplyServer(t)
	for _, doc := range []string{
		fixtureStarterDoc(t, applyCustomID) + "secrets:\n  api_key: <redacted>\n",
		fixtureStarterDoc(t, applyCustomID) + "provider:\n  api_key: ${env:OPENAI_API_KEY}\n  api_key_env: OPENAI_API_KEY\n",
		fixtureStarterDoc(t, applyCustomID) + "provider:\n  token: ghp_literalsecretvalue0123456789\n",
	} {
		env := postApply(t, s, applyTemplatePath, map[string]any{"document": doc}, nil, http.StatusBadRequest)
		if env.Data["write_back_violations"] == nil {
			t.Fatalf("expected write_back_violations for:\n%s\ngot %+v", doc, env)
		}
	}
	// A reference name is not a secret and is accepted by the guard.
	if v := configWriteBackViolations(map[string]any{"api_key_env": "OPENAI_API_KEY", "max_tokens": 10.0}); len(v) != 0 {
		t.Fatalf("reference flagged: %v", v)
	}
	mockMet(t, mock)
}

func TestApplyTemplate_InvalidDocumentBlockedWithoutToken(t *testing.T) {
	s, mock := newApplyServer(t)
	bad := strings.Replace(fixtureStarterDoc(t, applyCustomID), "work_mode: propose_only", "work_mode: wild", 1)
	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
	mock.ExpectExec("INSERT INTO log_entries").WithArgs(auditArgs(auditSourceConfigApplyPreview)...).WillReturnResult(sqlmock.NewResult(1, 1))
	env := postApply(t, s, applyTemplatePath, map[string]any{"document": bad}, nil, http.StatusAccepted)
	preview := env.Data["preview"].(map[string]any)
	if env.Data["blocked"] != true || env.Data["confirm_token"] != nil || preview["valid"] != false {
		t.Fatalf("blocked preview = %+v", env.Data)
	}
	// A confirm attempt with an invalid document never reaches the token.
	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
	postApply(t, s, applyTemplatePath, map[string]any{"document": bad, "confirm_token": applyToken}, nil, http.StatusBadRequest)
	mockMet(t, mock)
}

func TestApplyConfig_StatusPaths(t *testing.T) {
	s, mock := newApplyServer(t)
	body := map[string]any{"document": "id: x\n"}
	postApply(t, s, "/api/v1/config/widgets/x/apply", body, nil, http.StatusBadRequest)
	postApply(t, s, "/api/v1/config/teams/admin-core/apply", body, nil, http.StatusConflict)
	postApply(t, s, "/api/v1/config/organizations/org-1/apply", body, nil, http.StatusConflict)
	postApply(t, s, applyTemplatePath, map[string]any{"document": ""}, nil, http.StatusBadRequest)
	postApply(t, s, applyTemplatePath, map[string]any{"document": "- a list", "format": "yaml"}, nil, http.StatusBadRequest)
	postApply(t, s, applyTemplatePath, map[string]any{"document": "{}", "format": "toml"}, nil, http.StatusBadRequest)

	// A file starter is read-only: 409. An unknown template: 404.
	expectOperatorTemplate(mock, applyFixtureSrc, nil)
	postApply(t, s, "/api/v1/config/templates/"+applyFixtureSrc+"/apply", body, nil, http.StatusConflict)
	expectOperatorTemplate(mock, "nobody-here", nil)
	postApply(t, s, "/api/v1/config/templates/nobody-here/apply", body, nil, http.StatusNotFound)

	// Missing migration 041: 503 naming compose.migrate.
	mock.ExpectQuery("FROM operator_templates WHERE id").WillReturnError(errors.New(`pq: relation "operator_templates" does not exist`))
	env := postApply(t, s, applyTemplatePath, body, nil, http.StatusServiceUnavailable)
	if !strings.Contains(env.Error, "compose.migrate") {
		t.Fatalf("error = %q", env.Error)
	}
	mockMet(t, mock)
}

func TestApplyConfig_Authorization(t *testing.T) {
	s, mock := newApplyServer(t)
	body := map[string]any{"document": "id: x\n"}
	rr := doRequest(t, applyMux(s), "POST", applyTemplatePath, `{"document":"id: x"}`)
	assertStatus(t, rr, http.StatusUnauthorized)
	postApply(t, s, applyTemplatePath, body, &RequestIdentity{UserID: "u", Role: "operator", Scopes: []string{"*"}}, http.StatusForbidden)
	postApply(t, s, applyTemplatePath, body, &RequestIdentity{UserID: "u", Role: "admin", Scopes: []string{"templates:read"}}, http.StatusForbidden)
	// config:write alone is not enough for groups; groups:write is also required.
	configOnly := &RequestIdentity{UserID: "u", Role: "admin", Scopes: []string{"config:write"}}
	postApply(t, s, "/api/v1/config/groups/"+applyGroupID+"/apply", body, configOnly, http.StatusForbidden)
	mockMet(t, mock)
}
