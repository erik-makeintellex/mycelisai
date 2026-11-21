package server

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

type packageOutputFile struct {
	Path        string                        `json:"path"`
	Title       string                        `json:"title,omitempty"`
	ContentType string                        `json:"content_type,omitempty"`
	Bytes       int                           `json:"bytes"`
	SHA256      string                        `json:"sha256"`
	ArtifactID  string                        `json:"artifact_id,omitempty"`
	Proof       *protocol.OutputProofEnvelope `json:"proof,omitempty"`
}

type packageOutputSummary struct {
	PackageKind string                     `json:"package_kind"`
	PackageSlug string                     `json:"package_slug"`
	PackageID   string                     `json:"package_id"`
	TeamID      string                     `json:"team_id,omitempty"`
	TeamName    string                     `json:"team_name,omitempty"`
	BasePath    string                     `json:"base_path"`
	ProofPath   string                     `json:"proof_path"`
	RunID       string                     `json:"run_id"`
	ProofID     string                     `json:"proof_id"`
	Files       []packageOutputFile        `json:"files"`
	Artifacts   []protocol.ChatArtifactRef `json:"artifacts,omitempty"`
	OutputRefs  []protocol.OutputRef       `json:"output_refs,omitempty"`
	PrimaryRef  *protocol.OutputRef        `json:"primary_output_ref,omitempty"`
	ProofRef    *protocol.OutputRef        `json:"proof_ref,omitempty"`
}

func plannedTeamIdentity(planned []protocol.PlannedToolCall) (string, string) {
	teamID, teamName := "", ""
	for _, call := range planned {
		if strings.TrimSpace(call.Name) == "create_team" {
			cid, cname := firstStringArgument(call.Arguments, "team_id"), firstStringArgument(call.Arguments, "name")
			if teamID == "" || !strings.HasSuffix(cid, "-review-team") {
				teamID, teamName = cid, cname
			}
			continue
		}
		if teamID == "" {
			teamID, teamName = firstStringArgument(call.Arguments, "team_id"), firstStringArgument(call.Arguments, "team_name")
		}
	}
	return teamID, teamName
}

func (s *AdminServer) finalizeCodexTaskBundleTx(ctx context.Context, tx *sql.Tx, planned []protocol.PlannedToolCall, runID, proofID string) (*packageOutputSummary, error) {
	basePath, slug, packageKind, ok := deliverablePackageMetadataFromCalls(planned)
	if !ok {
		return nil, nil
	}
	files := make([]packageOutputFile, 0, len(planned))
	artifacts := make([]protocol.ChatArtifactRef, 0, len(planned)+1)
	for _, call := range planned {
		if strings.TrimSpace(call.Name) != "write_file" {
			continue
		}
		file, artifact, err := packageFileFromPlannedCall(call)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
		artifacts = append(artifacts, artifact)
	}
	proofPath := filepath.ToSlash(path.Join(basePath, "proof.md"))
	proofContent := buildDeliverablePackageProof(packageKind, runID, proofID, basePath, files)
	proofFile, proofArtifact, err := writeCodexPackageProof(proofPath, proofContent)
	if err != nil {
		return nil, err
	}
	files = append(files, proofFile)
	artifacts = append(artifacts, proofArtifact)
	teamID, teamName := plannedTeamIdentity(planned)
	summary := &packageOutputSummary{
		PackageKind: packageKind, PackageSlug: slug, PackageID: packageKind + ":" + slug,
		TeamID: teamID, TeamName: teamName, BasePath: basePath, ProofPath: proofPath,
		RunID: runID, ProofID: proofID, Files: files, Artifacts: artifacts,
	}
	if err := s.persistCodexPackageArtifactsTx(ctx, tx, summary); err != nil {
		return nil, err
	}
	summary.OutputRefs = outputRefsFromPackageSummary(summary)
	summary.PrimaryRef, summary.ProofRef = outputRefPointers(summary.OutputRefs)
	if err := s.insertPackageArtifactEventTx(tx, summary); err != nil {
		return nil, err
	}
	return summary, nil
}

