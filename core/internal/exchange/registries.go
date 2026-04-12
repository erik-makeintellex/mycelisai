package exchange

var SeedFields = []FieldDefinition{
	{Name: "summary", Type: "string", SemanticMeaning: "Operator-readable summary of the exchange output.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "status", Type: "enum", SemanticMeaning: "Lifecycle state of the work or output.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "priority", Type: "enum", SemanticMeaning: "Relative urgency or handling priority.", Indexed: true, Visibility: "soma", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "artifact_uri", Type: "reference", SemanticMeaning: "Stable pointer to a stored artifact or generated output.", Indexed: true, Visibility: "soma", UsageContexts: []string{"artifact", "message"}},
	{Name: "source_role", Type: "string", SemanticMeaning: "Role or system that produced the exchange item.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message"}},
	{Name: "source_team", Type: "string", SemanticMeaning: "Team scope that produced the exchange item when applicable.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message"}},
	{Name: "target_role", Type: "string", SemanticMeaning: "Role, team, or system expected to consume the exchange item.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "target_team", Type: "string", SemanticMeaning: "Team scope explicitly targeted for consumption or review.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "confidence", Type: "number", SemanticMeaning: "Confidence signal attached to an output or review.", Indexed: true, Visibility: "soma", UsageContexts: []string{"artifact", "message"}},
	{Name: "memory_layer", Type: "enum", SemanticMeaning: "Target memory layer for a learning or promotion candidate.", Indexed: true, Visibility: "admin", UsageContexts: []string{"message", "thread"}},
	{Name: "classification", Type: "enum", SemanticMeaning: "Explicit classification assigned before any candidate is promoted into memory.", Indexed: true, Visibility: "admin", UsageContexts: []string{"message", "thread"}},
	{Name: "classification_reason", Type: "string", SemanticMeaning: "Short rationale explaining why the candidate received its classification.", Indexed: true, Visibility: "admin", UsageContexts: []string{"message", "thread"}},
	{Name: "promotion_target", Type: "enum", SemanticMeaning: "Specific store or memory class the candidate may be promoted into after review.", Indexed: true, Visibility: "admin", UsageContexts: []string{"message", "thread"}},
	{Name: "evidence_refs", Type: "array", SemanticMeaning: "Pointers to source exchange items, artifacts, runs, or events supporting a learning candidate.", Indexed: true, Visibility: "admin", UsageContexts: []string{"message", "thread"}},
	{Name: "tags", Type: "array", SemanticMeaning: "Loose classification labels for discovery and filtering.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "continuity_key", Type: "string", SemanticMeaning: "Stable continuity hint linking related work over time.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "sensitivity_class", Type: "enum", SemanticMeaning: "Scope boundary for who may view or act on this exchange object.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "capability_id", Type: "string", SemanticMeaning: "Capability identifier used to create or normalize the output.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message"}},
	{Name: "trust_class", Type: "enum", SemanticMeaning: "Trust boundary classification for the producing source or capability.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message"}},
	{Name: "review_required", Type: "boolean", SemanticMeaning: "Whether higher-level review is required before relying on the output.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message"}},
	{Name: "allowed_consumers", Type: "array", SemanticMeaning: "Roles allowed to consume the exchange object downstream.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message"}},
	{Name: "created_at", Type: "string", SemanticMeaning: "Creation timestamp for ordering and retention.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "updated_at", Type: "string", SemanticMeaning: "Most recent update timestamp for the exchange object.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message", "thread"}},
}

