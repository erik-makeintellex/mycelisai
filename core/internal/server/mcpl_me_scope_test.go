package server

import (
	"encoding/json"
	"net/http"
	"reflect"
	"testing"

	"github.com/mycelis/core/internal/mcp"
)

// MCPL item 4: /user/me carries the identity's scopes and the same
// isApprover answer MCPS uses, additively, as hints only.
func TestMcplMeReportsScopesAndApprover(t *testing.T) {
	cases := map[string]struct {
		identity   *RequestIdentity
		wantScopes []any
		approver   bool
	}{
		"web admin wildcard":   {mcpsWebAdmin(), []any{"*"}, true},
		"admin with decide":    {adminWithScopes("approvals:decide", "outputs:read"), []any{"approvals:decide", "outputs:read"}, true},
		"admin without decide": {adminWithScopes("outputs:read"), []any{"outputs:read"}, false},
		"standard user":        {standardUserIdentity(), []any{"soma:work", "runs:read", "outputs:read"}, false},
		"operator with decide": {&RequestIdentity{UserID: "u-op", Role: "operator", Scopes: []string{"approvals:decide"}}, []any{"approvals:decide"}, false},
		"no scopes":            {&RequestIdentity{UserID: "u-none", Role: "operator"}, []any{}, false},
	}
	for name, c := range cases {
		t.Run(name, func(t *testing.T) {
			s := newTestServer()
			rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleMe), "GET", "/api/v1/user/me", "", c.identity)
			assertStatus(t, rr, http.StatusOK)
			var body map[string]any
			if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if got := body["scopes"]; !reflect.DeepEqual(got, c.wantScopes) {
				t.Fatalf("scopes = %#v, want %#v", got, c.wantScopes)
			}
			if got, ok := body["is_approver"].(bool); !ok || got != c.approver || got != isApprover(c.identity) {
				t.Fatalf("is_approver = %#v, want %v", body["is_approver"], c.approver)
			}
			if body["role"] != c.identity.Role {
				t.Fatalf("existing fields changed: %v", body)
			}
		})
	}
}

// MCPL item 3: the server's recorded scope is the store's own normalization.
func TestMcplStoredScopeUsesStoreNormalization(t *testing.T) {
	for _, in := range [][2]string{{"", ""}, {" ALL ", "x"}, {"Group", " g1 "}, {"host", "h"}, {"group", ""}, {"bogus", " r "}} {
		kind, ref := mcpToolSetRequest{ScopeKind: in[0], ScopeRef: in[1]}.storedScope()
		wantKind, wantRef, err := mcp.NormalizeToolSetScope(in[0], in[1])
		if err != nil {
			continue // the store refuses these; the audit keeps the trimmed input
		}
		if kind != wantKind || ref != wantRef {
			t.Fatalf("storedScope(%q,%q) = %q,%q, want %q,%q", in[0], in[1], kind, ref, wantKind, wantRef)
		}
	}
	if kind, ref := (mcpToolSetRequest{ScopeKind: " Bogus ", ScopeRef: " r "}).storedScope(); kind != "bogus" || ref != "r" {
		t.Fatalf("invalid kind recorded as %q,%q, want trimmed input", kind, ref)
	}
}
