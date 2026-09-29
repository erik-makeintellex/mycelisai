package server

import (
	"context"
	"testing"
)

// MEM-LANES-2 C1: only the M2 authority (root admin with memory:write) lets a
// memory tool save org-wide; Core derives it from the request identity.
func TestMemLanes2_RecallAccessCarriesOrgWideWriteOnlyForScopedRootAdmin(t *testing.T) {
	s := &AdminServer{}
	for _, tc := range []struct {
		name string
		who  *RequestIdentity
		want bool
	}{
		{"operator with memory:write", memoryUser("u-op", "op", "operator", "memory:write"), false},
		{"admin without memory:write", memoryUser("u-admin", "root", "admin", "groups:read"), false},
		{"admin with memory:write", memoryUser("u-root", "root", "admin", "memory:write"), true},
		{"admin with every scope", memoryUser("u-key", "key", "admin", "*"), true},
		{"no identity", nil, false},
	} {
		ctx := context.Background()
		if tc.who != nil {
			ctx = context.WithValue(ctx, ctxKeyIdentity, tc.who)
		}
		if got := s.recallAccessFor(ctx).OrgWideWrite; got != tc.want {
			t.Errorf("%s: OrgWideWrite=%t, want %t", tc.name, got, tc.want)
		}
	}
}