var SeedSchemas = []SchemaDefinition{
	{ID: "TextResult", Label: "Text Result", Description: "Plain language output intended for review or handoff.", RequiredFields: []string{"summary", "status", "source_role", "target_role", "created_at"}, OptionalFields: []string{"confidence", "tags", "continuity_key", "updated_at", "sensitivity_class", "capability_id", "trust_class", "review_required", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"text_output"}},
	{ID: "PlanResult", Label: "Plan Result", Description: "Structured planning output for work design and coordination.", RequiredFields: []string{"summary", "status", "priority", "source_role", "target_role", "created_at"}, OptionalFields: []string{"tags", "continuity_key", "updated_at", "sensitivity_class", "capability_id", "trust_class", "review_required", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"planning"}},
	{ID: "ReviewResult", Label: "Review Result", Description: "Review, critique, or governance feedback on work already produced.", RequiredFields: []string{"summary", "status", "confidence", "source_role", "target_role", "created_at"}, OptionalFields: []string{"tags", "continuity_key", "updated_at", "sensitivity_class", "capability_id", "trust_class", "review_required", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"review"}},
	{ID: "MediaResult", Label: "Media Result", Description: "Media output such as generated imagery or referenced audio/video assets.", RequiredFields: []string{"summary", "status", "artifact_uri", "source_role", "target_role", "created_at"}, OptionalFields: []string{"confidence", "tags", "continuity_key", "updated_at", "sensitivity_class", "capability_id", "trust_class", "review_required", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"media_output"}},
	{ID: "FileResult", Label: "File Result", Description: "File-oriented output intended for later delivery or inspection.", RequiredFields: []string{"summary", "status", "artifact_uri", "source_role", "target_role", "created_at"}, OptionalFields: []string{"tags", "continuity_key", "updated_at", "sensitivity_class", "capability_id", "trust_class", "review_required", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"file_output"}},
	{ID: "ToolResult", Label: "Tool Result", Description: "Normalized output from MCP or service/tool invocations.", RequiredFields: []string{"summary", "status", "source_role", "target_role", "created_at"}, OptionalFields: []string{"artifact_uri", "confidence", "tags", "continuity_key", "updated_at", "sensitivity_class", "capability_id", "trust_class", "review_required", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"tool_execution"}},
	{ID: "LearningCandidate", Label: "Learning Candidate", Description: "Classified candidate learning signal surfaced for future synthesis or promotion, not direct memory mutation.", RequiredFields: []string{"summary", "status", "classification", "memory_layer", "confidence", "review_required", "tags", "continuity_key", "created_at"}, OptionalFields: []string{"classification_reason", "promotion_target", "evidence_refs", "source_role", "target_role", "updated_at", "sensitivity_class", "capability_id", "trust_class", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"learning"}},
	{ID: "Escalation", Label: "Escalation", Description: "Escalation or blocker state requiring higher-level review.", RequiredFields: []string{"summary", "status", "priority", "source_role", "target_role", "created_at"}, OptionalFields: []string{"confidence", "tags", "continuity_key", "updated_at", "sensitivity_class", "capability_id", "trust_class", "review_required", "allowed_consumers", "source_team", "target_team"}, RequiredCapabilities: []string{"escalation"}},
}

