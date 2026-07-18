package inputs

import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	ErrInvalidInput = errors.New("invalid input source")
	ErrNotFound     = errors.New("input source not found")
	ErrUnavailable  = errors.New("input source store unavailable")

	sourceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{2,80}$`)
	secretRefRe     = regexp.MustCompile(`^(env:[A-Z0-9_]+|secret:[A-Za-z0-9_.:/-]+|vault:[A-Za-z0-9_.:/-]+|[A-Z][A-Z0-9_]{2,})$`)
)

type SourceInput struct {
	ID                    string          `json:"id,omitempty"`
	Name                  string          `json:"name"`
	SourceType            string          `json:"source_type"`
	AdapterKind           string          `json:"adapter_kind"`
	ScopeKind             string          `json:"scope_kind,omitempty"`
	ScopeRef              string          `json:"scope_ref,omitempty"`
	TargetOutcomeID       string          `json:"target_outcome_id,omitempty"`
	TargetGroupID         string          `json:"target_group_id,omitempty"`
	TargetHostID          string          `json:"target_host_id,omitempty"`
	AuthScheme            string          `json:"auth_scheme,omitempty"`
	SecretRef             string          `json:"secret_ref,omitempty"`
	AllowedIngressSubject string          `json:"allowed_ingress_subject,omitempty"`
	PayloadSchemaRef      string          `json:"payload_schema_ref,omitempty"`
	BufferMode            string          `json:"buffer_mode,omitempty"`
	BufferPolicy          json.RawMessage `json:"buffer_policy,omitempty"`
	SensitivityClass      string          `json:"sensitivity_class,omitempty"`
	TrustClass            string          `json:"trust_class,omitempty"`
	Status                string          `json:"status,omitempty"`
	Recovery              string          `json:"recovery,omitempty"`
}

func NormalizeSourceInput(in SourceInput) (Source, error) {
	id := normalizeID(in.ID)
	name := strings.TrimSpace(in.Name)
	sourceType := normalizeToken(in.SourceType, "api")
	adapterKind := normalizeAdapter(in.AdapterKind)
	if id == "" {
		id = normalizeID(name)
	}
	if !sourceIDPattern.MatchString(id) {
		return Source{}, fmt.Errorf("%w: id must be 3-81 lowercase characters, numbers, dashes, or underscores", ErrInvalidInput)
	}
	if name == "" {
		return Source{}, fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	scopeKind, scopeRef, err := normalizeScope(in.ScopeKind, in.ScopeRef)
	if err != nil {
		return Source{}, err
	}
	authScheme := normalizeAuth(in.AuthScheme)
	secretRef := strings.TrimSpace(in.SecretRef)
	if authScheme == AuthNone {
		secretRef = ""
	} else if !safeSecretRef(secretRef) {
		return Source{}, fmt.Errorf("%w: secret_ref must reference a deployment secret, not a raw credential", ErrInvalidInput)
	}
	bufferMode := normalizeBufferMode(in.BufferMode)
	bufferPolicy := normalizeJSON(in.BufferPolicy)
	subject := strings.TrimSpace(in.AllowedIngressSubject)
	if subject == "" {
		subject = "swarm.global.input." + id
	}
	if !safeIngressSubject(subject) {
		return Source{}, fmt.Errorf("%w: allowed_ingress_subject must be a concrete swarm.global.input.* subject", ErrInvalidInput)
	}
	return Source{
		ID:                    id,
		Name:                  name,
		SourceType:            sourceType,
		AdapterKind:           adapterKind,
		ScopeKind:             scopeKind,
		ScopeRef:              scopeRef,
		TargetOutcomeID:       strings.TrimSpace(in.TargetOutcomeID),
		TargetGroupID:         strings.TrimSpace(in.TargetGroupID),
		TargetHostID:          strings.TrimSpace(in.TargetHostID),
		AuthScheme:            authScheme,
		SecretRef:             secretRef,
		AllowedIngressSubject: subject,
		PayloadSchemaRef:      strings.TrimSpace(in.PayloadSchemaRef),
		BufferMode:            bufferMode,
		BufferPolicy:          bufferPolicy,
		SensitivityClass:      normalizeToken(in.SensitivityClass, "governed"),
		TrustClass:            normalizeToken(in.TrustClass, "bounded_external"),
		Status:                normalizeStatus(in.Status),
		Recovery:              strings.TrimSpace(in.Recovery),
		TenantID:              "default",
	}, nil
}

func normalizeID(value string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	v = strings.NewReplacer(" ", "-", ".", "-", "/", "-", "\\", "-", ":", "-").Replace(v)
	v = regexp.MustCompile(`[^a-z0-9_-]+`).ReplaceAllString(v, "")
	return strings.Trim(v, "-_")
}

func normalizeToken(value, fallback string) string {
	v := strings.ToLower(strings.TrimSpace(value))
	if v == "" {
		return fallback
	}
	return strings.NewReplacer(" ", "_", "-", "_").Replace(v)
}

func normalizeAdapter(value string) string {
	switch normalizeToken(value, AdapterAPI) {
	case AdapterAPI, AdapterWebhook, AdapterMCP, AdapterDevice, AdapterSensor, AdapterDatabase, AdapterFile:
		return normalizeToken(value, AdapterAPI)
	default:
		return AdapterAPI
	}
}

func normalizeAuth(value string) string {
	switch normalizeToken(value, AuthNone) {
	case AuthSecretRef, AuthBearerToken, AuthAPIKey, AuthBasic:
		return normalizeToken(value, AuthNone)
	default:
		return AuthNone
	}
}

func normalizeBufferMode(value string) string {
	switch normalizeToken(value, BufferAppendLog) {
	case BufferLatestState, BufferAppendLatest, BufferWindowedRollup:
		return normalizeToken(value, BufferAppendLog)
	default:
		return BufferAppendLog
	}
}

func normalizeStatus(value string) string {
	switch normalizeToken(value, StatusAvailable) {
	case StatusPaused, StatusError:
		return normalizeToken(value, StatusAvailable)
	default:
		return StatusAvailable
	}
}

func normalizeScope(kind, ref string) (string, string, error) {
	k := normalizeToken(kind, ScopeAll)
	r := strings.TrimSpace(ref)
	switch k {
	case ScopeAll:
		return ScopeAll, "", nil
	case ScopeGroup, ScopeHost:
		if r == "" {
			return "", "", fmt.Errorf("%w: scope_ref is required for %s scope", ErrInvalidInput, k)
		}
		return k, r, nil
	default:
		return "", "", fmt.Errorf("%w: unsupported scope_kind %q", ErrInvalidInput, kind)
	}
}

func normalizeJSON(raw json.RawMessage) json.RawMessage {
	if len(raw) == 0 || strings.TrimSpace(string(raw)) == "" {
		return json.RawMessage(`{}`)
	}
	return raw
}

func safeSecretRef(value string) bool {
	value = strings.TrimSpace(value)
	return value != "" && secretRefRe.MatchString(value)
}

func safeIngressSubject(subject string) bool {
	if strings.ContainsAny(subject, " *>") {
		return false
	}
	return strings.HasPrefix(subject, "swarm.global.input.") && subject != "swarm.global.input."
}

func ErrorStatus(err error) int {
	switch {
	case errors.Is(err, ErrInvalidInput):
		return 400
	case errors.Is(err, ErrNotFound):
		return 404
	case errors.Is(err, ErrUnavailable):
		return 503
	default:
		return 500
	}
}
