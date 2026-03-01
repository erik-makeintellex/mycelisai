package protocol

// DecisionPath defines the high-level execution strategy selected by Soma.
type DecisionPath string

const (
	DecisionPathDirect          DecisionPath = "direct"
	DecisionPathManifestTeam    DecisionPath = "manifest_team"
	DecisionPathPropose         DecisionPath = "propose"
	DecisionPathScheduledRepeat DecisionPath = "scheduled_repeat"
)

// TeamLifetime controls how long an instantiated team should remain active.
type TeamLifetime string

const (
	TeamLifetimeEphemeral  TeamLifetime = "ephemeral"
	TeamLifetimePersistent TeamLifetime = "persistent"
	TeamLifetimeAuto       TeamLifetime = "auto"
)

// SomaDecisionFrame is the canonical decision contract for extension-of-self routing.
type SomaDecisionFrame struct {
	RequestID            string       `json:"request_id"`
	RunID                string       `json:"run_id,omitempty"`
	PathSelected         DecisionPath `json:"path_selected"`
	TeamLifetime         TeamLifetime `json:"team_lifetime"`
	RiskLevel            string       `json:"risk_level"`
	Confidence           float64      `json:"confidence"`
	ApprovalRequired     bool         `json:"approval_required"`
	RequiredCapabilities []string     `json:"required_capabilities,omitempty"`
	FallbackPlan         string       `json:"fallback_plan,omitempty"`
}

// HeartbeatBudget bounds long-horizon autonomy loops.
type HeartbeatBudget struct {
	MaxDurationMinutes int     `json:"max_duration_minutes"`
	MaxActions         int     `json:"max_actions"`
	MaxSpendUSD        float64 `json:"max_spend_usd"`
	MaxEscalations     int     `json:"max_escalations"`
}

// UniversalInvokeResult is the canonical action execution envelope used by all adapters.
type UniversalInvokeResult struct {
	RequestID      string `json:"request_id"`
	RunID          string `json:"run_id,omitempty"`
	ActionID       string `json:"action_id"`
	ProviderType   string `json:"provider_type"` // mcp|openapi|python|hardware
	Status         string `json:"status"`        // success|degraded|failure
	IdempotencyKey string `json:"idempotency_key"`
	Output         any    `json:"output,omitempty"`
}

// InceptionContractBundle exposes the frozen P0 contract shapes and allowed values.
type InceptionContractBundle struct {
	DecisionFrame    SomaDecisionFrame     `json:"decision_frame"`
	HeartbeatBudget  HeartbeatBudget       `json:"heartbeat_budget"`
	UniversalInvoke  UniversalInvokeResult `json:"universal_invoke"`
	AllowedPaths     []DecisionPath        `json:"allowed_paths"`
	AllowedLifetimes []TeamLifetime        `json:"allowed_lifetimes"`
}

// DefaultInceptionContractBundle returns the P0 baseline contract bundle.
func DefaultInceptionContractBundle() InceptionContractBundle {
	return InceptionContractBundle{
		DecisionFrame: SomaDecisionFrame{
			RequestID:            "req_example",
			RunID:                "run_example",
			PathSelected:         DecisionPathDirect,
			TeamLifetime:         TeamLifetimeEphemeral,
			RiskLevel:            "low",
			Confidence:           0.8,
			ApprovalRequired:     false,
			RequiredCapabilities: []string{"mcp:filesystem/*"},
			FallbackPlan:         "propose",
		},
		HeartbeatBudget: HeartbeatBudget{
			MaxDurationMinutes: 180,
			MaxActions:         50,
			MaxSpendUSD:        5.0,
			MaxEscalations:     3,
		},
		UniversalInvoke: UniversalInvokeResult{
			RequestID:      "req_example",
			RunID:          "run_example",
			ActionID:       "filesystem.read_file",
			ProviderType:   "mcp",
			Status:         "success",
			IdempotencyKey: "idem_example",
			Output:         map[string]any{"text": "ok"},
		},
		AllowedPaths: []DecisionPath{
			DecisionPathDirect,
			DecisionPathManifestTeam,
			DecisionPathPropose,
			DecisionPathScheduledRepeat,
		},
		AllowedLifetimes: []TeamLifetime{
			TeamLifetimeEphemeral,
			TeamLifetimePersistent,
			TeamLifetimeAuto,
		},
	}
}
