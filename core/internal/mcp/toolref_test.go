package mcp

import (
	"testing"
)

func TestIsMCPRef(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"mcp:filesystem/read_file", true},
		{"mcp:github/*", true},
		{"mcp:fetch", true},
		{"read_file", false},
		{"toolset:workspace", false},
		{"", false},
		{"mc:something", false},
	}
	for _, tt := range tests {
		if got := IsMCPRef(tt.input); got != tt.want {
			t.Errorf("IsMCPRef(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestIsToolSetRef(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"toolset:workspace", true},
		{"toolset:development", true},
		{"mcp:filesystem/*", false},
		{"read_file", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := IsToolSetRef(tt.input); got != tt.want {
			t.Errorf("IsToolSetRef(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestToolSetName(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"toolset:workspace", "workspace"},
		{"toolset:development", "development"},
		{"mcp:filesystem/*", ""},
		{"read_file", ""},
	}
	for _, tt := range tests {
		if got := ToolSetName(tt.input); got != tt.want {
			t.Errorf("ToolSetName(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestParseToolRef(t *testing.T) {
	tests := []struct {
		input      string
		wantNil    bool
		wantServer string
		wantTool   string
	}{
		{"mcp:filesystem/read_file", false, "filesystem", "read_file"},
		{"mcp:github/*", false, "github", "*"},
		{"mcp:filesystem", false, "filesystem", "*"}, // no slash → wildcard
		{"mcp:brave-search/web_search", false, "brave-search", "web_search"},
		{"read_file", true, "", ""},
		{"toolset:workspace", true, "", ""},
		{"", true, "", ""},
	}
	for _, tt := range tests {
		ref := ParseToolRef(tt.input)
		if tt.wantNil {
			if ref != nil {
				t.Errorf("ParseToolRef(%q) = %+v, want nil", tt.input, ref)
			}
			continue
		}
		if ref == nil {
			t.Errorf("ParseToolRef(%q) = nil, want non-nil", tt.input)
			continue
		}
		if ref.ServerName != tt.wantServer {
			t.Errorf("ParseToolRef(%q).ServerName = %q, want %q", tt.input, ref.ServerName, tt.wantServer)
		}
		if ref.ToolName != tt.wantTool {
			t.Errorf("ParseToolRef(%q).ToolName = %q, want %q", tt.input, ref.ToolName, tt.wantTool)
		}
	}
}

func TestToolRef_MatchesTool(t *testing.T) {
	tests := []struct {
		ref       ToolRef
		server    string
		tool      string
		wantMatch bool
	}{
		{ToolRef{"filesystem", "read_file"}, "filesystem", "read_file", true},
		{ToolRef{"filesystem", "read_file"}, "filesystem", "write_file", false},
		{ToolRef{"filesystem", "read_file"}, "github", "read_file", false},
		{ToolRef{"filesystem", "*"}, "filesystem", "read_file", true},
		{ToolRef{"filesystem", "*"}, "filesystem", "write_file", true},
		{ToolRef{"filesystem", "*"}, "github", "create_issue", false},
	}
	for _, tt := range tests {
		got := tt.ref.MatchesTool(tt.server, tt.tool)
		if got != tt.wantMatch {
			t.Errorf("ToolRef{%q,%q}.MatchesTool(%q,%q) = %v, want %v",
				tt.ref.ServerName, tt.ref.ToolName, tt.server, tt.tool, got, tt.wantMatch)
		}
	}
}
