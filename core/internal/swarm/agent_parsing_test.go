package swarm

import "testing"

func TestParseConversationPayload_PlainText(t *testing.T) {
	a := &Agent{}

	input, history := a.parseConversationPayload([]byte("hello world"))
	if input != "hello world" {
		t.Fatalf("input = %q, want plain payload", input)
	}
	if history != nil {
		t.Fatalf("history = %#v, want nil", history)
	}
}

func TestParseConversationPayload_JSONConversation(t *testing.T) {
	a := &Agent{}
	payload := []byte(`[
		{"role":"user","content":"first"},
		{"role":"admin","content":"second"},
		{"role":"system","content":"third"},
		{"role":"user","content":"final"}
	]`)

	input, history := a.parseConversationPayload(payload)
	if input != "final" {
		t.Fatalf("input = %q, want final", input)
	}
	if len(history) != 3 {
		t.Fatalf("history len = %d, want 3", len(history))
	}
	if history[0].Role != "user" || history[0].Content != "first" {
		t.Fatalf("history[0] = %#v, want user/first", history[0])
	}
	if history[1].Role != "assistant" || history[1].Content != "second" {
		t.Fatalf("history[1] = %#v, want assistant/second", history[1])
	}
	if history[2].Role != "user" || history[2].Content != "third" {
		t.Fatalf("history[2] = %#v, want fallback user/third", history[2])
	}
}

func TestParseConversationPayload_InvalidJSONFallsBackToText(t *testing.T) {
	a := &Agent{}
	raw := `[{"role":"user","content":"x"}`
	input, history := a.parseConversationPayload([]byte(raw))
	if input != raw {
		t.Fatalf("input = %q, want raw payload", input)
	}
	if history != nil {
		t.Fatalf("history = %#v, want nil", history)
	}
}

func TestParseConversationPayload_EmptyArray(t *testing.T) {
	a := &Agent{}
	input, history := a.parseConversationPayload([]byte("[]"))
	if input != "" {
		t.Fatalf("input = %q, want empty", input)
	}
	if history != nil {
		t.Fatalf("history = %#v, want nil", history)
	}
}
