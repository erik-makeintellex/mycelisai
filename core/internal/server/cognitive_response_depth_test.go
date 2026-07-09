package server

import (
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestInferResponseDepthFromRequest(t *testing.T) {
	tests := []struct {
		name       string
		request    string
		isMutation bool
		want       protocol.ResponseDepth
	}{
		{
			name:    "quick links",
			request: "Just show the top 5 links.",
			want:    protocol.ResponseDepthQuickBox,
		},
		{
			name:    "summary comparison",
			request: "Compare these options and summarize the main patterns.",
			want:    protocol.ResponseDepthStructuredSummary,
		},
		{
			name:    "decision brief",
			request: "What should we do and what is your recommendation?",
			want:    protocol.ResponseDepthDecisionBrief,
		},
		{
			name:       "execution proposal",
			request:    "Give me a brief list, then create the file.",
			isMutation: true,
			want:       protocol.ResponseDepthExecutionProposal,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := inferResponseDepthFromRequest(tc.request, tc.isMutation)
			if got != tc.want {
				t.Fatalf("depth = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResponseDepthDoesNotInferMutation(t *testing.T) {
	request := "Give me a detailed comparison table of the current options."
	if tools := inferMutationToolsFromText(request); len(tools) != 0 {
		t.Fatalf("mutation tools = %#v, want none for depth-only request", tools)
	}
	if got := inferResponseDepthFromRequest(request, false); got != protocol.ResponseDepthStructuredSummary {
		t.Fatalf("depth = %q, want %q", got, protocol.ResponseDepthStructuredSummary)
	}
}
