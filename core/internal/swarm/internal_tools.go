package swarm

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/mycelis/core/internal/catalogue"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/comms"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/inception"
	"github.com/mycelis/core/internal/memory"
	"github.com/nats-io/nats.go"
)

// Phase 0 security: workspace sandbox for file tools
const maxWriteSize = 1 << 20 // 1 MB

func normalizeWorkspaceRelativePath(rawPath string) string {
	trimmed := strings.TrimSpace(rawPath)
	if trimmed == "" || filepath.IsAbs(trimmed) {
		return trimmed
	}

	normalized := strings.ReplaceAll(trimmed, "\\", "/")
	normalized = path.Clean(normalized)
	normalized = strings.TrimPrefix(normalized, "./")

	switch normalized {
	case ".", "workspace":
		return "."
	}
	if strings.HasPrefix(normalized, "workspace/") {
		normalized = strings.TrimPrefix(normalized, "workspace/")
	}
	return filepath.FromSlash(normalized)
}

func validateToolPath(rawPath string) (string, error) {
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("invalid workspace path: %w", err)
	}

	var absTarget string
	if filepath.IsAbs(rawPath) {
		absTarget = filepath.Clean(rawPath)
	} else {
		absTarget = filepath.Clean(filepath.Join(absWorkspace, normalizeWorkspaceRelativePath(rawPath)))
	}

	rel, err := filepath.Rel(absWorkspace, absTarget)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %q escapes workspace boundary", rawPath)
	}

	checkPath := absTarget
	if _, statErr := os.Lstat(absTarget); os.IsNotExist(statErr) {
		checkPath = filepath.Dir(absTarget)
	}
	if realPath, evalErr := filepath.EvalSymlinks(checkPath); evalErr == nil {
		relReal, _ := filepath.Rel(absWorkspace, realPath)
		if strings.HasPrefix(relReal, "..") {
			return "", fmt.Errorf("symlink at %q resolves outside workspace", rawPath)
		}
	}
	return absTarget, nil
}

// InternalTool describes a built-in tool available to agents without an external MCP server.
type InternalTool struct {
	Name        string
	Description string
	InputSchema map[string]any
	Handler     func(ctx context.Context, args map[string]any) (string, error)
}

// InternalToolRegistry holds all built-in tools and their dependencies.
type InternalToolRegistry struct {
	tools     map[string]*InternalTool
	nc        *nats.Conn
	brain     *cognitive.Router
	mem       *memory.Service
	architect *cognitive.MetaArchitect
	catalogue *catalogue.Service
	inception *inception.Store
	comms     *comms.Gateway
	db        *sql.DB
	exchange  *exchange.Service
	somaRef   *Soma
}

// InternalToolDeps bundles all optional dependencies for the internal tools.
type InternalToolDeps struct {
	NC        *nats.Conn
	Brain     *cognitive.Router
	Mem       *memory.Service
	Architect *cognitive.MetaArchitect
	Catalogue *catalogue.Service
	Inception *inception.Store
	Comms     *comms.Gateway
	DB        *sql.DB
	Exchange  *exchange.Service
}

// NewInternalToolRegistry creates and populates the built-in tool set.
func NewInternalToolRegistry(deps InternalToolDeps) *InternalToolRegistry {
	r := &InternalToolRegistry{
		tools:     make(map[string]*InternalTool),
		nc:        deps.NC,
		brain:     deps.Brain,
		mem:       deps.Mem,
		architect: deps.Architect,
		catalogue: deps.Catalogue,
		inception: deps.Inception,
		comms:     deps.Comms,
		db:        deps.DB,
		exchange:  deps.Exchange,
	}
	r.registerAll()
	return r
}

// SetSoma wires the Soma reference after construction.
func (r *InternalToolRegistry) SetSoma(s *Soma) {
	r.somaRef = s
}

func (r *InternalToolRegistry) Get(name string) *InternalTool { return r.tools[name] }

func (r *InternalToolRegistry) Has(name string) bool {
	_, ok := r.tools[name]
	return ok
}

func (r *InternalToolRegistry) ListNames() []string {
	names := make([]string, 0, len(r.tools))
	for k := range r.tools {
		names = append(names, k)
	}
	return names
}

func (r *InternalToolRegistry) ListDescriptions() map[string]string {
	m := make(map[string]string, len(r.tools))
	for k, v := range r.tools {
		m[k] = v.Description
	}
	return m
}

// registerAll keeps registration grouped by operator-facing capability families.
func (r *InternalToolRegistry) registerAll() {
	r.registerCoordinationTools()
	r.registerExchangeAndPlanningTools()
	r.registerMemoryAndArtifactTools()
	r.registerExecutionAndMediaTools()
}
