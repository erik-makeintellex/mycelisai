package swarm

import (
	"time"

	"github.com/mycelis/core/pkg/protocol"
)

// SymbioticSeedBlueprint returns a pre-built MissionBlueprint for single-host testing.
// It defines a "Symbiotic Sensors" team with Gmail and Weather polling agents.
// No LLM generation required — this blueprint can be committed directly.
func SymbioticSeedBlueprint() *protocol.MissionBlueprint {
	return &protocol.MissionBlueprint{
		MissionID: "mission-symbiotic-seed",
		Intent:    "Internal sensor team: Gmail inbox monitoring + local weather polling",
		Teams: []protocol.BlueprintTeam{
			{
				Name: "Symbiotic Sensors",
				Role: "Automated sensor data acquisition (no LLM inference)",
				Agents: []protocol.AgentManifest{
					{
						ID:      "sensor-gmail-poller",
						Role:    "gmail_sensor",
						Inputs:  []string{},
						Outputs: []string{protocol.TopicSensorDataEmail},
					},
					{
						ID:      "sensor-weather-poller",
						Role:    "weather_sensor",
						Inputs:  []string{},
						Outputs: []string{protocol.TopicSensorDataWeather},
					},
				},
			},
		},
		Constraints: []protocol.Constraint{
			{Description: "Poll interval: 60 seconds"},
			{Description: "No LLM inference — pure data acquisition"},
			{Description: "Sensor trust score: 1.0 (fully trusted)"},
		},
	}
}

// SymbioticSeedSensorConfigs returns default SensorConfig entries for the
// symbiotic seed agents. Empty endpoints = heartbeat-only mode for testing.
func SymbioticSeedSensorConfigs() map[string]SensorConfig {
	return map[string]SensorConfig{
		"sensor-gmail-poller": {
			Type:     SensorTypeHTTP,
			Endpoint: "", // Heartbeat-only: no real Gmail API in test mode
			Interval: 60 * time.Second,
		},
		"sensor-weather-poller": {
			Type:     SensorTypeHTTP,
			Endpoint: "", // Heartbeat-only: no real Weather API in test mode
			Interval: 60 * time.Second,
		},
	}
}
