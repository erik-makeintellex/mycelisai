package protocol

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
)

const ConfigDocumentAPIVersionV1 = "mycelis.ai/v1"

type ConfigDocumentKind string
type ConfigDocumentScopeKind string
type ConfigDocumentSourceKind string
type ConfigDocumentRiskLevel string
type ConfigDocumentEffectAction string

const (
	ConfigDocumentKindOutcomeTemplate   ConfigDocumentKind = "OutcomeTemplate"
	ConfigDocumentKindWorkerProfile     ConfigDocumentKind = "WorkerProfile"
	ConfigDocumentKindCodeContextSource ConfigDocumentKind = "CodeContextSource"

	ConfigDocumentScopeBuiltIn      ConfigDocumentScopeKind = "built_in"
	ConfigDocumentScopeOrganization ConfigDocumentScopeKind = "organization"
	ConfigDocumentScopeWorkspace    ConfigDocumentScopeKind = "workspace"
	ConfigDocumentScopeOperator     ConfigDocumentScopeKind = "operator"

	ConfigDocumentSourceBuiltIn ConfigDocumentSourceKind = "built_in"
	ConfigDocumentSourceFile    ConfigDocumentSourceKind = "file"
	ConfigDocumentSourceSoma    ConfigDocumentSourceKind = "soma"
	ConfigDocumentSourceAPI     ConfigDocumentSourceKind = "api"

	ConfigDocumentRiskLow    ConfigDocumentRiskLevel = "low"
	ConfigDocumentRiskMedium ConfigDocumentRiskLevel = "medium"
	ConfigDocumentRiskHigh   ConfigDocumentRiskLevel = "high"

	ConfigDocumentEffectNone       ConfigDocumentEffectAction = "none"
	ConfigDocumentEffectActivate   ConfigDocumentEffectAction = "activate"
	ConfigDocumentEffectDeactivate ConfigDocumentEffectAction = "deactivate"
)

