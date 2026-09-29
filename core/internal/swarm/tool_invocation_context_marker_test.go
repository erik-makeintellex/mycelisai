package swarm

import (
	"context"
	"testing"
)

// TestWithRecallAccessKeepsConfirmedDispatchMarker pins the SRU x TPD merge
// fix: setting the recall scope on a confirmed-dispatch context must not
// clear the Core-only marker, and must never grant it to a plain context.
func TestWithRecallAccessKeepsConfirmedDispatchMarker(t *testing.T) {
	confirmed := WithConfirmedDispatchToolContext(context.Background(), ToolInvocationContext{RunID: "run-1"})
	scoped := WithRecallAccess(confirmed, RecallAccess{})
	meta, ok := ToolInvocationContextFromContext(scoped)
	if !ok || !meta.confirmedDispatch || meta.RunID != "run-1" {
		t.Fatalf("confirmed-dispatch marker lost after WithRecallAccess: ok=%v meta=%+v", ok, meta)
	}

	plain := WithToolInvocationContext(context.Background(), ToolInvocationContext{RunID: "run-2"})
	meta, _ = ToolInvocationContextFromContext(WithRecallAccess(plain, RecallAccess{}))
	if meta.confirmedDispatch {
		t.Fatal("WithRecallAccess granted the confirmed-dispatch marker to a plain context")
	}

	if got := WithRecallAccess(context.Background(), RecallAccess{}); got != context.Background() {
		t.Fatal("WithRecallAccess must not invent an invocation context")
	}
}
