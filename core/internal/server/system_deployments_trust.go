package server

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const unknownTrustValue = "unknown"

type DeploymentTrustSnapshot struct {
	DeploymentRoot  string                  `json:"deployment_root"`
	ExecutionRoot   string                  `json:"execution_root"`
	WorkspaceRoot   string                  `json:"workspace_root"`
	ArtifactRoot    string                  `json:"artifact_root"`
	CurrentCommit   string                  `json:"current_commit"`
	ImageTag        string                  `json:"image_tag"`
	ChartVersion    string                  `json:"chart_version"`
	DeploymentLane  string                  `json:"deployment_lane"`
	EndpointPosture string                  `json:"endpoint_posture"`
	RuntimeHealth   DeploymentRuntimeHealth `json:"runtime_health"`
	ProofLane       string                  `json:"proof_lane"`
	RecoveryPosture string                  `json:"recovery_posture"`
	CheckedAt       time.Time               `json:"checked_at"`
	Sources         []DeploymentTrustSource `json:"sources"`
}

type DeploymentRuntimeHealth struct {
	Status        string          `json:"status"`
	Online        int             `json:"online"`
	Degraded      int             `json:"degraded"`
	Offline       int             `json:"offline"`
	Total         int             `json:"total"`
	ServiceStatus []ServiceStatus `json:"service_status"`
}

type DeploymentTrustSource struct {
	Field  string `json:"field"`
	Source string `json:"source"`
}

// HandleDeploymentTrust returns the deploy/runtime trust snapshot shown in
// System -> Deployments. It only reports non-secret posture values.
func (s *AdminServer) HandleDeploymentTrust(w http.ResponseWriter, r *http.Request) {
	services := s.buildServiceStatuses(r.Context())
	contract := ResolveDeploymentContract()
	repoRoot := discoverDeploymentRoot()

	snapshot := DeploymentTrustSnapshot{
		DeploymentRoot:  valueOrUnknown(envFirst("MYCELIS_DEPLOYMENT_ROOT"), repoRoot),
		ExecutionRoot:   valueOrUnknown(mustGetwd()),
		WorkspaceRoot:   valueOrUnknown(envFirst("MYCELIS_BACKEND_WORKSPACE_ROOT", "MYCELIS_WORKSPACE")),
		ArtifactRoot:    valueOrUnknown(envFirst("MYCELIS_ARTIFACT_ROOT", "MYCELIS_ARTIFACTS_ROOT")),
		CurrentCommit:   valueOrUnknown(envFirst("MYCELIS_CURRENT_COMMIT", "GIT_COMMIT", "SOURCE_COMMIT"), readGitCommit(repoRoot)),
		ImageTag:        valueOrUnknown(envFirst("MYCELIS_IMAGE_TAG", "IMAGE_TAG")),
		ChartVersion:    valueOrUnknown(envFirst("MYCELIS_CHART_VERSION"), readChartVersion(repoRoot)),
		DeploymentLane:  valueOrUnknown(envFirst("MYCELIS_DEPLOYMENT_LANE")),
		EndpointPosture: deploymentEndpointPosture(contract),
		RuntimeHealth:   summarizeRuntimeHealth(services),
		ProofLane:       valueOrUnknown(envFirst("MYCELIS_PROOF_LANE")),
		RecoveryPosture: deploymentRecoveryPosture(contract),
		CheckedAt:       time.Now().UTC(),
	}
	snapshot.Sources = deploymentTrustSources(snapshot)

	respondJSON(w, map[string]any{"ok": true, "data": snapshot})
}

func summarizeRuntimeHealth(services []ServiceStatus) DeploymentRuntimeHealth {
	summary := DeploymentRuntimeHealth{
		Status:        unknownTrustValue,
		Total:         len(services),
		ServiceStatus: services,
	}
	for _, svc := range services {
		switch svc.Status {
		case "online":
			summary.Online++
		case "degraded":
			summary.Degraded++
		case "offline":
			summary.Offline++
		}
	}
	switch {
	case summary.Total == 0:
		summary.Status = unknownTrustValue
	case summary.Offline > 0:
		summary.Status = "offline"
	case summary.Degraded > 0:
		summary.Status = "degraded"
	default:
		summary.Status = "online"
	}
	return summary
}

func deploymentEndpointPosture(contract DeploymentContract) string {
	if contract.IdentityMode == "" || contract.ProductEdition == "" {
		return unknownTrustValue
	}
	return contract.ProductEdition + "/" + contract.IdentityMode
}

func deploymentRecoveryPosture(contract DeploymentContract) string {
	if contract.RequiresBreakGlassRecovery() {
		return "break_glass_required"
	}
	return "local_owner_recovery"
}

func deploymentTrustSources(snapshot DeploymentTrustSnapshot) []DeploymentTrustSource {
	fields := []struct {
		field string
		value string
	}{
		{"deployment_root", snapshot.DeploymentRoot},
		{"execution_root", snapshot.ExecutionRoot},
		{"workspace_root", snapshot.WorkspaceRoot},
		{"artifact_root", snapshot.ArtifactRoot},
		{"current_commit", snapshot.CurrentCommit},
		{"image_tag", snapshot.ImageTag},
		{"chart_version", snapshot.ChartVersion},
		{"deployment_lane", snapshot.DeploymentLane},
		{"endpoint_posture", snapshot.EndpointPosture},
		{"proof_lane", snapshot.ProofLane},
		{"recovery_posture", snapshot.RecoveryPosture},
	}
	sources := make([]DeploymentTrustSource, 0, len(fields))
	for _, field := range fields {
		source := "runtime"
		if field.value == unknownTrustValue {
			source = "unavailable"
		}
		sources = append(sources, DeploymentTrustSource{Field: field.field, Source: source})
	}
	return sources
}

func valueOrUnknown(values ...string) string {
	for _, value := range values {
		if trimmed := strings.TrimSpace(value); trimmed != "" {
			return trimmed
		}
	}
	return unknownTrustValue
}

func envFirst(keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return value
		}
	}
	return ""
}

func mustGetwd() string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return wd
}

func discoverDeploymentRoot() string {
	wd := mustGetwd()
	for wd != "" {
		if _, err := os.Stat(filepath.Join(wd, ".git")); err == nil {
			return wd
		}
		if _, err := os.Stat(filepath.Join(wd, "charts", "mycelis-core", "Chart.yaml")); err == nil {
			return wd
		}
		parent := filepath.Dir(wd)
		if parent == wd {
			break
		}
		wd = parent
	}
	return ""
}

func readGitCommit(root string) string {
	if root == "" {
		return ""
	}
	headPath := filepath.Join(root, ".git", "HEAD")
	head, err := os.ReadFile(headPath)
	if err != nil {
		return ""
	}
	ref := strings.TrimSpace(string(head))
	if strings.HasPrefix(ref, "ref: ") {
		refPath := filepath.Join(root, ".git", filepath.FromSlash(strings.TrimPrefix(ref, "ref: ")))
		data, err := os.ReadFile(refPath)
		if err != nil {
			return ""
		}
		return strings.TrimSpace(string(data))
	}
	return ref
}

func readChartVersion(root string) string {
	if root == "" {
		return ""
	}
	file, err := os.Open(filepath.Join(root, "charts", "mycelis-core", "Chart.yaml"))
	if err != nil {
		return ""
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "version:") {
			return strings.Trim(strings.TrimSpace(strings.TrimPrefix(line, "version:")), `"`)
		}
	}
	return ""
}