func packageFileFromPlannedCall(call protocol.PlannedToolCall) (packageOutputFile, protocol.ChatArtifactRef, error) {
	relativePath := firstStringArgument(call.Arguments, "path")
	absPath, err := workspacePathForServer(relativePath)
	if err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, err
	}
	data, err := os.ReadFile(absPath)
	if err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, fmt.Errorf("read package artifact %s: %w", relativePath, err)
	}
	title := firstNonEmptyString(call.Arguments["title"], call.Arguments["artifact_role"])
	if title == "" {
		title = filepath.Base(relativePath)
	}
	file := packageOutputFileFromBytes(relativePath, title, data)
	return file, protocol.ChatArtifactRef{
		Type: packageArtifactTypeForPath(relativePath), Title: title,
		ContentType: file.ContentType, Content: string(data), SavedPath: filepath.ToSlash(relativePath),
	}, nil
}

func writeCodexPackageProof(proofPath, proofContent string) (packageOutputFile, protocol.ChatArtifactRef, error) {
	proofAbsPath, err := workspacePathForServer(proofPath)
	if err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, err
	}
	if err := os.MkdirAll(filepath.Dir(proofAbsPath), 0o755); err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, fmt.Errorf("create package proof directory: %w", err)
	}
	if err := os.WriteFile(proofAbsPath, []byte(proofContent), 0o644); err != nil {
		return packageOutputFile{}, protocol.ChatArtifactRef{}, fmt.Errorf("write package proof: %w", err)
	}
	return packageOutputFileFromBytes(proofPath, "Package Proof", []byte(proofContent)), protocol.ChatArtifactRef{
		Type: "document", Title: "Package Proof",
		ContentType: "text/markdown", Content: proofContent, SavedPath: proofPath,
	}, nil
}

func packageContentTypeForPath(relativePath string) string {
	lower := strings.ToLower(strings.TrimSpace(relativePath))
	switch {
	case strings.HasSuffix(lower, ".html"): return "text/html"
	case strings.HasSuffix(lower, ".json"): return "application/json"
	case strings.HasSuffix(lower, ".yaml"), strings.HasSuffix(lower, ".yml"): return "text/yaml"
	default: return "text/markdown"
	}
}
func packageArtifactTypeForPath(relativePath string) string {
	if strings.HasSuffix(strings.ToLower(strings.TrimSpace(relativePath)), ".html") { return "code" }
	return "document"
}
func packageOutputFileFromBytes(relativePath, title string, data []byte) packageOutputFile {
	sum := sha256.Sum256(data)
	return packageOutputFile{Path: filepath.ToSlash(relativePath), Title: title, ContentType: packageContentTypeForPath(relativePath), Bytes: len(data), SHA256: fmt.Sprintf("%x", sum)}
}
func isDeliverablePackageKind(kind string) bool { return isSupportedDeliverablePackageKind(kind) }
func isSupportedDeliverablePackageKind(kind string) bool {
	switch strings.TrimSpace(kind) {
	case "codex_task_bundle", "marketing_launch_bundle", "interactive_game_bundle", "customer_onboarding_bundle", "kpi_dashboard_bundle", "product_wireframe_bundle", "mission_blueprint_bundle", "chained_pipeline_bundle", "team_execution_bundle", "dev_project_bundle", "multi_app_bundle", "mcp_addition_bundle":
		return true
	default: return false
	}
}

func codexPackageBaseFromCalls(planned []protocol.PlannedToolCall) (string, string, bool) {
	basePath, slug, _, ok := deliverablePackageMetadataFromCalls(planned)
	return basePath, slug, ok
}
func deliverablePackageMetadataFromCalls(planned []protocol.PlannedToolCall) (string, string, string, bool) {
	if len(planned) < 2 {
		return "", "", "", false
	}
	basePath, slug, packageKind := "", "", ""
	for _, call := range planned {
		if strings.TrimSpace(call.Name) != "write_file" {
			continue
		}
		kind := firstStringArgument(call.Arguments, "package_kind")
		if !isSupportedDeliverablePackageKind(kind) {
			return "", "", "", false
		}
		relativePath := filepath.ToSlash(firstStringArgument(call.Arguments, "path"))
		dir := path.Dir(relativePath)
		if dir == "." || dir == "" {
			return "", "", "", false
		}
		if basePath == "" {
			basePath = dir
			slug = firstStringArgument(call.Arguments, "package_slug")
			packageKind = kind
			continue
		}
		if dir != basePath || kind != packageKind {
			return "", "", "", false
		}
	}
	if basePath == "" {
		return "", "", "", false
	}
	if packageKind == "" {
		packageKind = "codex_task_bundle"
	}
	if slug == "" {
		slug = path.Base(basePath)
	}
	return basePath, slug, packageKind, true
}

