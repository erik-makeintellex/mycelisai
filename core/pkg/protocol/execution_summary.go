package protocol

import "strings"

// OutputRef is the shared normalization boundary for retained outputs.
// It projects files, packages, team results, MCP results, and future execution
// outputs into one inspectable shape without forcing one storage model.
type OutputRef struct {
	ID               string               `json:"id,omitempty"`
	ArtifactID       string               `json:"artifact_id,omitempty"`
	Kind             string               `json:"kind"`
	Label            string               `json:"label"`
	StorageRef       string               `json:"storage_ref,omitempty"`
	OpenURL          string               `json:"open_url,omitempty"`
	RevealURL        string               `json:"reveal_url,omitempty"`
	SourceRunID      string               `json:"source_run_id,omitempty"`
	SourceContractID string               `json:"source_contract_id,omitempty"`
	ProofArtifactID  string               `json:"proof_artifact_id,omitempty"`
	PackageID        string               `json:"package_id,omitempty"`
	Role             string               `json:"role,omitempty"`
	Status           string               `json:"status,omitempty"`
	ContentType      string               `json:"content_type,omitempty"`
	Bytes            int                  `json:"bytes,omitempty"`
	SHA256           string               `json:"sha256,omitempty"`
	Proof            *OutputProofEnvelope `json:"proof,omitempty"`
	ProofEnvelope    *OutputProofEnvelope `json:"proof_envelope,omitempty"`
}

// OutputProofEnvelope is the minimum proof projection attached to retained outputs.
type OutputProofEnvelope struct {
	ProofID            string `json:"proof_id,omitempty"`
	OutputRefID        string `json:"output_ref_id,omitempty"`
	ArtifactID         string `json:"artifact_id,omitempty"`
	StorageRef         string `json:"storage_ref,omitempty"`
	SourceRunID        string `json:"source_run_id,omitempty"`
	SourceContractID   string `json:"source_contract_id,omitempty"`
	ExecutionStatus    string `json:"execution_status,omitempty"`
	PathBoundaryStatus string `json:"path_boundary_status,omitempty"`
	Checksum           string `json:"checksum,omitempty"`
	ChecksumAlgorithm  string `json:"checksum_algorithm,omitempty"`
	Bytes              int    `json:"bytes,omitempty"`
	ContentType        string `json:"content_type,omitempty"`
	CreatedAt          string `json:"created_at,omitempty"`
	ReadbackStatus     string `json:"readback_status,omitempty"`
	RecoveryHint       string `json:"recovery_hint,omitempty"`
}

// ExecutionOutput is a compatibility output descriptor retained alongside
// normalized OutputRef entries during MYC-RC-002 convergence.
type ExecutionOutput struct {
	ID               string               `json:"id,omitempty"`
	ArtifactID       string               `json:"artifact_id,omitempty"`
	ProofArtifactID  string               `json:"proof_artifact_id,omitempty"`
	Kind             string               `json:"kind"`
	Title            string               `json:"title"`
	Summary          string               `json:"summary,omitempty"`
	StorageRef       string               `json:"storage_ref,omitempty"`
	Href             string               `json:"href,omitempty"`
	OpenURL          string               `json:"open_url,omitempty"`
	RevealURL        string               `json:"reveal_url,omitempty"`
	SourceRunID      string               `json:"source_run_id,omitempty"`
	SourceContractID string               `json:"source_contract_id,omitempty"`
	PackageID        string               `json:"package_id,omitempty"`
	Role             string               `json:"role,omitempty"`
	Status           string               `json:"status,omitempty"`
	ContentType      string               `json:"content_type,omitempty"`
	Bytes            int                  `json:"bytes,omitempty"`
	SHA256           string               `json:"sha256,omitempty"`
	Retained         bool                 `json:"retained,omitempty"`
	Proof            *OutputProofEnvelope `json:"proof,omitempty"`
}

// TeamOutputRef is a team/group-scoped output reference retained alongside
// normalized OutputRef entries during MYC-RC-002 convergence.
type TeamOutputRef struct {
	OutputID      string               `json:"output_id,omitempty"`
	ArtifactID    string               `json:"artifact_id,omitempty"`
	TeamID        string               `json:"team_id,omitempty"`
	WorkItemID    string               `json:"work_item_id,omitempty"`
	RunID         string               `json:"run_id,omitempty"`
	Kind          string               `json:"kind"`
	Label         string               `json:"label"`
	StorageRef    string               `json:"storage_ref,omitempty"`
	OpenURL       string               `json:"open_url,omitempty"`
	RevealURL     string               `json:"reveal_url,omitempty"`
	ValidationRef string               `json:"validation_ref,omitempty"`
	ProofRef      string               `json:"proof_ref,omitempty"`
	ContractID    string               `json:"contract_id,omitempty"`
	ProofID       string               `json:"proof_id,omitempty"`
	PackageID     string               `json:"package_id,omitempty"`
	Role          string               `json:"role,omitempty"`
	Status        string               `json:"status,omitempty"`
	ContentType   string               `json:"content_type,omitempty"`
	Bytes         int                  `json:"bytes,omitempty"`
	SHA256        string               `json:"sha256,omitempty"`
	Proof         *OutputProofEnvelope `json:"proof,omitempty"`
}

// DeliveryPackageSummary is the user-facing grouping contract for a mission result.
type DeliveryPackageSummary struct {
	PackageID        string      `json:"package_id,omitempty"`
	RunID            string      `json:"run_id,omitempty"`
	Title            string      `json:"title"`
	Summary          string      `json:"summary,omitempty"`
	PackageKind      string      `json:"package_kind"`
	Status           string      `json:"status"`
	PrimaryOutputRef *OutputRef  `json:"primary_output_ref,omitempty"`
	Items            []OutputRef `json:"items,omitempty"`
	NextActions      []string    `json:"next_actions,omitempty"`
	ProofRef         *OutputRef  `json:"proof_ref,omitempty"`
	InspectRefs      []OutputRef `json:"inspect_refs,omitempty"`
}

