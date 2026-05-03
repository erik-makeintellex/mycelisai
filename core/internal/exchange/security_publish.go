package exchange

import (
	"fmt"
	"strings"
)

func buildAuditMetadata(input PublishInput, capability CapabilityDefinition, channel *Channel) map[string]any {
	metadata := map[string]any{}
	for k, v := range input.Metadata {
		metadata[k] = v
	}
	metadata["security"] = map[string]any{
		"sensitivity_class": input.SensitivityClass,
		"allowed_consumers": input.AllowedConsumers,
		"trust_class":       input.TrustClass,
		"review_required":   input.ReviewRequired,
		"visibility":        input.Visibility,
	}
	metadata["audit"] = map[string]any{
		"published_by":    input.CreatedBy,
		"source_role":     input.SourceRole,
		"source_team":     input.SourceTeam,
		"target_role":     input.TargetRole,
		"target_team":     input.TargetTeam,
		"channel_name":    channel.Name,
		"thread_id":       input.ThreadID,
		"capability_id":   input.CapabilityID,
		"trust_class":     input.TrustClass,
		"audit_required":  capability.AuditRequired,
		"approval_future": capability.ApprovalRequired,
		"review_required": input.ReviewRequired,
	}
	return metadata
}

func enrichPublishInput(input PublishInput, channel *Channel) (PublishInput, CapabilityDefinition, error) {
	input = enrichPublishRoles(input)
	input = enrichPublishSensitivityAndCapability(input, channel)
	capability, ok := CapabilityByID(input.CapabilityID)
	if !ok {
		return input, CapabilityDefinition{}, fmt.Errorf("exchange capability %s is not registered", input.CapabilityID)
	}
	input = enrichPublishPolicy(input, channel, capability)
	input.Metadata = buildAuditMetadata(input, capability, channel)
	return input, capability, nil
}

func enrichPublishRoles(input PublishInput) PublishInput {
	if strings.TrimSpace(input.SourceRole) == "" {
		input.SourceRole = normalizeRole(input.CreatedBy)
	}
	if strings.TrimSpace(input.TargetRole) == "" {
		if role, _ := input.Payload["target_role"].(string); strings.TrimSpace(role) != "" {
			input.TargetRole = role
		} else {
			input.TargetRole = input.AddressedTo
		}
	}
	input.SourceRole = normalizeRole(input.SourceRole)
	input.TargetRole = normalizeRole(input.TargetRole)
	return input
}

func enrichPublishSensitivityAndCapability(input PublishInput, channel *Channel) PublishInput {
	if strings.TrimSpace(input.SensitivityClass) == "" {
		if raw, _ := input.Payload["sensitivity_class"].(string); strings.TrimSpace(raw) != "" {
			input.SensitivityClass = raw
		} else {
			input.SensitivityClass = defaultSensitivityForChannel(channel)
		}
	}
	if strings.TrimSpace(input.CapabilityID) == "" {
		if raw, _ := input.Payload["capability_id"].(string); strings.TrimSpace(raw) != "" {
			input.CapabilityID = raw
		} else {
			input.CapabilityID = defaultCapabilityForSchema(input.SchemaID)
		}
	}
	return input
}

func enrichPublishPolicy(input PublishInput, channel *Channel, capability CapabilityDefinition) PublishInput {
	if strings.TrimSpace(input.TrustClass) == "" {
		if raw, _ := input.Payload["trust_class"].(string); strings.TrimSpace(raw) != "" {
			input.TrustClass = raw
		} else {
			input.TrustClass = classifyTrustBoundary(input.SourceRole, capability)
		}
	}
	if len(input.AllowedConsumers) == 0 {
		input.AllowedConsumers = defaultAllowedConsumers(channel, input.TargetRole, capability)
	}
	if !input.ReviewRequired {
		if raw, ok := input.Payload["review_required"].(bool); ok {
			input.ReviewRequired = raw
		} else {
			input.ReviewRequired = capability.AuditRequired || capability.RiskClass != "low-risk" || input.TrustClass != "trusted_internal"
		}
	}
	if strings.TrimSpace(input.Visibility) == "" {
		input.Visibility = channel.Visibility
	}
	return syncPublishPayload(input)
}

func syncPublishPayload(input PublishInput) PublishInput {
	input.Payload["source_role"] = input.SourceRole
	input.Payload["target_role"] = input.TargetRole
	input.Payload["sensitivity_class"] = input.SensitivityClass
	input.Payload["capability_id"] = input.CapabilityID
	input.Payload["trust_class"] = input.TrustClass
	input.Payload["review_required"] = input.ReviewRequired
	input.Payload["allowed_consumers"] = toAnySlice(input.AllowedConsumers)
	return input
}
