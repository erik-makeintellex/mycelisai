package exchange

import "testing"

func TestCapabilityByIDResolvesRiskClass(t *testing.T) {
	capability, ok := CapabilityByID("api_data_access")
	if !ok {
		t.Fatal("expected api_data_access capability to exist")
	}
	if capability.RiskClass != "high-risk" {
		t.Fatalf("expected high-risk capability, got %s", capability.RiskClass)
	}
	if !capability.AuditRequired {
		t.Fatal("expected api_data_access to require audit")
	}
}

func TestChannelPermissionsRespectReadersWritersAndReviewers(t *testing.T) {
	channel := SeedChannels[0]

	if !canReadChannel(Actor{Role: "soma"}, &channel) {
		t.Fatal("expected soma to read planning channel")
	}
	if canWriteChannel(Actor{Role: "specialist"}, &channel) {
		t.Fatal("did not expect specialist to write planning channel")
	}
	if !canReviewChannel(Actor{Role: "review"}, &channel) {
		t.Fatal("expected review role to review planning channel")
	}
}

func TestThreadPermissionsRespectParticipantsAndReviewers(t *testing.T) {
	channel := SeedChannels[1]
	thread := Thread{
		Participants:     []string{"soma", "team_lead"},
		AllowedReviewers: []string{"review"},
		EscalationRights: []string{"soma", "review"},
	}

	if !canAccessThread(Actor{Role: "team_lead"}, &channel, &thread) {
		t.Fatal("expected participant to access thread")
	}
	if !canPublishToThread(Actor{Role: "review"}, &channel, &thread) {
		t.Fatal("expected allowed reviewer to publish into thread")
	}
	if canPublishToThread(Actor{Role: "specialist"}, &channel, &thread) {
		t.Fatal("did not expect specialist to publish into restricted thread")
	}
}

func TestItemVisibilityHonorsSensitivityAndAllowedConsumers(t *testing.T) {
	channel := SeedChannels[2]
	item := ExchangeItem{
		SensitivityClass: "role_scoped",
		SourceRole:       "automation",
		TargetRole:       "soma",
		AllowedConsumers: []string{"soma", "team_lead", "review"},
	}

	if !canReadItem(Actor{Role: "soma"}, &channel, &item) {
		t.Fatal("expected soma to consume allowed item")
	}
	if canReadItem(Actor{Role: "specialist"}, &channel, &item) {
		t.Fatal("did not expect specialist to consume restricted item")
	}

	item.SensitivityClass = "admin_only"
	if canReadItem(Actor{Role: "soma"}, &channel, &item) {
		t.Fatal("did not expect non-admin to consume admin_only item")
	}
}

func TestTrustClassificationDoesNotTreatExternalAsInternal(t *testing.T) {
	mcpCapability, _ := CapabilityByID("browser_research")
	if trust := classifyTrustBoundary("mcp", mcpCapability); trust != "bounded_external" {
		t.Fatalf("expected bounded_external trust, got %s", trust)
	}

	nodeCapability, _ := CapabilityByID("remote_node_execution")
	if trust := classifyTrustBoundary("remote_node", nodeCapability); trust != "restricted" {
		t.Fatalf("expected restricted trust, got %s", trust)
	}
}

func TestEnrichPublishInputAddsAuditSecurityMetadata(t *testing.T) {
	channel := SeedChannels[3]
	input := PublishInput{
		SchemaID:     "ToolResult",
		ChannelName:  channel.Name,
		CreatedBy:    "mcp:fetch",
		AddressedTo:  "soma",
		Payload: map[string]any{
			"summary":     "Fetched competitive notes.",
			"status":      "completed",
			"source_role": "mcp",
			"target_role": "soma",
			"created_at":  "2026-03-24T10:00:00Z",
		},
	}

	enriched, capability, err := enrichPublishInput(input, &channel)
	if err != nil {
		t.Fatalf("enrichPublishInput error = %v", err)
	}
	if capability.ID == "" {
		t.Fatal("expected capability to resolve")
	}
	if enriched.TrustClass != "bounded_external" {
		t.Fatalf("expected bounded_external trust, got %s", enriched.TrustClass)
	}
	if !enriched.ReviewRequired {
		t.Fatal("expected review_required for bounded external tool output")
	}
	security, ok := enriched.Metadata["security"].(map[string]any)
	if !ok {
		t.Fatal("expected security metadata block")
	}
	if security["sensitivity_class"] != "team_scoped" {
		t.Fatalf("expected team_scoped sensitivity, got %#v", security["sensitivity_class"])
	}
	audit, ok := enriched.Metadata["audit"].(map[string]any)
	if !ok {
		t.Fatal("expected audit metadata block")
	}
	if audit["capability_id"] != capability.ID {
		t.Fatalf("expected audit capability %s, got %#v", capability.ID, audit["capability_id"])
	}
}
