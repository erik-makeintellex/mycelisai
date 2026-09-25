package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/bootstrap"
	"gopkg.in/yaml.v3"
)

// groupConfigDocument projects a collaboration group to the
// CollaborationGroup document shape that apply accepts (RC-004G).
func (s *AdminServer) groupConfigDocument(r *http.Request, id string) (*configDocument, error) {
	if _, err := uuid.Parse(id); err != nil {
		return nil, configNotFound("group", id)
	}
	if s.getDB() == nil {
		return nil, configUnavailable("database not configured; group config is unavailable")
	}
	g, err := s.getGroupDB(r.Context(), id)
	if err != nil {
		log.Printf("[config-view] group query: %v", err)
		return nil, configUnavailable("group query failed")
	}
	if g == nil {
		return nil, configNotFound("group", id)
	}
	meta := map[string]any{
		"id": g.ID, "name": g.Name, "tenant_id": g.TenantID, "status": g.Status,
		"created_by": g.CreatedBy, "created_at": configTime(g.CreatedAt), "updated_at": configTime(g.UpdatedAt),
	}
	return &configDocument{doc: configDoc("CollaborationGroup", meta, groupConfigSpec(g)), source: configSourceDatabase}, nil
}

// groupConfigSpec is the group spec block View config shows and apply accepts.
func groupConfigSpec(g *CollaborationGroup) map[string]any {
	spec := map[string]any{
		"name":                 g.Name,
		"goal_statement":       g.GoalStatement,
		"work_mode":            g.WorkMode,
		"approval_policy_ref":  g.ApprovalPolicyRef,
		"coordinator_profile":  g.CoordinatorProfile,
		"allowed_capabilities": stringsOrEmpty(g.AllowedCapabilities),
		"team_ids":             stringsOrEmpty(g.TeamIDs),
		"member_user_ids":      stringsOrEmpty(g.MemberUserIDs),
	}
	if g.Expiry != nil {
		spec["expiry"] = configTime(*g.Expiry)
	}
	return spec
}

// organizationConfigDocument returns the organization home document,
// including template id, delivery defaults, AI engine, and response contract.
func (s *AdminServer) organizationConfigDocument(r *http.Request, id string) (*configDocument, error) {
	store := s.organizationStore()
	home, ok, err := store.GetContext(r.Context(), id)
	if err != nil {
		log.Printf("[config-view] organization read: %v", err)
		return nil, configUnavailable("organization storage unavailable")
	}
	if !ok {
		return nil, configNotFound("organization", id)
	}
	home = normalizeOrganizationHome(home)
	raw, err := json.Marshal(home)
	if err != nil {
		return nil, err
	}
	var spec map[string]any
	if err := json.Unmarshal(raw, &spec); err != nil {
		return nil, err
	}
	meta := map[string]any{"id": home.ID, "name": home.Name, "template_id": home.TemplateID}
	source := configSourceDatabase
	if store.Backend() == "memory" {
		source = configSourceRuntime
	}
	return &configDocument{doc: configDoc("Organization", meta, spec), source: source}, nil
}

// templateConfigDocument returns a template bundle: a file bundle (a
// top-level bundle or a starter under <templates>/starters/) as loaded from
// disk, else an operator template (RC-004G clone) from operator_templates.
// Files are parsed generically so every field is shown; redaction still
// applies. Top-level bundles win an id clash, matching the starter catalog.
func (s *AdminServer) templateConfigDocument(r *http.Request, id string) (*configDocument, error) {
	doc, err := s.fileTemplateDocument(id)
	if err != nil || doc != nil {
		return doc, err
	}
	if s.getDB() == nil || !operatorTemplateIDPattern.MatchString(id) {
		return nil, configNotFound("template", id)
	}
	row, err := s.getOperatorTemplate(r.Context(), id)
	if err != nil {
		log.Printf("[config-view] operator template query: %v", err)
		if isOperatorTemplateTableMissing(err) {
			return nil, configNotFound("template", id)
		}
		return nil, configUnavailable("operator template query failed")
	}
	if row == nil {
		return nil, configNotFound("template", id)
	}
	var body map[string]any
	if err := json.Unmarshal(row.Document, &body); err != nil {
		return nil, err
	}
	return &configDocument{doc: body, source: configSourceOperator, sourcePath: "operator_templates/" + row.ID}, nil
}

// fileTemplateDocument finds a file bundle by id, or returns nil.
func (s *AdminServer) fileTemplateDocument(id string) (*configDocument, error) {
	root := s.templateBundlesPath()
	for _, dir := range []string{"", bootstrap.StartersDirName} {
		doc, rel, err := findTemplateFile(root, dir, id)
		if err != nil {
			return nil, err
		}
		if doc != nil {
			displayRoot := filepath.Base(filepath.Clean(root))
			return &configDocument{
				doc:        doc,
				source:     configSourceFile,
				sourcePath: filepath.ToSlash(filepath.Join(displayRoot, rel)),
			}, nil
		}
	}
	return nil, nil
}

// findTemplateFile scans one directory (non-recursive) for a YAML bundle
// whose `id` (or file stem, when id is absent) equals id. The id is only
// compared, never used to build a path. Unparseable files are skipped.
func findTemplateFile(root, dir, id string) (any, string, error) {
	path := filepath.Join(root, dir)
	entries, err := os.ReadDir(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", configUnavailable("template directory unreadable")
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if !e.IsDir() && (ext == ".yaml" || ext == ".yml") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(path, name))
		if err != nil {
			continue
		}
		var doc map[string]any
		if err := yaml.Unmarshal(raw, &doc); err != nil || doc == nil {
			continue
		}
		bundleID := strings.TrimSpace(fmt.Sprint(doc["id"]))
		if doc["id"] == nil {
			bundleID = strings.TrimSuffix(name, filepath.Ext(name))
		}
		if bundleID == id {
			return doc, filepath.Join(dir, name), nil
		}
	}
	return nil, "", nil
}

func configTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
