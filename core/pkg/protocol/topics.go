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

	// Global Broadcast (Mission Control → All Teams)
	TopicGlobalBroadcast = "swarm.global.broadcast"

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

	// Council request-reply (direct addressing)
	TopicCouncilRequestFmt = "swarm.council.%s.request" // agent ID

	// Sensor Data (Ingress feeds)
	TopicSensorDataWild    = "swarm.data.>"
	TopicSensorDataEmail   = "swarm.data.email.>"
	TopicSensorDataWeather = "swarm.data.weather.>"
	TopicSensorDataMCP     = "swarm.data.mcp.>"

	// V7 Event Spine (Team A) — mission run lifecycle signals.
	// CTS signals carry mission_event_id to link back to the persistent audit record.
	// Persistent audit records live in mission_events table (DB-first rule).
	TopicMissionEvents    = "swarm.mission.events"
	TopicMissionEventsFmt = "swarm.mission.events.%s" // run_id
	TopicMissionRunsFmt   = "swarm.mission.runs.%s"   // mission_id
)
