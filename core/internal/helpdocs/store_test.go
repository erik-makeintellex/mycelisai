package helpdocs

import (
	"strings"
	"testing"
)

func TestStoreReadAndSearchCuratedDocs(t *testing.T) {
	store, err := NewStore("")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}

	doc, err := store.Read("soma-chat")
	if err != nil {
		t.Fatalf("Read(soma-chat): %v", err)
	}
	if doc.Path != "docs/user/soma-chat.md" {
		t.Fatalf("unexpected doc path: %s", doc.Path)
	}
	if !strings.Contains(doc.Content, "# Using Soma Chat") {
		t.Fatalf("expected Soma chat document content, got: %.120s", doc.Content)
	}

	results, err := store.Search("governed context memory", 3)
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) == 0 {
		t.Fatalf("expected search results")
	}
	for _, result := range results {
		if result.Slug == "" || result.Path == "" || result.Excerpt == "" {
			t.Fatalf("search result missing citation fields: %+v", result)
		}
	}
}

func TestStoreReadUnknownSlug(t *testing.T) {
	store, err := NewStore("")
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	if _, err := store.Read("missing-doc"); err == nil {
		t.Fatalf("expected unknown slug error")
	}
}
