package swarm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
)

func TestCompactProjectPackageInferenceHistoryHasDeterministicCeiling(t *testing.T) {
	base := []cognitive.ChatMessage{{Role: "system", Content: "base policy"}, {Role: "user", Content: "authoritative contract"}}
	evidence := make([]successfulToolEvidence, 0, 100)
	for index := 0; index < 100; index++ {
		evidence = append(evidence, successfulToolEvidence{
			ToolName: "write_file", Path: strings.Repeat("segment/", 80) + fmt.Sprint(index), Content: strings.Repeat("SECRET-CONTENT", 100),
		})
	}
	latest := cognitive.ChatMessage{Role: "user", Content: strings.Repeat("correction ", 500)}
	messages := compactProjectPackageInferenceHistory(base, evidence, latest)
	if len(messages) != len(base)+2 {
		t.Fatalf("message count = %d, want %d", len(messages), len(base)+2)
	}
	addedChars := 0
	for _, message := range messages[len(base):] {
		addedChars += len(message.Content)
	}
	maxAdded := 100 + maxProjectPackageLedgerEntries*(30+maxProjectPackageLedgerPath) + maxProjectPackageLatestChars + 3
	if addedChars > maxAdded {
		t.Fatalf("added chars = %d, ceiling = %d", addedChars, maxAdded)
	}
	if strings.Contains(messagesText(messages), "SECRET-CONTENT") {
		t.Fatal("ledger exposed file contents")
	}
}

func TestProjectPackageInferenceBaseKeepsSystemsAndLatestUserOnly(t *testing.T) {
	requirement := &teamResultRequirement{Kind: "project_package"}
	messages := []cognitive.ChatMessage{
		{Role: "system", Content: "base"}, {Role: "user", Content: "old request"},
		{Role: "assistant", Content: "malformed secret response"}, {Role: "system", Content: "contract"},
		{Role: "user", Content: "authoritative request"},
	}
	base := projectPackageInferenceBase(messages, requirement)
	joined := messagesText(base)
	for _, want := range []string{"base", "contract", "authoritative request"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("base missing %q: %#v", want, base)
		}
	}
	for _, forbidden := range []string{"old request", "malformed secret response"} {
		if strings.Contains(joined, forbidden) {
			t.Fatalf("base retained superseded %q: %#v", forbidden, base)
		}
	}
}

func TestCompactProjectPackageInferenceHistoryPreservesLatestInterjection(t *testing.T) {
	interjection := "[OPERATOR INTERJECTION]: Keep the moon gate and use violet lighting."
	messages := compactProjectPackageInferenceHistory(
		[]cognitive.ChatMessage{{Role: "system", Content: "base"}, {Role: "user", Content: "contract"}}, nil,
		cognitive.ChatMessage{Role: "user", Content: interjection},
	)
	if messages[len(messages)-1].Content != interjection {
		t.Fatalf("interjection changed: %#v", messages)
	}
}

func messagesText(messages []cognitive.ChatMessage) string {
	var builder strings.Builder
	for _, message := range messages {
		builder.WriteString(message.Content)
		builder.WriteByte('\n')
	}
	return builder.String()
}
