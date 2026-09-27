package cognitive

// Execution blocker copy (UX1). RecommendedAction is read by anyone, so it is
// plain language with no API path, env var, URL or permission string.
// AdminAction carries the admin remedy and is never serialized.
const (
	// UserEngineSetupAction is the user-safe remedy when Soma cannot run.
	UserEngineSetupAction = "Try again in a moment. If it keeps happening, ask an admin to finish setting up Soma's AI engine."
	// AdminEngineSetupAction is the admin remedy when no specific hint exists.
	AdminEngineSetupAction = "Open Settings > AI Engines and make sure the engine Soma uses is on and has a model selected, or use the default engine."
	// SummaryRouterUnavailable is the summary when the router is not loaded.
	SummaryRouterUnavailable = "Soma can't answer right now because its AI engine service is offline."
)
