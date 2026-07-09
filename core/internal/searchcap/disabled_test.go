package searchcap

import (
	"context"
	"testing"
)

func TestServiceDisabledReturnsStructuredBlocker(t *testing.T) {
	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)

	resp, err := svc.Search(context.Background(), Request{Query: "can you search the web?", SourceScope: "web"})
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if resp.Status != "blocked" {
		t.Fatalf("Status = %q, want blocked", resp.Status)
	}
	if resp.Blocker == nil || resp.Blocker.Code != "search_provider_disabled" {
		t.Fatalf("Blocker = %+v", resp.Blocker)
	}
}