func workspacePathForServer(rawPath string) (string, error) {
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	absWorkspace, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("invalid workspace path: %w", err)
	}
	normalized := strings.TrimSpace(strings.ReplaceAll(rawPath, "\\", "/"))
	normalized = path.Clean(normalized)
	normalized = strings.TrimPrefix(normalized, "./")
	if strings.HasPrefix(normalized, "workspace/") {
		normalized = strings.TrimPrefix(normalized, "workspace/")
	}
	if normalized == "." || normalized == "" {
		return "", fmt.Errorf("workspace file path is required")
	}
	target := filepath.Clean(filepath.Join(absWorkspace, filepath.FromSlash(normalized)))
	rel, err := filepath.Rel(absWorkspace, target)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("path %q escapes workspace boundary", rawPath)
	}
	return target, nil
}

func deliverablePackageProofTitle(packageKind string) string {
	switch strings.TrimSpace(packageKind) {
	case "marketing_launch_bundle": return "Marketing Launch Bundle Proof"
	case "interactive_game_bundle": return "Interactive Game Bundle Proof"
	case "customer_onboarding_bundle": return "Customer Onboarding Bundle Proof"
	case "kpi_dashboard_bundle": return "KPI Dashboard Bundle Proof"
	case "product_wireframe_bundle": return "Product Wireframe Bundle Proof"
	case "mission_blueprint_bundle": return "Mission Blueprint Bundle Proof"
	case "chained_pipeline_bundle": return "Chained Team Pipeline Proof"
	case "team_execution_bundle": return "Team Execution Plan & Next Steps Proof"
	case "dev_project_bundle": return "Development Project Bundle Proof"
	case "multi_app_bundle": return "Interacting Multi-App Suite Proof"
	case "mcp_addition_bundle": return "Deployment MCP Addition & Binding Proof"
	default: return "Codex Task Bundle Proof"
	}
}
func buildCodexPackageProof(runID, proofID, basePath string, files []packageOutputFile) string {
	return buildDeliverablePackageProof("codex_task_bundle", runID, proofID, basePath, files)
}

func buildDeliverablePackageProof(packageKind, runID, proofID, basePath string, files []packageOutputFile) string {
	var b strings.Builder
	b.WriteString("# " + deliverablePackageProofTitle(packageKind) + "\n\n")
	b.WriteString("- Package kind: `" + packageKind + "`\n")
	b.WriteString("- Base path: `" + basePath + "`\n")
	b.WriteString("- Run ID: `" + runID + "`\n")
	b.WriteString("- Intent proof ID: `" + proofID + "`\n\n")
	b.WriteString("## Files\n\n")
	for _, file := range files {
		b.WriteString("- `" + file.Path + "`")
		b.WriteString(fmt.Sprintf(" (%d bytes, sha256 `%s`)\n", file.Bytes, file.SHA256))
	}
	return b.String()
}

func (s *AdminServer) insertPackageArtifactEventTx(tx *sql.Tx, summary *packageOutputSummary) error {
	if tx == nil || summary == nil {
		return nil
	}
	payload, _ := json.Marshal(map[string]any{
		"package_kind":       summary.PackageKind,
		"package_slug":       summary.PackageSlug,
		"package_id":         summary.PackageID,
		"base_path":          summary.BasePath,
		"proof_path":         summary.ProofPath,
		"proof_id":           summary.ProofID,
		"files":              summary.Files,
		"proof_envelopes":    proofEnvelopesFromPackageSummary(summary),
		"output_refs":        summary.OutputRefs,
		"primary_output_ref": summary.PrimaryRef,
		"proof_ref":          summary.ProofRef,
	})
	_, err := tx.Exec(
		`INSERT INTO mission_events
			(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		uuid.New().String(), summary.RunID, "default", string(protocol.EventArtifactCreated), string(protocol.SeverityInfo),
		"admin", "governance", payload, time.Now(),
	)
	return err
}
