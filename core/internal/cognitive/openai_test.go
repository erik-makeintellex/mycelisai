package cognitive

import (
	"testing"

	openai "github.com/sashabaranov/go-openai"
)

func TestNormalizeOpenAIMessage_PrefersStructuredToolCallsWhenContentEmpty(t *testing.T) {
	msg := openai.ChatCompletionMessage{
		ToolCalls: []openai.ToolCall{
			{
				Function: openai.FunctionCall{
					Name:      "write_file",
					Arguments: `{"path":"workspace/test.txt","content":"hello"}`,
				},
			},
		},
	}

	got := normalizeOpenAIMessage(msg)
	want := `{"tool_call":{"arguments":{"content":"hello","path":"workspace/test.txt"},"name":"write_file"}}`
	if got != want {
		t.Fatalf("normalizeOpenAIMessage() = %q, want %q", got, want)
	}
}

func TestNormalizeOpenAIMessage_FallsBackToFunctionCall(t *testing.T) {
	msg := openai.ChatCompletionMessage{
		FunctionCall: &openai.FunctionCall{
			Name:      "delegate",
			Arguments: `{"team_id":"admin-core"}`,
		},
	}

	got := normalizeOpenAIMessage(msg)
	want := `{"tool_call":{"arguments":{"team_id":"admin-core"},"name":"delegate"}}`
	if got != want {
		t.Fatalf("normalizeOpenAIMessage() = %q, want %q", got, want)
	}
}

func TestNormalizeOpenAIMessage_PrefersStructuredToolCallsOverProse(t *testing.T) {
	msg := openai.ChatCompletionMessage{
		Content: "Here is the draft answer, but I also need to call a tool.",
		ToolCalls: []openai.ToolCall{
			{
				Function: openai.FunctionCall{
					Name:      "write_file",
					Arguments: `{"path":"workspace/test.txt","content":"hello"}`,
				},
			},
		},
	}

	got := normalizeOpenAIMessage(msg)
	want := `{"tool_call":{"arguments":{"content":"hello","path":"workspace/test.txt"},"name":"write_file"}}`
	if got != want {
		t.Fatalf("normalizeOpenAIMessage() = %q, want %q", got, want)
	}
}
