package server

import (
	"errors"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStarterCatalog_ListsOperatorTemplatesAsCustom(t *testing.T) {
	s, mock := newApplyServer(t)
	now := time.Now()
	rows := operatorTemplateRows().
		AddRow(applyCustomID, "Acme Studio", applyFixtureSrc, fixtureStarterJSON(t, applyCustomID), "sha", "u1", now, now).
		AddRow(applyFixtureSrc, "Shadow", applyFixtureSrc, fixtureStarterJSON(t, applyFixtureSrc), "sha", "u1", now, now)
	mock.ExpectQuery("FROM operator_templates WHERE tenant_id").WillReturnRows(rows)

	items := listStarterTemplates(t, s)
	var custom, shadow map[string]any
	for _, item := range items {
		if item["source"] != "operator" {
			continue
		}
		switch item["id"] {
		case applyCustomID:
			custom = item
		case applyFixtureSrc:
			shadow = item
		}
	}
	if custom == nil || custom["available"] != true || custom["source_template_id"] != applyFixtureSrc || custom["delivery_summary"] == nil {
		t.Fatalf("custom entry = %+v", custom)
	}
	if shadow == nil || shadow["available"] != false || !strings.Contains(shadow["unavailable_reason"].(string), "already used by a template file") {
		t.Fatalf("shadowing entry = %+v", shadow)
	}
	mockMet(t, mock)
}

func TestStarterCatalog_OperatorStorageFailureKeepsFileStarters(t *testing.T) {
	s, mock := newApplyServer(t)
	mock.ExpectQuery("FROM operator_templates WHERE tenant_id").WillReturnError(errors.New(`pq: relation "operator_templates" does not exist`))
	items := listStarterTemplates(t, s)
	found := false
	for _, item := range items {
		if item["id"] == applyFixtureSrc && item["available"] == true {
			found = true
		}
	}
	if !found {
		t.Fatalf("file starter missing when operator storage fails: %+v", items)
	}
	mockMet(t, mock)
}

func TestConfigView_OperatorTemplateSource(t *testing.T) {
	s, mock := newApplyServer(t)
	expectOperatorTemplate(mock, applyCustomID, fixtureStarterJSON(t, applyCustomID))
	rr := doAuthenticatedRequest(t, applyMux(s), "GET", "/api/v1/config/templates/"+applyCustomID, "")
	assertStatus(t, rr, http.StatusOK)
	body := rr.Body.String()
	for _, want := range []string{`"source":"operator"`, `operator_templates/` + applyCustomID, "fixture-delivery"} {
		if !strings.Contains(body, want) {
			t.Fatalf("view missing %q: %s", want, body)
		}
	}
	mockMet(t, mock)
}

func TestConfigDiff_PathsOnly(t *testing.T) {
	before := map[string]any{"a": 1.0, "b": map[string]any{"c": "x", "gone": true}, "list": []any{"p", "q"}}
	after := map[string]any{"a": 2.0, "b": map[string]any{"c": "x", "new": "secret-value"}, "list": []any{"p"}}
	diff := diffConfigDocuments(before, after)
	if strings.Join(diff.Changed, ",") != "a" || strings.Join(diff.Added, ",") != "b.new" ||
		strings.Join(diff.Removed, ",") != "b.gone,list[1]" {
		t.Fatalf("diff = %+v", diff)
	}
}