var SeedCapabilities = []CapabilityDefinition{
	{ID: "text_output", Label: "Text Output", Source: "internal_tool", RiskClass: "low-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "specialist", "review", "automation"}, AuditRequired: false, ApprovalRequired: false, Description: "Internal text generation and summarization."},
	{ID: "planning", Label: "Planning", Source: "internal_tool", RiskClass: "low-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "automation"}, AuditRequired: false, ApprovalRequired: false, Description: "Planning and structured coordination outputs."},
	{ID: "review", Label: "Review", Source: "internal_tool", RiskClass: "medium-risk", DefaultAllowedRoles: []string{"soma", "review", "team_lead"}, AuditRequired: true, ApprovalRequired: false, Description: "Review and critique capabilities that can affect governed decisions."},
	{ID: "media_output", Label: "Media Output", Source: "mcp", RiskClass: "medium-risk", DefaultAllowedRoles: []string{"soma", "creative", "team_lead", "mcp"}, AuditRequired: true, ApprovalRequired: false, Description: "Generated or imported media outputs."},
	{ID: "file_output", Label: "File Output", Source: "internal_tool", RiskClass: "medium-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "specialist"}, AuditRequired: true, ApprovalRequired: false, Description: "File-oriented outputs prepared for downstream delivery."},
	{ID: "tool_execution", Label: "Tool Execution", Source: "mcp", RiskClass: "high-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "mcp"}, AuditRequired: true, ApprovalRequired: false, Description: "Normalized MCP or tool execution outputs."},
	{ID: "learning", Label: "Learning", Source: "internal_tool", RiskClass: "medium-risk", DefaultAllowedRoles: []string{"soma", "automation", "team_lead"}, AuditRequired: true, ApprovalRequired: false, Description: "Learning signals surfaced into governed exchange."},
	{ID: "escalation", Label: "Escalation", Source: "internal_tool", RiskClass: "medium-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "review"}, AuditRequired: true, ApprovalRequired: false, Description: "Escalation outputs requiring broader review."},
	{ID: "browser_research", Label: "Browser Research", Source: "mcp", RiskClass: "medium-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "mcp"}, AuditRequired: true, ApprovalRequired: false, Description: "Bounded external browser or research results."},
	{ID: "media_generation", Label: "Media Generation", Source: "mcp", RiskClass: "medium-risk", DefaultAllowedRoles: []string{"soma", "creative", "team_lead", "mcp"}, AuditRequired: true, ApprovalRequired: false, Description: "Generated media outputs from external or MCP-backed systems."},
	{ID: "api_data_access", Label: "API Data Access", Source: "api", RiskClass: "high-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "mcp"}, AuditRequired: true, ApprovalRequired: false, Description: "Structured data retrieved from external APIs or providers."},
	{ID: "remote_node_execution", Label: "Remote Node Execution", Source: "node", RiskClass: "high-risk", DefaultAllowedRoles: []string{"soma", "team_lead", "automation"}, AuditRequired: true, ApprovalRequired: true, Description: "Future remote execution-node capabilities requiring tighter trust review."},
}

var SeedChannels = []Channel{
	{Name: "organization.planning.work", Type: "planning", Owner: "system", SchemaID: "PlanResult", RetentionPolicy: "90d", Visibility: "advanced", SensitivityClass: "role_scoped", Description: "Organization-level planning and design outputs coordinated by Soma and operational leaders.", Reviewers: []string{"review", "admin"}, Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}, {Role: "specialist", CanRead: true, CanWrite: false}}},
	{Name: "organization.review.output", Type: "review", Owner: "system", SchemaID: "ReviewResult", RetentionPolicy: "120d", Visibility: "advanced", SensitivityClass: "role_scoped", Description: "Review and critique outputs routed back into governance and refinement flows.", Reviewers: []string{"review", "admin"}, Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "review", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "organization.learning.candidates", Type: "learning", Owner: "system", SchemaID: "LearningCandidate", RetentionPolicy: "180d", Visibility: "advanced", SensitivityClass: "team_scoped", Description: "Learning candidates discovered by teams, reviews, and automations.", Reviewers: []string{"review", "admin"}, Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "automation", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "browser.research.results", Type: "output", Owner: "mcp", SchemaID: "ToolResult", RetentionPolicy: "30d", Visibility: "advanced", SensitivityClass: "team_scoped", Description: "Normalized browser and research outputs from tools or MCP services.", Reviewers: []string{"review", "admin"}, Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "mcp", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "media.image.output", Type: "output", Owner: "mcp", SchemaID: "MediaResult", RetentionPolicy: "30d", Visibility: "advanced", SensitivityClass: "team_scoped", Description: "Generated or curated image/media outputs available for review and iteration.", Reviewers: []string{"review", "admin"}, Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "creative", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "api.data.output", Type: "output", Owner: "mcp", SchemaID: "ToolResult", RetentionPolicy: "30d", Visibility: "advanced", SensitivityClass: "team_scoped", Description: "Structured API and service output normalized for downstream reasoning.", Reviewers: []string{"review", "admin"}, Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "mcp", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
}

func SchemaByID(id string) (SchemaDefinition, bool) {
	for _, def := range SeedSchemas {
		if def.ID == id {
			return def, true
		}
	}
	return SchemaDefinition{}, false
}

func FieldByName(name string) (FieldDefinition, bool) {
	for _, def := range SeedFields {
		if def.Name == name {
			return def, true
		}
	}
	return FieldDefinition{}, false
}

func CapabilityByID(id string) (CapabilityDefinition, bool) {
	for _, def := range SeedCapabilities {
		if def.ID == id {
			return def, true
		}
	}
	return CapabilityDefinition{}, false
}
