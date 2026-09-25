package protocol

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
	DeliveryPackage *DeliveryPackageSummary `json:"delivery_package,omitempty"`
}
