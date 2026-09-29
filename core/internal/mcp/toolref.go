package mcp

import "strings"

// ToolRef represents a parsed MCP tool reference from an agent manifest.
// Format: "mcp:<server_name>/<tool_name>" where tool_name may be "*" for wildcard.
type ToolRef struct {
	ServerName string
	ToolName   string // "*" means all tools from this server
}

// ParseToolRef parses an "mcp:server/tool" string into a ToolRef.
// Returns nil if the string is not an MCP reference.
func ParseToolRef(s string) *ToolRef {
	if !IsMCPRef(s) {
		return nil
	}
	body := s[4:] // strip "mcp:"
	slash := strings.IndexByte(body, '/')
	if slash < 0 {
		// "mcp:filesystem" with no slash → treat as wildcard
		return &ToolRef{ServerName: body, ToolName: "*"}
	}
	return &ToolRef{
		ServerName: body[:slash],
		ToolName:   body[slash+1:],
	}
}

// IsMCPRef returns true if the tool name starts with "mcp:".
func IsMCPRef(s string) bool {
	return strings.HasPrefix(s, "mcp:")
}

// IsToolSetRef returns true if the tool name starts with "toolset:".
func IsToolSetRef(s string) bool {
	return strings.HasPrefix(s, "toolset:")
}

// ToolSetName extracts the set name from a "toolset:name" reference.
func ToolSetName(s string) string {
	if !IsToolSetRef(s) {
		return ""
	}
	return s[8:] // strip "toolset:"
}

// MatchesTool checks if this ToolRef matches a given server + tool combination.
// Wildcard "*" matches any tool on the server.
func (ref *ToolRef) MatchesTool(serverName, toolName string) bool {
	if ref.ServerName != serverName {
		return false
	}
	if ref.ToolName == "*" {
		return true
	}
	return ref.ToolName == toolName
}
