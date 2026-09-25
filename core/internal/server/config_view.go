package server

import (
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/configview"
	"github.com/mycelis/core/pkg/protocol"
)

// View config (RC-004F): read-only, redacted configuration documents for the
// four config kinds that Deployments `config_ref.path` points at. Every
// document is built from ids and declared fields only, then rendered through
// configview.Render, which redacts secret values before any text exists.

// Config view kinds (singular, matching DeploymentConfigRef.Kind).
const (
	configKindTeam         = "team"
	configKindGroup        = "group"
	configKindOrganization = "organization"
	configKindTemplate     = "template"
)

// Config view sources.
const (
	configSourceFile     = "file"
	configSourceDatabase = "database"
	configSourceRuntime  = "runtime"
)

// configViewKinds maps the plural route segment to the kind.
var configViewKinds = map[string]string{
	"teams":         configKindTeam,
	"groups":        configKindGroup,
	"organizations": configKindOrganization,
	"templates":     configKindTemplate,
}

// ConfigView is the data payload of GET /api/v1/config/{kind}/{id}.
type ConfigView struct {
	Kind        string    `json:"kind"`
	ID          string    `json:"id"`
	Format      string    `json:"format"`
	Document    string    `json:"document"`
	Source      string    `json:"source"`
	SourcePath  string    `json:"source_path,omitempty"`
	Redactions  []string  `json:"redactions"`
	GeneratedAt time.Time `json:"generated_at"`
}

// configDocument is an unrendered document from one source adapter.
type configDocument struct {
	doc        any
	source     string
	sourcePath string
}

// configViewError carries the HTTP status for a source adapter failure.
type configViewError struct {
	status int
	msg    string
}

func (e *configViewError) Error() string { return e.msg }

func configNotFound(kind, id string) error {
	return &configViewError{status: http.StatusNotFound, msg: kind + " not found: " + id}
}

func configUnavailable(msg string) error {
	return &configViewError{status: http.StatusServiceUnavailable, msg: msg}
}

// GET /api/v1/config/{kind}/{id}?format=yaml|json
func (s *AdminServer) HandleGetConfigView(w http.ResponseWriter, r *http.Request) {
	segment := strings.TrimSpace(r.PathValue("kind"))
	id := strings.TrimSpace(r.PathValue("id"))
	kind, ok := configViewKinds[segment]
	if !ok {
		respondAPIError(w, "unknown config kind: "+segment+" (expected teams, groups, organizations, or templates)", http.StatusBadRequest)
		return
	}
	if id == "" {
		respondAPIError(w, "config id is required", http.StatusBadRequest)
		return
	}
	format, err := configview.ParseFormat(r.URL.Query().Get("format"))
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}
	if !s.authorizeConfigView(w, r, kind) {
		return
	}

	var doc *configDocument
	switch kind {
	case configKindTeam:
		doc, err = s.teamConfigDocument(r, id)
	case configKindGroup:
		doc, err = s.groupConfigDocument(r, id)
	case configKindOrganization:
		doc, err = s.organizationConfigDocument(r, id)
	case configKindTemplate:
		doc, err = s.templateConfigDocument(id)
	}
	if err != nil {
		var viewErr *configViewError
		if errors.As(err, &viewErr) {
			respondAPIError(w, viewErr.msg, viewErr.status)
			return
		}
		log.Printf("[config-view] %s/%s: %v", kind, id, err)
		respondAPIError(w, "config view unavailable", http.StatusInternalServerError)
		return
	}

	rendered, err := configview.Render(doc.doc, format)
	if err != nil {
		log.Printf("[config-view] render %s/%s: %v", kind, id, err)
		respondAPIError(w, "config document could not be rendered", http.StatusInternalServerError)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(ConfigView{
		Kind:        kind,
		ID:          id,
		Format:      rendered.Format,
		Document:    rendered.Document,
		Source:      doc.source,
		SourcePath:  doc.sourcePath,
		Redactions:  rendered.Redactions,
		GeneratedAt: time.Now().UTC(),
	}))
}

// authorizeConfigView applies the access rule of the underlying resource.
// Groups follow /api/v1/groups (root admin with groups:read). Teams,
// organizations, and templates are readable by any authenticated operator,
// as their own list/detail routes are.
func (s *AdminServer) authorizeConfigView(w http.ResponseWriter, r *http.Request, kind string) bool {
	if kind == configKindGroup {
		_, ok := requireRootAdminScope(w, r, "groups:read")
		return ok
	}
	if IdentityFromContext(r.Context()) == nil {
		respondAPIError(w, "Authentication required", http.StatusUnauthorized)
		return false
	}
	return true
}

// configDoc builds the common apiVersion/kind/metadata/spec shape.
func configDoc(kind string, metadata map[string]any, spec any) map[string]any {
	return map[string]any{
		"apiVersion": "mycelis/v1alpha1",
		"kind":       kind,
		"metadata":   metadata,
		"spec":       spec,
	}
}
