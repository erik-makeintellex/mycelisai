package mcp

import "testing"

// MCPL item 3: NormalizeToolSetScope is the one exported scope normalizer.
func TestMcplNormalizeToolSetScopeExported(t *testing.T) {
	ok := map[[2]string][2]string{
		{"", ""}:          {"all", ""},
		{" ALL ", "x"}:    {"all", ""},
		{"Group", " g1 "}: {"group", "g1"},
		{" host ", "h-1"}: {"host", "h-1"},
	}
	for in, want := range ok {
		kind, ref, err := NormalizeToolSetScope(in[0], in[1])
		if err != nil || kind != want[0] || ref != want[1] {
			t.Fatalf("NormalizeToolSetScope(%q,%q) = %q,%q,%v, want %q,%q", in[0], in[1], kind, ref, err, want[0], want[1])
		}
	}
	for _, in := range [][2]string{{"group", " "}, {"host", ""}, {"team", "t"}} {
		if _, _, err := NormalizeToolSetScope(in[0], in[1]); err == nil {
			t.Fatalf("NormalizeToolSetScope(%q,%q) accepted an invalid scope", in[0], in[1])
		}
	}
}
