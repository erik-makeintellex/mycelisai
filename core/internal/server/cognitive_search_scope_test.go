package server

import "testing"

func TestDirectSearchScopeInference(t *testing.T) {
	tests := []struct {
		name  string
		query string
		want  string
	}{
		{
			name:  "generic search defaults to web",
			query: "search on what's the latest popular multi agent framework?",
			want:  "web",
		},
		{
			name:  "local retained context",
			query: "search local sources for the latest retained design note",
			want:  "local_sources",
		},
		{
			name:  "explicit local only",
			query: "search retained Mycelis context for outcome packages",
			want:  "local_sources",
		},
		{
			name:  "mixed local and web",
			query: "compare internal docs and public web research for agent frameworks",
			want:  "all",
		},
		{
			name:  "explicit public web",
			query: "search the public web for agent frameworks",
			want:  "web",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := inferDirectSearchSourceScope(tt.query); got != tt.want {
				t.Fatalf("inferDirectSearchSourceScope(%q) = %q, want %q", tt.query, got, tt.want)
			}
		})
	}
}

func TestSearchCapabilityQuestionAllowsConcreteSearchAsk(t *testing.T) {
	prompt := "can you search on what's the latest popular multi agent framework?"

	if isSearchCapabilityQuestion(prompt) {
		t.Fatal("concrete search ask should not be treated as a capability question")
	}
	query, ok := shouldHandleDirectSearch(prompt)
	if !ok {
		t.Fatal("concrete search ask should route to direct search")
	}
	if query != "what's the latest popular multi agent framework" {
		t.Fatalf("query = %q, want extracted topic", query)
	}
}

func TestSearchCapabilityQuestionStillHandlesAbilityQuestion(t *testing.T) {
	prompt := "can you search the web?"

	if !isSearchCapabilityQuestion(prompt) {
		t.Fatal("ability question should still receive capability summary")
	}
}

func TestDirectSearchStillHandlesPlainLookupPrompt(t *testing.T) {
	prompt := "look up latest AI agent architecture research"

	if query, ok := shouldHandleDirectSearch(prompt); !ok || query != "latest AI agent architecture research" {
		t.Fatalf("plain lookup direct search = (%q, %v), want extracted query and ok", query, ok)
	}
}

func TestDirectSearchHandlesMixedPublicWebResearchPrompt(t *testing.T) {
	prompt := "compare internal docs and public web research for agent frameworks"

	if query, ok := shouldHandleDirectSearch(prompt); !ok || query != prompt {
		t.Fatalf("mixed research direct search = (%q, %v), want original query and ok", query, ok)
	}
	if scope := inferDirectSearchSourceScope(prompt); scope != "all" {
		t.Fatalf("scope = %q, want all", scope)
	}
}

func TestDirectSearchRequestCarriesScopeFromOriginalPrompt(t *testing.T) {
	request, ok := shouldHandleDirectSearchRequest("search the public web for agent frameworks")
	if !ok {
		t.Fatal("direct search request should be detected")
	}
	if request.Query != "agent frameworks" {
		t.Fatalf("query = %q, want extracted topic", request.Query)
	}
	if request.SourceScope != "web" {
		t.Fatalf("source scope = %q, want web", request.SourceScope)
	}
}