// ExecutionSummary is the response-level summary returned by governed execution.
type ExecutionSummary struct {
	Title           string                  `json:"title"`
	Summary         string                  `json:"summary,omitempty"`
	Status          string                  `json:"status"`
	RunID           string                  `json:"run_id,omitempty"`
	ProofID         string                  `json:"proof_id,omitempty"`
	OutputRefs      []OutputRef             `json:"output_refs,omitempty"`
	Outputs         []ExecutionOutput       `json:"outputs,omitempty"`
	ArtifactRefs    []ChatArtifactRef       `json:"artifact_refs,omitempty"`
	TeamOutputRefs  []TeamOutputRef         `json:"team_output_refs,omitempty"`
	DeliveryPackage *DeliveryPackageSummary `json:"delivery_package,omitempty"`
}

// NormalizeOutputRef ensures an OutputRef carries canonical fallback defaults
// and synchronizes Proof / ProofEnvelope pointers when proof metadata exists.
func NormalizeOutputRef(ref OutputRef) OutputRef {
	env := ref.Proof
	if env == nil {
		env = ref.ProofEnvelope
	}
	if env != nil {
		ref.Proof = env
		ref.ProofEnvelope = env
		ref.ID = firstNonEmptyValue(ref.ID, env.OutputRefID, env.ArtifactID)
		ref.ArtifactID = firstNonEmptyValue(ref.ArtifactID, env.ArtifactID)
		ref.StorageRef = firstNonEmptyValue(ref.StorageRef, env.StorageRef)
		ref.SourceRunID = firstNonEmptyValue(ref.SourceRunID, env.SourceRunID)
		ref.SourceContractID = firstNonEmptyValue(ref.SourceContractID, env.SourceContractID)
		ref.ProofArtifactID = firstNonEmptyValue(ref.ProofArtifactID, env.ProofID, env.ArtifactID)
		ref.ContentType = firstNonEmptyValue(ref.ContentType, env.ContentType)
		if ref.Bytes == 0 && env.Bytes > 0 {
			ref.Bytes = env.Bytes
		}
		ref.SHA256 = firstNonEmptyValue(ref.SHA256, env.Checksum)
	}
	ref.ID = firstNonEmptyValue(ref.ID, ref.ArtifactID, ref.StorageRef)
	ref.Kind = firstNonEmptyValue(ref.Kind, "file")
	ref.Label = firstNonEmptyValue(ref.Label, ref.StorageRef, ref.ID, "Output")
	ref.Role = firstNonEmptyValue(ref.Role, "primary")
	ref.Status = firstNonEmptyValue(ref.Status, "recorded")
	return ref
}

// OutputRefFromExecutionOutput projects an ExecutionOutput into a normalized OutputRef.
func OutputRefFromExecutionOutput(out ExecutionOutput) OutputRef {
	status := strings.TrimSpace(out.Status)
	if status == "" {
		if out.Retained {
			status = "recorded"
		} else {
			status = "transient"
		}
	}
	return NormalizeOutputRef(OutputRef{
		ID:               firstNonEmptyValue(out.ID, out.ArtifactID, out.StorageRef, out.Title),
		ArtifactID:       out.ArtifactID,
		Kind:             firstNonEmptyValue(out.Kind, "file"),
		Label:            firstNonEmptyValue(out.Title, out.ID, out.StorageRef, "Output"),
		StorageRef:       firstNonEmptyValue(out.StorageRef, out.Href),
		OpenURL:          firstNonEmptyValue(out.OpenURL, out.Href),
		RevealURL:        out.RevealURL,
		SourceRunID:      out.SourceRunID,
		SourceContractID: out.SourceContractID,
		ProofArtifactID:  out.ProofArtifactID,
		PackageID:        out.PackageID,
		Role:             firstNonEmptyValue(out.Role, "primary"),
		Status:           status,
		ContentType:      out.ContentType,
		Bytes:            out.Bytes,
		SHA256:           out.SHA256,
		Proof:            out.Proof,
		ProofEnvelope:    out.Proof,
	})
}

// OutputRefFromTeamOutputRef projects a TeamOutputRef into a normalized OutputRef.
func OutputRefFromTeamOutputRef(ref TeamOutputRef) OutputRef {
	return NormalizeOutputRef(OutputRef{
		ID:               firstNonEmptyValue(ref.OutputID, ref.ArtifactID, ref.StorageRef, ref.Label),
		ArtifactID:       ref.ArtifactID,
		Kind:             firstNonEmptyValue(ref.Kind, "file"),
		Label:            firstNonEmptyValue(ref.Label, ref.OutputID, ref.StorageRef, "Team output"),
		StorageRef:       ref.StorageRef,
		OpenURL:          ref.OpenURL,
		RevealURL:        ref.RevealURL,
		SourceRunID:      ref.RunID,
		SourceContractID: ref.ContractID,
		ProofArtifactID:  firstNonEmptyValue(ref.ProofRef, ref.ProofID),
		PackageID:        firstNonEmptyValue(ref.PackageID, ref.TeamID),
		Role:             firstNonEmptyValue(ref.Role, "primary"),
		Status:           firstNonEmptyValue(ref.Status, "recorded"),
		ContentType:      ref.ContentType,
		Bytes:            ref.Bytes,
		SHA256:           ref.SHA256,
		Proof:            ref.Proof,
		ProofEnvelope:    ref.Proof,
	})
}
