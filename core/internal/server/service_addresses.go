package server

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/mycelis/core/internal/cognitive"
)

// ServiceAddressEntry describes the resolved address and configuration source for one backend dependency.
type ServiceAddressEntry struct {
	Name            string `json:"name"`
	Address         string `json:"address"`
	Protocol        string `json:"protocol"`
	Source          string `json:"source"`
	LoopbackWarning string `json:"loopback_warning,omitempty"`
}

// ServiceAddressBook captures all canonical backend service addresses and cluster-mode posture.
type ServiceAddressBook struct {
	ClusterMode bool                  `json:"cluster_mode"`
	Entries     []ServiceAddressEntry `json:"entries"`
	Warnings    []string              `json:"warnings,omitempty"`
}

// IsClusterRuntime reports whether Mycelis Core is executing inside a Kubernetes cluster runtime.
func IsClusterRuntime() bool {
	if strings.TrimSpace(os.Getenv("KUBERNETES_SERVICE_HOST")) != "" {
		return true
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("MYCELIS_DEPLOYMENT_ENV")), "kubernetes") {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(os.Getenv("MYCELIS_K8S_CLUSTER_MODE")), "true")
}

// IsLoopbackServiceAddress reports whether addr points at pod-local loopback.
func IsLoopbackServiceAddress(addr string) bool {
	trimmed := strings.TrimSpace(addr)
	if trimmed == "" {
		return false
	}

	host := trimmed
	if strings.Contains(trimmed, "://") {
		parsed, err := url.Parse(trimmed)
		if err != nil {
			return false
		}
		host = parsed.Hostname()
	} else if parsedHost, _, err := net.SplitHostPort(trimmed); err == nil {
		host = parsedHost
	} else {
		host = strings.Trim(trimmed, "[]")
	}

	switch strings.ToLower(strings.TrimSpace(host)) {
	case "localhost", "127.0.0.1", "::1", "0.0.0.0":
		return true
	default:
		return false
	}
}

// ResolveServiceAddressBook resolves canonical backend service addresses in precedence order.
func ResolveServiceAddressBook(cog *cognitive.Router) ServiceAddressBook {
	clusterMode := IsClusterRuntime()

	natsAddr, natsSrc := resolveEnvChain([]string{"MYCELIS_SERVICE_NATS_ADDRESS", "NATS_URL"}, "nats://localhost:4222")
	pgAddr, pgSrc := resolvePostgresServiceAddress()
	textAddr, textSrc := resolveTextAIServiceAddress(cog)
	mediaAddr, mediaSrc := resolveMediaAIServiceAddress(cog)
	searchAddr, searchSrc := resolveEnvChain([]string{"MYCELIS_SERVICE_SEARCH_ADDRESS", "MYCELIS_SEARXNG_ENDPOINT", "MYCELIS_SEARCH_LOCAL_API_ENDPOINT"}, "")
	mcpAddr, mcpSrc := resolveEnvChain([]string{"MYCELIS_SERVICE_MCP_GATEWAY_ADDRESS"}, "")
	corsAddr, corsSrc := resolveEnvChain([]string{"MYCELIS_SERVICE_CORS_ORIGIN", "CORS_ORIGIN"}, "http://localhost:3000")

	entries := []ServiceAddressEntry{
		{Name: "nats", Address: natsAddr, Protocol: "nats", Source: natsSrc},
		{Name: "postgres", Address: pgAddr, Protocol: "postgres", Source: pgSrc},
		{Name: "text_ai", Address: textAddr, Protocol: "http", Source: textSrc},
		{Name: "media_ai", Address: mediaAddr, Protocol: "http", Source: mediaSrc},
		{Name: "search", Address: searchAddr, Protocol: "http", Source: searchSrc},
		{Name: "mcp_gateway", Address: mcpAddr, Protocol: "http", Source: mcpSrc},
		{Name: "cors_origin", Address: corsAddr, Protocol: "http", Source: corsSrc},
	}

	var warnings []string
	if clusterMode {
		for i := range entries {
			if entries[i].Name == "cors_origin" || entries[i].Address == "" {
				continue
			}
			if IsLoopbackServiceAddress(entries[i].Address) {
				entries[i].LoopbackWarning = fmt.Sprintf(
					"%s address %q uses loopback inside a Kubernetes pod; configure a cluster DNS service name or reachable host/external service address",
					entries[i].Name,
					entries[i].Address,
				)
				warnings = append(warnings, entries[i].LoopbackWarning)
			}
		}
	}

	return ServiceAddressBook{
		ClusterMode: clusterMode,
		Entries:     entries,
		Warnings:    warnings,
	}
}

