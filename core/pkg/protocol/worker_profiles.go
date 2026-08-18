package protocol

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// WorkerProfileSpec is the declarative, user-owned definition used to create
// a bounded worker. Runtime identity and revision lineage come from the
// ConfigDocument envelope rather than being duplicated inside Spec.
type WorkerProfileSpec struct {
	Description          string                `json:"description,omitempty"`
	Role                 string                `json:"role"`
	SystemPrompt         string                `json:"system_prompt"`
	Model                string                `json:"model,omitempty"`
	CapabilityRefs       []string              `json:"capability_refs,omitempty"`
	ContextBindings      []AgentContextBinding `json:"context_bindings,omitempty"`
	UsagePolicy          AgentUsagePolicy      `json:"usage_policy,omitempty"`
	Inputs               []string              `json:"inputs,omitempty"`
	Outputs              []string              `json:"outputs,omitempty"`
	VerificationStrategy string                `json:"verification_strategy,omitempty"`
	VerificationRubric   []string              `json:"verification_rubric,omitempty"`
	ValidationCommand    string                `json:"validation_command,omitempty"`
}

type WorkerProfileSnapshot struct {
	ID      string `json:"id"`
	Version string `json:"version"`
	Digest  string `json:"digest"`
}

type WorkerProfileCompileResult struct {
	Profile  WorkerProfileSpec     `json:"profile"`
	Snapshot WorkerProfileSnapshot `json:"profile_snapshot"`
	Ready    bool                  `json:"ready"`
}

// DecodeWorkerProfileSpec rejects unknown fields so Soma-authored content and
// direct YAML/JSON files share the same family contract.
func DecodeWorkerProfileSpec(raw json.RawMessage) (WorkerProfileSpec, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var profile WorkerProfileSpec
	if err := decoder.Decode(&profile); err != nil {
		return WorkerProfileSpec{}, fmt.Errorf("decode worker profile: %w", err)
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		if err == nil {
			err = fmt.Errorf("multiple JSON values")
		}
		return WorkerProfileSpec{}, fmt.Errorf("decode worker profile: %w", err)
	}
	return profile, nil
}

// ValidateWorkerProfileSpec returns deterministic field-level issues for one
// profile family without mutating or normalizing the supplied payload.
func ValidateWorkerProfileSpec(raw json.RawMessage) []ConfigDocumentValidationIssue {
	profile, err := DecodeWorkerProfileSpec(raw)
	if err != nil {
		return []ConfigDocumentValidationIssue{{
			Code: "worker_profile.invalid_spec", Field: "spec", Message: err.Error(),
		}}
	}

	issues := make([]ConfigDocumentValidationIssue, 0)
	add := func(code, field, message string) {
		issues = append(issues, ConfigDocumentValidationIssue{Code: code, Field: field, Message: message})
	}
	if strings.TrimSpace(profile.Role) == "" {
		add("worker_profile.missing_role", "spec.role", "worker profile role is required")
	}
	if strings.TrimSpace(profile.SystemPrompt) == "" {
		add("worker_profile.missing_system_prompt", "spec.system_prompt", "worker profile instructions are required")
	}
	validateWorkerProfileStringList(profile.CapabilityRefs, "capability_refs", true, add)
	validateWorkerProfileStringList(profile.Inputs, "inputs", true, add)
	validateWorkerProfileStringList(profile.Outputs, "outputs", true, add)
	validateWorkerProfileContext(profile.ContextBindings, add)
	validateWorkerProfileUsage(profile.UsagePolicy, add)
	validateWorkerProfileVerification(profile, add)
	return issues
}

func validateWorkerProfileStringList(values []string, name string, unique bool, add func(string, string, string)) {
	seen := make(map[string]struct{}, len(values))
	for index, value := range values {
		field := fmt.Sprintf("spec.%s[%d]", name, index)
		trimmed := strings.TrimSpace(value)
		if trimmed == "" || trimmed != value {
			add("worker_profile.invalid_"+name, field, name+" entries must be non-empty and have no surrounding whitespace")
			continue
		}
		if unique {
			if _, exists := seen[value]; exists {
				add("worker_profile.duplicate_"+name, field, name+" entries must be unique")
			}
			seen[value] = struct{}{}
		}
	}
}

func validateWorkerProfileContext(bindings []AgentContextBinding, add func(string, string, string)) {
	seen := make(map[string]struct{}, len(bindings))
	for index, binding := range bindings {
		prefix := fmt.Sprintf("spec.context_bindings[%d]", index)
		if strings.TrimSpace(binding.Kind) == "" || binding.Kind != strings.TrimSpace(binding.Kind) {
			add("worker_profile.invalid_context_kind", prefix+".kind", "context kind is required and must not have surrounding whitespace")
		}
		if binding.Ref != strings.TrimSpace(binding.Ref) || binding.Access != strings.TrimSpace(binding.Access) {
			add("worker_profile.invalid_context_binding", prefix, "context ref and access must not have surrounding whitespace")
		}
		key := binding.Kind + "\x00" + binding.Ref + "\x00" + binding.Access
		if _, exists := seen[key]; exists {
			add("worker_profile.duplicate_context_binding", prefix, "context bindings must be unique")
		}
		seen[key] = struct{}{}
	}
}

func validateWorkerProfileUsage(usage AgentUsagePolicy, add func(string, string, string)) {
	selection := strings.TrimSpace(usage.Selection)
	if selection != usage.Selection {
		add("worker_profile.invalid_selection", "spec.usage_policy.selection", "selection must not have surrounding whitespace")
	}
	if selection != "" {
		switch selection {
		case "soma_or_manual", "suggested", "soma", "manual", "automatic":
		default:
			add("worker_profile.unsupported_selection", "spec.usage_policy.selection", fmt.Sprintf("unsupported selection %q", selection))
		}
	}
	scope := strings.TrimSpace(usage.Scope)
	if scope != usage.Scope {
		add("worker_profile.invalid_usage_scope", "spec.usage_policy.scope", "usage scope must not have surrounding whitespace")
	}
	if scope != "" {
		switch scope {
		case "workspace", "outcome", "organization", "operator":
		default:
			add("worker_profile.unsupported_usage_scope", "spec.usage_policy.scope", fmt.Sprintf("unsupported usage scope %q", scope))
		}
	}
}

func validateWorkerProfileVerification(profile WorkerProfileSpec, add func(string, string, string)) {
	strategy := strings.TrimSpace(profile.VerificationStrategy)
	if strategy != profile.VerificationStrategy {
		add("worker_profile.invalid_verification_strategy", "spec.verification_strategy", "verification strategy must not have surrounding whitespace")
	}
	if strategy != "" && strategy != string(VerifySemantic) && strategy != string(VerifyEmpirical) {
		add("worker_profile.unsupported_verification_strategy", "spec.verification_strategy", fmt.Sprintf("unsupported verification strategy %q", strategy))
	}
	validateWorkerProfileStringList(profile.VerificationRubric, "verification_rubric", true, add)
	if strategy == string(VerifySemantic) && len(profile.VerificationRubric) == 0 {
		add("worker_profile.missing_verification_rubric", "spec.verification_rubric", "semantic verification requires at least one rubric item")
	}
	if strategy == string(VerifyEmpirical) && strings.TrimSpace(profile.ValidationCommand) == "" {
		add("worker_profile.missing_validation_command", "spec.validation_command", "empirical verification requires a validation command")
	}
	if strategy != string(VerifyEmpirical) && strings.TrimSpace(profile.ValidationCommand) != "" {
		add("worker_profile.unexpected_validation_command", "spec.validation_command", "validation command requires empirical verification")
	}
}
