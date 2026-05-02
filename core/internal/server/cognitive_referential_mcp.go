package server

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/mycelis/core/internal/mcp"
)

func (s *AdminServer) summarizeRegisteredMCP(ctx context.Context) ([]string, []string) {
	if s.MCP == nil || s.MCP.DB == nil {
		return []string{"MCP registry unavailable in this runtime"}, nil
	}
	servers, err := s.MCP.List(ctx)
	if err != nil {
		return []string{"MCP registry lookup failed"}, nil
	}
	tools, err := s.MCP.ListAllTools(ctx)
	if err != nil {
		tools = nil
	}
	serverSummaries := make([]string, 0, len(servers))
	for _, srv := range servers {
		status := strings.TrimSpace(srv.Status)
		if status == "" {
			status = "installed"
		}
		serverSummaries = append(serverSummaries, fmt.Sprintf("%s (%s)", srv.Name, status))
	}
	toolSummaries := make([]string, 0, len(tools))
	for _, tool := range tools {
		name := strings.TrimSpace(tool.Name)
		if tool.ServerName != "" {
			name = tool.ServerName + "/" + name
		}
		if name != "" {
			toolSummaries = append(toolSummaries, name)
		}
	}
	sort.Strings(serverSummaries)
	sort.Strings(toolSummaries)
	return truncateStrings(serverSummaries, 8), truncateStrings(toolSummaries, 10)
}

func (s *AdminServer) summarizeRelevantMCPLibrary(text string, limit int) []string {
	if s.MCPLibrary == nil {
		return nil
	}
	tokens := intentTokenSet(text)
	var matches []string
	for _, category := range s.MCPLibrary.Categories {
		for _, entry := range category.Servers {
			parts := []string{entry.Name, entry.Title, entry.Description}
			haystack := strings.Join(append(parts, entry.Tags...), " ")
			if sharedIntentTokenCount(tokens, intentTokenSet(haystack)) == 0 {
				continue
			}
			label := entry.Name
			if envs := requiredMCPEnvNames(entry.EnvironmentVariables); len(envs) > 0 {
				label += " requires " + strings.Join(envs, "/")
			}
			matches = append(matches, label)
		}
	}
	sort.Strings(matches)
	return truncateStrings(uniqueOrderedTools(matches), limit)
}

func requiredMCPEnvNames(envVars []mcp.LibraryEnvVar) []string {
	var envs []string
	for _, env := range envVars {
		if env.Required {
			envs = append(envs, env.Name)
		}
	}
	return envs
}

func buildSomaMCPConfigurationAdvice(text string, registered, library []string) []string {
	lower := normalizeIntentText(text)
	var advice []string
	if strings.Contains(lower, "web") || strings.Contains(lower, "search") {
		advice = append(advice, "For web search, prefer Mycelis web_search with local_sources, searxng, or local_api; Brave is optional and needs BRAVE_API_KEY in .env only.")
	}
	if len(registered) == 0 || (len(registered) == 1 && strings.Contains(registered[0], "unavailable")) {
		advice = append(advice, "Open Resources -> Connected Tools to inspect installed MCP servers, then use Library to install or reapply curated entries.")
	} else if len(library) > 0 {
		advice = append(advice, "Use Resources -> Connected Tools -> Library for missing servers, then bind the resulting tool refs to the team or member template.")
	}
	return advice
}
