package server

import "testing"

func TestExtractRequestedPackageTitleUsesNaturalTitlePhrases(t *testing.T) {
	tests := map[string]string{
		"Use the package title Moonlit Keep First Playable.":                          "Moonlit Keep First Playable",
		"Build an original game titled Moonlit Keep First Playable. Keep it compact.": "Moonlit Keep First Playable",
		"Return a package title Moonlit Keep First Playable\nThen validate it.":       "Moonlit Keep First Playable",
	}
	for request, expected := range tests {
		if actual := extractRequestedPackageTitle(request); actual != expected {
			t.Fatalf("extractRequestedPackageTitle(%q) = %q, want %q", request, actual, expected)
		}
	}
}
