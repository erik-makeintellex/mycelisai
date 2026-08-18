package server

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/internal/swarm"
)

func TestConfirmedActionToolContextCarriesConfigDocumentActor(t *testing.T) {
	ctx := confirmedActionToolContext(
		t.Context(), " operator-1 ", "11111111-1111-1111-1111-111111111111",
	)
	invocation, ok := swarm.ToolInvocationContextFromContext(ctx)
	if !ok {
		t.Fatal("confirmed action context is missing invocation metadata")
	}
	if invocation.AgentID != "operator-1" || invocation.UserLabel != "operator-1" {
		t.Fatalf("actor metadata = %#v, want operator-1", invocation)
	}
}

func TestConfirmedActionActorUsesStableUserID(t *testing.T) {
	req := httptest.NewRequest("POST", "/api/v1/intent/confirm-action", nil)
	req = req.WithContext(context.WithValue(req.Context(), ctxKeyIdentity, &RequestIdentity{
		UserID: "user-immutable-1", Username: "changeable@example.com",
	}))
	if got := auditActorIDFromRequest(req); got != "user-immutable-1" {
		t.Fatalf("actor id = %q, want immutable user id", got)
	}
	if got := auditUserLabelFromRequest(req); got != "changeable@example.com" {
		t.Fatalf("user label = %q, want display username", got)
	}
}
