package server

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// MCPL item 2: the discovered-tool cache lists a tool the live server no
// longer offers. The pool call fails after authorization; Core records the
// outcome instead of answering 502 with nothing retained.

// mcplStaleHarness offers only read_text_file live but caches staleTool too.
func mcplStaleHarness(t *testing.T, staleTool string) *mcpsH {
	t.Helper()
	mcplWorkspace(t)
	h := newMCPSHarness(t, "filesystem", true, "read_text_file")
	h.tools = append(h.tools, staleTool)
	return h
}

// mcplExpectAuditWithID is expectAudit that also captures the row id.
func mcplExpectAuditWithID(h *mcpsH, fail bool) (id, ctx *mcpsCapture) {
	id, ctx = &mcpsCapture{}, &mcpsCapture{}
	exp := h.auditMock.ExpectExec("INSERT INTO log_entries").
		WithArgs(id, sqlmock.AnyArg(), sqlmock.AnyArg(), "audit", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), ctx)
	if fail {
		exp.WillReturnError(errors.New("audit store down"))
	} else {
		exp.WillReturnResult(sqlmock.NewResult(0, 1))
	}
	return id, ctx
}

func TestMcplStaleHighRiskCallRecordsFailureOutcome(t *testing.T) {
	h := mcplStaleHarness(t, "write_file")
	h.expectResolve()
	calledID, called := mcplExpectAuditWithID(h, false)
	failedID, failed := mcplExpectAuditWithID(h, false)
	payload := h.expectExchange()

	rr := h.callTool(mcpsWebAdmin(), "write_file", `{"arguments":{"path":"notes/x.md","content":"secret-body"}}`)

	assertStatus(t, rr, http.StatusBadGateway)
	h.assertCalls(0) // the live server has no such tool; nothing ran
	h.assertMocksMet()
	if called.object(t)["action"] != "mcp_tool_called" {
		t.Fatalf("first audit = %s", called.value())
	}
	ctx := failed.object(t)
	if ctx["action"] != "mcp_tool_call_failed" || ctx["audit_event_id"] != calledID.value() || calledID.value() == "" {
		t.Fatalf("failure audit = %s, want link to %s", failed.value(), calledID.value())
	}
	if ctx["failed_stage"] != "call" || ctx["tool"] != "write_file" {
		t.Fatalf("failure audit fields = %v", ctx)
	}
	for label, text := range map[string]string{"failure audit": failed.value(), "response": rr.Body.String()} {
		if mcplContainsAny(text, "secret-body", "notes/x.md") {
			t.Fatalf("%s holds an argument value: %s", label, text)
		}
	}
	item := payload.object(t)
	result := mcpsRetainedResult(t, payload)
	if item["status"] != "failed" || result["audit_event_id"] != calledID.value() || result["failure_audit_event_id"] != failedID.value() {
		t.Fatalf("exchange item = %s", payload.value())
	}
	env := decodeBlocker(t, rr)
	if env.OK || env.Error == "" || env.Data["exchange_item_id"] == nil || env.Data["audit_event_id"] != calledID.value() {
		t.Fatalf("502 body = %s", rr.Body.String())
	}
}

// The failure record is best-effort: an audit outage still answers 502 and
// still retains the failed Exchange item.
func TestMcplStaleFailureAuditOutageStillRetains(t *testing.T) {
	h := mcplStaleHarness(t, "write_file")
	h.expectResolve()
	mcplExpectAuditWithID(h, false)
	mcplExpectAuditWithID(h, true)
	payload := h.expectExchange()

	assertStatus(t, h.callTool(mcpsWebAdmin(), "write_file", `{"arguments":{"path":"a.md"}}`), http.StatusBadGateway)
	h.assertMocksMet()
	if result := mcpsRetainedResult(t, payload); result["failure_audit_event_id"] != nil {
		t.Fatalf("unwritten failure audit is linked: %v", result)
	}
}

// A low-risk stale read writes no audit but still retains a failed item.
func TestMcplStaleReadRetainsFailedItemWithoutAudit(t *testing.T) {
	h := mcplStaleHarness(t, "read_file")
	h.expectResolve()
	payload := h.expectExchange()

	rr := h.callTool(standardUserIdentity(), "read_file", `{"arguments":{"path":"a.md"}}`)

	assertStatus(t, rr, http.StatusBadGateway)
	h.assertMocksMet()
	if payload.object(t)["status"] != "failed" {
		t.Fatalf("exchange item = %s", payload.value())
	}
	if env := decodeBlocker(t, rr); env.Data["audit_event_id"] != nil || env.Data["exchange_item_id"] == nil {
		t.Fatalf("502 body = %s", rr.Body.String())
	}
}

func mcplContainsAny(text string, needles ...string) bool {
	for _, needle := range needles {
		if needle != "" && strings.Contains(text, needle) {
			return true
		}
	}
	return false
}
