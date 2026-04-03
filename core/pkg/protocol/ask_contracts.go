package protocol

// AskClass classifies the bounded operator intent posture resolved by chat.
// It is a runtime contract distinct from the user-visible terminal-state model.
type AskClass string

const (
	AskClassDirectAnswer     AskClass = "direct_answer"
	AskClassGovernedMutation AskClass = "governed_mutation"
	AskClassGovernedArtifact AskClass = "governed_artifact"
	AskClassSpecialist       AskClass = "specialist_consultation"
	AskClassExecutionBlocker AskClass = "execution_blocker"
)

// ApprovalPosture classifies how the ask is expected to interact with approval.
type ApprovalPosture string

const (
	ApprovalPostureAutoAllowed ApprovalPosture = "auto_allowed"
	ApprovalPostureOptional    ApprovalPosture = "optional"
	ApprovalPostureRequired    ApprovalPosture = "required"
)

// ArtifactPosture classifies how the ask expects artifact results to behave.
type ArtifactPosture string

const (
	ArtifactPostureNone     ArtifactPosture = "none"
	ArtifactPostureOptional ArtifactPosture = "optional"
	ArtifactPostureRequired ArtifactPosture = "required"
)

// AskContract is the shared machine-readable routing/output contract for a
// bounded ask class.
type AskContract struct {
	AskClass              AskClass        `json:"ask_class"`
	DefaultAgentTarget    string          `json:"default_agent_target"`
	AllowedAgentTargets   []string        `json:"allowed_agent_targets"`
	DefaultExecutionMode  ExecutionMode   `json:"default_execution_mode"`
	AllowedExecutionModes []ExecutionMode `json:"allowed_execution_modes"`
	TemplateID            TemplateID      `json:"template_id"`
	RequiresConfirmation  bool            `json:"requires_confirmation"`
	ApprovalPosture       ApprovalPosture `json:"approval_posture"`
	ArtifactPosture       ArtifactPosture `json:"artifact_posture"`
	Description           string          `json:"description"`
}

var askContractRegistry = map[AskClass]AskContract{
	AskClassDirectAnswer: {
		AskClass:              AskClassDirectAnswer,
		DefaultAgentTarget:    "soma",
		AllowedAgentTargets:   []string{"soma"},
		DefaultExecutionMode:  ModeAnswer,
		AllowedExecutionModes: []ExecutionMode{ModeAnswer},
		TemplateID:            TemplateChatToAnswer,
		RequiresConfirmation:  false,
		ApprovalPosture:       ApprovalPostureAutoAllowed,
		ArtifactPosture:       ArtifactPostureOptional,
		Description:           "Readable non-mutating response through Soma.",
	},
	AskClassGovernedMutation: {
		AskClass:              AskClassGovernedMutation,
		DefaultAgentTarget:    "soma",
		AllowedAgentTargets:   []string{"soma"},
		DefaultExecutionMode:  ModeProposal,
		AllowedExecutionModes: []ExecutionMode{ModeProposal},
		TemplateID:            TemplateChatToProposal,
		RequiresConfirmation:  true,
		ApprovalPosture:       ApprovalPostureRequired,
		ArtifactPosture:       ArtifactPostureOptional,
		Description:           "Mutation request that must enter governed proposal posture before execution.",
	},
	AskClassGovernedArtifact: {
		AskClass:              AskClassGovernedArtifact,
		DefaultAgentTarget:    "soma",
		AllowedAgentTargets:   []string{"soma", "specialist"},
		DefaultExecutionMode:  ModeAnswer,
		AllowedExecutionModes: []ExecutionMode{ModeAnswer},
		TemplateID:            TemplateChatToAnswer,
		RequiresConfirmation:  false,
		ApprovalPosture:       ApprovalPostureAutoAllowed,
		ArtifactPosture:       ArtifactPostureRequired,
		Description:           "Readable answer that returns durable or inline artifact output by default.",
	},
	AskClassSpecialist: {
		AskClass:              AskClassSpecialist,
		DefaultAgentTarget:    "specialist",
		AllowedAgentTargets:   []string{"soma", "specialist"},
		DefaultExecutionMode:  ModeAnswer,
		AllowedExecutionModes: []ExecutionMode{ModeAnswer},
		TemplateID:            TemplateChatToAnswer,
		RequiresConfirmation:  false,
		ApprovalPosture:       ApprovalPostureAutoAllowed,
		ArtifactPosture:       ArtifactPostureOptional,
		Description:           "Readable answer shaped by direct specialist routing or consultation.",
	},
}

// AskContractForClass returns the bounded shared contract for a known ask class.
func AskContractForClass(class AskClass) (AskContract, bool) {
	contract, ok := askContractRegistry[class]
	return contract, ok
}
