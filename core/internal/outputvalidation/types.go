package outputvalidation

import (
	"context"
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// Status describes the authoritative outcome of a browser validation attempt.
type Status string

const (
	StatusPassed      Status = "passed"
	StatusFailed      Status = "failed"
	StatusUnavailable Status = "unavailable"
	StatusError       Status = "error"
)

// The plan vocabulary aliases the governed protocol rather than defining a
// second adapter-owned contract.
type Check = protocol.OutputValidationCheck
type Action = protocol.OutputValidationActionKind
type Observation = protocol.OutputValidationObservationKind
type ProbeAction = protocol.OutputValidationAction
type ProbeObservation = protocol.OutputValidationObservation
type Probe = protocol.OutputValidationProbe
type Plan = protocol.OutputValidationPlan

const (
	KindInteractiveBrowser  = protocol.OutputValidationInteractiveBrowser
	CheckLoad               = protocol.OutputValidationCheckLoad
	CheckNoPageErrors       = protocol.OutputValidationCheckNoPageErrors
	CheckNoFailedLocalAsset = protocol.OutputValidationCheckNoFailedAssets
	ActionClick             = protocol.OutputValidationActionClick
	ActionKeyPress          = protocol.OutputValidationActionKeyPress
	ActionKeyHold           = protocol.OutputValidationActionKeyHold
	ActionFill              = protocol.OutputValidationActionFill
	ActionPointer           = protocol.OutputValidationActionPointer
	ObserveVisualChange     = protocol.OutputValidationObserveVisualChange
	ObserveTextChange       = protocol.OutputValidationObserveTextChange
	ObserveValueChange      = protocol.OutputValidationObserveValueChange
	ObserveElementVisible   = protocol.OutputValidationObserveElementVisible
	ObserveURLChange        = protocol.OutputValidationObserveURLChange
)

// Request binds validation to the exact launch target and retained content digest.
type Request struct {
	LaunchURL     string `json:"launch_url"`
	ContentDigest string `json:"content_digest"`
	EvidencePath  string `json:"evidence_path"`
	Plan          Plan   `json:"plan"`
}

// Diagnostic is a bounded, operator-readable validation finding.
type Diagnostic struct {
	Code     string `json:"code"`
	Message  string `json:"message"`
	Severity string `json:"severity"`
}

// EvidenceRef points only to evidence retained beneath the caller-supplied path.
type EvidenceRef struct {
	Kind        string `json:"kind"`
	Path        string `json:"path"`
	ContentType string `json:"content_type,omitempty"`
	SHA256      string `json:"sha256,omitempty"`
}

// CheckResult records one page-level check without exposing raw browser internals.
type CheckResult struct {
	Check  Check  `json:"check"`
	Passed bool   `json:"passed"`
	Detail string `json:"detail,omitempty"`
}

// ProbeResult records the requested action and whether its observation was satisfied.
type ProbeResult struct {
	Action      Action      `json:"action"`
	Observation Observation `json:"observation"`
	Passed      bool        `json:"passed"`
	Before      string      `json:"before,omitempty"`
	After       string      `json:"after,omitempty"`
}

// CriterionEvidence binds an explicit acceptance criterion to retained proof.
type CriterionEvidence struct {
	Criterion    string   `json:"criterion"`
	Passed       bool     `json:"passed"`
	EvidenceRefs []string `json:"evidence_refs,omitempty"`
}

// Report is the typed, digest-bound result returned by any output validator adapter.
type Report struct {
	Status            Status              `json:"status"`
	ContentDigest     string              `json:"content_digest"`
	LaunchURL         string              `json:"launch_url"`
	StartedAt         time.Time           `json:"started_at"`
	FinishedAt        time.Time           `json:"finished_at"`
	Diagnostics       []Diagnostic        `json:"diagnostics,omitempty"`
	EvidenceRefs      []EvidenceRef       `json:"evidence_refs,omitempty"`
	Checks            []CheckResult       `json:"checks,omitempty"`
	Probe             *ProbeResult        `json:"probe,omitempty"`
	CriterionEvidence []CriterionEvidence `json:"criterion_evidence,omitempty"`
}

// Validator allows the server to replace Playwright without changing projection semantics.
type Validator interface {
	Validate(ctx context.Context, request Request) (Report, error)
}