// AddressFor returns the resolved address for a named service entry.
func (b ServiceAddressBook) AddressFor(name string) string {
	if entry, ok := b.EntryFor(name); ok {
		return entry.Address
	}
	return ""
}

// EntryFor returns the full ServiceAddressEntry for a named service.
func (b ServiceAddressBook) EntryFor(name string) (ServiceAddressEntry, bool) {
	for _, entry := range b.Entries {
		if entry.Name == name {
			return entry, true
		}
	}
	return ServiceAddressEntry{}, false
}

func resolveEnvChain(keys []string, fallback string) (string, string) {
	for _, key := range keys {
		if val := strings.TrimSpace(os.Getenv(key)); val != "" {
			return val, key
		}
	}
	if fallback != "" {
		return fallback, "default"
	}
	return "", "unconfigured"
}

func resolvePostgresServiceAddress() (string, string) {
	if raw := strings.TrimSpace(os.Getenv("MYCELIS_SERVICE_DB_ADDRESS")); raw != "" {
		return raw, "MYCELIS_SERVICE_DB_ADDRESS"
	}
	host := strings.TrimSpace(os.Getenv("DB_HOST"))
	port := strings.TrimSpace(os.Getenv("DB_PORT"))
	source := "default"
	if host != "" || port != "" {
		source = "DB_HOST/DB_PORT"
	}
	if host == "" {
		host = "mycelis-core-postgresql"
	}
	if port == "" {
		port = "5432"
	}
	return host + ":" + port, source
}

func resolveTextAIServiceAddress(cog *cognitive.Router) (string, string) {
	if addr, src := resolveEnvChain([]string{
		"MYCELIS_SERVICE_TEXT_AI_ADDRESS",
		"MYCELIS_K8S_TEXT_ENDPOINT",
		"MYCELIS_PROVIDER_OLLAMA_ENDPOINT",
	}, ""); addr != "" {
		return addr, src
	}
	if cog != nil && cog.Config != nil {
		if ollamaCfg, ok := cog.Config.Providers["ollama"]; ok && strings.TrimSpace(ollamaCfg.Endpoint) != "" {
			return strings.TrimSpace(ollamaCfg.Endpoint), "cognitive.providers.ollama"
		}
	}
	return "http://127.0.0.1:11434/v1", "default"
}

func resolveMediaAIServiceAddress(cog *cognitive.Router) (string, string) {
	if addr, src := resolveEnvChain([]string{
		"MYCELIS_SERVICE_MEDIA_AI_ADDRESS",
		"MYCELIS_MEDIA_ENDPOINT",
		"MYCELIS_K8S_MEDIA_ENDPOINT",
	}, ""); addr != "" {
		return addr, src
	}
	if cog != nil && cog.Config != nil && cog.Config.Media != nil {
		mp := cog.Config.Media.EffectiveProvider()
		if mp.IsEnabled() && strings.TrimSpace(mp.Endpoint) != "" {
			return strings.TrimSpace(mp.Endpoint), "cognitive.media"
		}
	}
	return "", "unconfigured"
}

// GET /api/v1/services/addresses — canonical resolved backend service address book.
func (s *AdminServer) HandleServiceAddresses(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		respondError(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	book := ResolveServiceAddressBook(s.Cognitive)
	respondJSON(w, map[string]any{"ok": true, "data": book})
}