var (
	configDocumentIDPattern      = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{2,80}$`)
	configDocumentVersionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,63}$`)
	configDocumentEnvRefPattern  = regexp.MustCompile(`^[A-Z_][A-Z0-9_]*$`)
	configDocumentPathRefPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.:/-]*$`)
)

// ConfigDocument is the file-authoritative v1 envelope shared by direct and
// Soma-authored configuration. Family-specific adapters own the shape of Spec.
type ConfigDocument struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       ConfigDocumentKind     `json:"kind"`
	Metadata   ConfigDocumentMetadata `json:"metadata"`
	Spec       json.RawMessage        `json:"spec"`
}

type ConfigDocumentMetadata struct {
	ID         string                   `json:"id"`
	Name       string                   `json:"name"`
	Version    string                   `json:"version"`
	OwnerID    string                   `json:"owner_id"`
	Scope      ConfigDocumentScope      `json:"scope"`
	Enabled    bool                     `json:"enabled"`
	Source     ConfigDocumentSource     `json:"source"`
	SecretRefs []string                 `json:"secret_refs,omitempty"`
	Governance ConfigDocumentGovernance `json:"governance"`
}

type ConfigDocumentScope struct {
	Kind ConfigDocumentScopeKind `json:"kind"`
	Ref  string                  `json:"ref,omitempty"`
}

type ConfigDocumentSource struct {
	Kind ConfigDocumentSourceKind `json:"kind"`
	Ref  string                   `json:"ref"`
}

type ConfigDocumentGovernance struct {
	RiskLevel       ConfigDocumentRiskLevel `json:"risk_level"`
	ApprovalPosture ApprovalPosture         `json:"approval_posture"`
}

type ConfigDocumentValidationIssue struct {
	Code    string `json:"code"`
	Field   string `json:"field"`
	Message string `json:"message"`
}

type ConfigDocumentDryRunResult struct {
	Valid  bool                            `json:"valid"`
	Digest string                          `json:"digest,omitempty"`
	Effect ConfigDocumentDryRunEffect      `json:"effect"`
	Issues []ConfigDocumentValidationIssue `json:"issues"`
}

type ConfigDocumentDryRunEffect struct {
	Action          ConfigDocumentEffectAction `json:"action"`
	DocumentID      string                     `json:"document_id"`
	Version         string                     `json:"version"`
	Scope           ConfigDocumentScope        `json:"scope"`
	Enabled         bool                       `json:"enabled"`
	RiskLevel       ConfigDocumentRiskLevel    `json:"risk_level"`
	ApprovalPosture ApprovalPosture            `json:"approval_posture"`
}

// ValidateConfigDocument returns all deterministic, structured validation
// issues. It never normalizes or mutates the supplied document.
func ValidateConfigDocument(document ConfigDocument) []ConfigDocumentValidationIssue {
	issues := make([]ConfigDocumentValidationIssue, 0)
	add := func(code, field, message string) {
		issues = append(issues, ConfigDocumentValidationIssue{Code: code, Field: field, Message: message})
	}

	if document.APIVersion != ConfigDocumentAPIVersionV1 {
		add("config.unsupported_api_version", "apiVersion", fmt.Sprintf("unsupported apiVersion %q", document.APIVersion))
	}
	if document.Kind != ConfigDocumentKindOutcomeTemplate && document.Kind != ConfigDocumentKindWorkerProfile && document.Kind != ConfigDocumentKindCodeContextSource {
		add("config.unsupported_kind", "kind", fmt.Sprintf("unsupported config document kind %q", document.Kind))
	}

	metadata := document.Metadata
	if metadata.ID == "" {
		add("metadata.missing_id", "metadata.id", "config document id is required")
	} else if !configDocumentIDPattern.MatchString(metadata.ID) {
		add("metadata.invalid_id", "metadata.id", "config document id must be a stable lowercase identifier")
	}
	if strings.TrimSpace(metadata.Name) == "" {
		add("metadata.missing_name", "metadata.name", "config document name is required")
	} else if metadata.Name != strings.TrimSpace(metadata.Name) {
		add("metadata.invalid_name", "metadata.name", "config document name must not have surrounding whitespace")
	}
	if metadata.Version == "" {
		add("metadata.missing_version", "metadata.version", "config document version is required")
	} else if !configDocumentVersionPattern.MatchString(metadata.Version) {
		add("metadata.invalid_version", "metadata.version", "config document version must be a stable version token")
	}
	if strings.TrimSpace(metadata.OwnerID) == "" {
		add("metadata.missing_owner", "metadata.owner_id", "config document owner is required")
	} else if metadata.OwnerID != strings.TrimSpace(metadata.OwnerID) {
		add("metadata.invalid_owner", "metadata.owner_id", "config document owner must not have surrounding whitespace")
	}

	validateConfigDocumentScope(metadata.Scope, add)
	validateConfigDocumentSource(metadata.Source, add)
	validateConfigDocumentSecretRefs(metadata.SecretRefs, add)
	validateConfigDocumentGovernance(metadata.Governance, add)

	spec, err := decodeConfigDocumentSpec(document.Spec)
	if err != nil {
		add("spec.invalid_json", "spec", err.Error())
	} else if len(spec) == 0 {
		add("spec.empty", "spec", "config document spec must contain at least one field")
	} else {
		validateConfigDocumentSpecSecrets(spec, "spec", &issues)
		if document.Kind == ConfigDocumentKindWorkerProfile {
			issues = append(issues, ValidateWorkerProfileSpec(document.Spec)...)
		}
		if document.Kind == ConfigDocumentKindCodeContextSource {
			issues = append(issues, ValidateCodeContextSourceSpec(document.Spec)...)
		}
	}

	return issues
}

// CanonicalConfigDocumentDigest returns a digest over the complete validated
// envelope after canonicalizing Spec object keys.
func CanonicalConfigDocumentDigest(document ConfigDocument) (string, error) {
	issues := ValidateConfigDocument(document)
	if len(issues) != 0 {
		return "", fmt.Errorf("invalid config document: %s", issues[0].Code)
	}

	spec, err := decodeConfigDocumentSpec(document.Spec)
	if err != nil {
		return "", fmt.Errorf("decode config document spec: %w", err)
	}
	canonicalSpec, err := json.Marshal(spec)
	if err != nil {
		return "", fmt.Errorf("marshal canonical config document spec: %w", err)
	}
	document.Spec = canonicalSpec
	canonicalDocument, err := json.Marshal(document)
	if err != nil {
		return "", fmt.Errorf("marshal canonical config document: %w", err)
	}

	sum := sha256.Sum256(canonicalDocument)
	return "sha256:" + hex.EncodeToString(sum[:]), nil
}

// DryRunConfigDocument validates and describes the requested activation effect.
// It is a pure protocol projection and has no store or activation side effects.
func DryRunConfigDocument(document ConfigDocument) ConfigDocumentDryRunResult {
	issues := ValidateConfigDocument(document)
	effect := ConfigDocumentDryRunEffect{
		Action:          ConfigDocumentEffectNone,
		DocumentID:      document.Metadata.ID,
		Version:         document.Metadata.Version,
		Scope:           document.Metadata.Scope,
		Enabled:         document.Metadata.Enabled,
		RiskLevel:       document.Metadata.Governance.RiskLevel,
		ApprovalPosture: document.Metadata.Governance.ApprovalPosture,
	}
	result := ConfigDocumentDryRunResult{Valid: len(issues) == 0, Effect: effect, Issues: issues}
	if !result.Valid {
		return result
	}

	digest, err := CanonicalConfigDocumentDigest(document)
	if err != nil {
		result.Valid = false
		result.Issues = append(result.Issues, ConfigDocumentValidationIssue{
			Code: "digest.failed", Field: "spec", Message: err.Error(),
		})
		return result
	}
	result.Digest = digest
	if document.Metadata.Enabled {
		result.Effect.Action = ConfigDocumentEffectActivate
	} else {
		result.Effect.Action = ConfigDocumentEffectDeactivate
	}
	return result
}

func validateConfigDocumentScope(scope ConfigDocumentScope, add func(string, string, string)) {
	switch scope.Kind {
	case ConfigDocumentScopeBuiltIn:
		if scope.Ref != "" {
			add("metadata.invalid_scope_ref", "metadata.scope.ref", "built-in scope must not include a ref")
		}
	case ConfigDocumentScopeOrganization, ConfigDocumentScopeWorkspace, ConfigDocumentScopeOperator:
		if strings.TrimSpace(scope.Ref) == "" {
			add("metadata.missing_scope_ref", "metadata.scope.ref", fmt.Sprintf("%s scope requires a ref", scope.Kind))
		} else if scope.Ref != strings.TrimSpace(scope.Ref) {
			add("metadata.invalid_scope_ref", "metadata.scope.ref", "scope ref must not have surrounding whitespace")
		}
	default:
		add("metadata.unsupported_scope", "metadata.scope.kind", fmt.Sprintf("unsupported config document scope %q", scope.Kind))
	}
}

func validateConfigDocumentSource(source ConfigDocumentSource, add func(string, string, string)) {
	switch source.Kind {
	case ConfigDocumentSourceBuiltIn, ConfigDocumentSourceFile, ConfigDocumentSourceSoma, ConfigDocumentSourceAPI:
	default:
		add("metadata.unsupported_source", "metadata.source.kind", fmt.Sprintf("unsupported config document source %q", source.Kind))
	}
	if strings.TrimSpace(source.Ref) == "" {
		add("metadata.missing_source_ref", "metadata.source.ref", "config document source provenance ref is required")
	} else if source.Ref != strings.TrimSpace(source.Ref) {
		add("metadata.invalid_source_ref", "metadata.source.ref", "source provenance ref must not have surrounding whitespace")
	}
}

func validateConfigDocumentSecretRefs(refs []string, add func(string, string, string)) {
	seen := make(map[string]struct{}, len(refs))
	for index, ref := range refs {
		field := fmt.Sprintf("metadata.secret_refs[%d]", index)
		if !isConfigDocumentSecretRef(ref) {
			add("metadata.invalid_secret_ref", field, "secret refs must name a managed secret, not contain a raw credential")
			continue
		}
		if _, exists := seen[ref]; exists {
			add("metadata.duplicate_secret_ref", field, "secret refs must be unique")
			continue
		}
		seen[ref] = struct{}{}
	}
}

func validateConfigDocumentGovernance(governance ConfigDocumentGovernance, add func(string, string, string)) {
	switch governance.RiskLevel {
	case ConfigDocumentRiskLow, ConfigDocumentRiskMedium, ConfigDocumentRiskHigh:
	default:
		add("metadata.unsupported_risk_level", "metadata.governance.risk_level", fmt.Sprintf("unsupported risk level %q", governance.RiskLevel))
	}
	switch governance.ApprovalPosture {
	case ApprovalPostureAutoAllowed, ApprovalPostureOptional, ApprovalPostureRequired:
	default:
		add("metadata.unsupported_approval_posture", "metadata.governance.approval_posture", fmt.Sprintf("unsupported approval posture %q", governance.ApprovalPosture))
	}
}
