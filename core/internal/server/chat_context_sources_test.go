package server

import (
	"encoding/json"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestChatContextSources_TravelFromAgentResultToChatPayload(t *testing.T) {
	raw := `{"text":"Blueberry-lavender scones, 3 for $10. See you at the counter!","provider_id":"",
"context_sources":[{"artifact_id":"a-1","title":"Juniper & Rye Bakery","knowledge_class":"company_knowledge","retrieval_mode":"keyword","used":true}]}`
	result := decodeChatAgentResult([]byte(raw))
	if len(result.ContextSources) != 1 || !result.ContextSources[0].Used {
		t.Fatalf("agent result must decode context_sources: %+v", result)
	}
	payload := &protocol.ChatResponsePayload{Text: result.Text}
	applyBrainProvenance(&AdminServer{}, payload, result)
	if len(payload.ContextSources) != 1 || payload.ContextSources[0].Title != "Juniper & Rye Bakery" {
		t.Fatalf("context sources must be copied even without brain provenance: %+v", payload)
	}
	encoded, _ := json.Marshal(payload)
	var decoded map[string]any
	_ = json.Unmarshal(encoded, &decoded)
	sources, _ := decoded["context_sources"].([]any)
	first, _ := sources[0].(map[string]any)
	if first["used"] != true || first["retrieval_mode"] != "keyword" {
		t.Fatalf("chat payload JSON must carry context_sources: %s", encoded)
	}
}
