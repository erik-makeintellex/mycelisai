package exchange

var SeedFields = []FieldDefinition{
	{Name: "summary", Type: "string", SemanticMeaning: "Operator-readable summary of the exchange output.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "status", Type: "enum", SemanticMeaning: "Lifecycle state of the work or output.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "priority", Type: "enum", SemanticMeaning: "Relative urgency or handling priority.", Indexed: true, Visibility: "soma", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "artifact_uri", Type: "reference", SemanticMeaning: "Stable pointer to a stored artifact or generated output.", Indexed: true, Visibility: "soma", UsageContexts: []string{"artifact", "message"}},
	{Name: "source_role", Type: "string", SemanticMeaning: "Role or system that produced the exchange item.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message"}},
	{Name: "target_role", Type: "string", SemanticMeaning: "Role, team, or system expected to consume the exchange item.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "confidence", Type: "number", SemanticMeaning: "Confidence signal attached to an output or review.", Indexed: true, Visibility: "soma", UsageContexts: []string{"artifact", "message"}},
	{Name: "tags", Type: "array", SemanticMeaning: "Loose classification labels for discovery and filtering.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "continuity_key", Type: "string", SemanticMeaning: "Stable continuity hint linking related work over time.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "created_at", Type: "string", SemanticMeaning: "Creation timestamp for ordering and retention.", Indexed: true, Visibility: "default", UsageContexts: []string{"artifact", "message", "thread"}},
	{Name: "updated_at", Type: "string", SemanticMeaning: "Most recent update timestamp for the exchange object.", Indexed: true, Visibility: "admin", UsageContexts: []string{"artifact", "message", "thread"}},
}

var SeedSchemas = []SchemaDefinition{
	{ID: "TextResult", Label: "Text Result", Description: "Plain language output intended for review or handoff.", RequiredFields: []string{"summary", "status", "source_role", "target_role", "created_at"}, OptionalFields: []string{"confidence", "tags", "continuity_key", "updated_at"}, RequiredCapabilities: []string{"text_output"}},
	{ID: "PlanResult", Label: "Plan Result", Description: "Structured planning output for work design and coordination.", RequiredFields: []string{"summary", "status", "priority", "source_role", "target_role", "created_at"}, OptionalFields: []string{"tags", "continuity_key", "updated_at"}, RequiredCapabilities: []string{"planning"}},
	{ID: "ReviewResult", Label: "Review Result", Description: "Review, critique, or governance feedback on work already produced.", RequiredFields: []string{"summary", "status", "confidence", "source_role", "target_role", "created_at"}, OptionalFields: []string{"tags", "continuity_key", "updated_at"}, RequiredCapabilities: []string{"review"}},
	{ID: "MediaResult", Label: "Media Result", Description: "Media output such as generated imagery or referenced audio/video assets.", RequiredFields: []string{"summary", "status", "artifact_uri", "source_role", "target_role", "created_at"}, OptionalFields: []string{"confidence", "tags", "continuity_key", "updated_at"}, RequiredCapabilities: []string{"media_output"}},
	{ID: "FileResult", Label: "File Result", Description: "File-oriented output intended for later delivery or inspection.", RequiredFields: []string{"summary", "status", "artifact_uri", "source_role", "target_role", "created_at"}, OptionalFields: []string{"tags", "continuity_key", "updated_at"}, RequiredCapabilities: []string{"file_output"}},
	{ID: "ToolResult", Label: "Tool Result", Description: "Normalized output from MCP or service/tool invocations.", RequiredFields: []string{"summary", "status", "source_role", "target_role", "created_at"}, OptionalFields: []string{"artifact_uri", "confidence", "tags", "continuity_key", "updated_at"}, RequiredCapabilities: []string{"tool_execution"}},
	{ID: "LearningCandidate", Label: "Learning Candidate", Description: "Candidate learning signal surfaced for future synthesis or promotion.", RequiredFields: []string{"summary", "status", "confidence", "tags", "continuity_key", "created_at"}, OptionalFields: []string{"source_role", "target_role", "updated_at"}, RequiredCapabilities: []string{"learning"}},
	{ID: "Escalation", Label: "Escalation", Description: "Escalation or blocker state requiring higher-level review.", RequiredFields: []string{"summary", "status", "priority", "source_role", "target_role", "created_at"}, OptionalFields: []string{"confidence", "tags", "continuity_key", "updated_at"}, RequiredCapabilities: []string{"escalation"}},
}

var SeedChannels = []Channel{
	{Name: "organization.planning.work", Type: "planning", Owner: "system", SchemaID: "PlanResult", RetentionPolicy: "90d", Visibility: "advanced", Description: "Organization-level planning and design outputs coordinated by Soma and operational leaders.", Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}, {Role: "specialist", CanRead: true, CanWrite: false}}},
	{Name: "organization.review.output", Type: "review", Owner: "system", SchemaID: "ReviewResult", RetentionPolicy: "120d", Visibility: "advanced", Description: "Review and critique outputs routed back into governance and refinement flows.", Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "review", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "organization.learning.candidates", Type: "learning", Owner: "system", SchemaID: "LearningCandidate", RetentionPolicy: "180d", Visibility: "advanced", Description: "Learning candidates discovered by teams, reviews, and automations.", Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "automation", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "browser.research.results", Type: "output", Owner: "mcp", SchemaID: "ToolResult", RetentionPolicy: "30d", Visibility: "advanced", Description: "Normalized browser and research outputs from tools or MCP services.", Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "mcp", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "media.image.output", Type: "output", Owner: "mcp", SchemaID: "MediaResult", RetentionPolicy: "30d", Visibility: "advanced", Description: "Generated or curated image/media outputs available for review and iteration.", Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "creative", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
	{Name: "api.data.output", Type: "output", Owner: "mcp", SchemaID: "ToolResult", RetentionPolicy: "30d", Visibility: "advanced", Description: "Structured API and service output normalized for downstream reasoning.", Participants: []ChannelParticipant{{Role: "soma", CanRead: true, CanWrite: true}, {Role: "mcp", CanRead: true, CanWrite: true}, {Role: "team_lead", CanRead: true, CanWrite: true}}},
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
