package protocol

// Centralized NATS topic constants for the Mycelis Nervous System.
// All components MUST use these constants instead of hardcoded strings.

const (
	// Global Frequencies (Layer 1)
	TopicGlobalHeartbeat = "swarm.global.heartbeat"
	TopicGlobalAnnounce  = "swarm.global.announce"
	TopicAuditTrace      = "swarm.audit.trace"

	// Global Input (Ingress - guarded)
	TopicGlobalInputWild = "swarm.global.input.>"
	TopicGlobalInputUser = "swarm.global.input.user"

	// Format strings requiring fmt.Sprintf with agent/team IDs
	TopicAgentOutputFmt      = "swarm.agent.%s.output"           // agent ID
	TopicTeamInternalTrigger = "swarm.team.%s.internal.trigger"  // team ID
	TopicTeamInternalRespond = "swarm.team.%s.internal.response" // team ID
	TopicTeamInternalCommand = "swarm.team.%s.internal.command"  // team ID
	TopicTeamSignalStatus    = "swarm.team.%s.signal.status"     // team ID

	// Telemetry (CTS Envelope bus)
	TopicTeamTelemetryFmt  = "swarm.team.%s.telemetry" // team ID
	TopicTeamTelemetryWild = "swarm.team.*.telemetry"

	// Mission DAG (Overseer)
	TopicMissionTask = "swarm.mission.task"

	// Wildcard subscriptions
	TopicTeamInternalWild = "swarm.team.*.internal.>"
	TopicSwarmWild        = "swarm.>"

	// Sensor Data (Ingress feeds)
	TopicSensorDataWild    = "swarm.data.>"
	TopicSensorDataEmail   = "swarm.data.email.>"
	TopicSensorDataWeather = "swarm.data.weather.>"
	TopicSensorDataMCP     = "swarm.data.mcp.>"
)
