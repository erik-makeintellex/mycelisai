package server

import (
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestBuildSensorConfigs(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: []protocol.BlueprintTeam{
			{
				Name: "sensors",
				Agents: []protocol.AgentManifest{
					{ID: "weather-sensor", Role: "sensory"},
					{ID: "gmail-sensor", Role: "Sensor Agent"},
					{ID: "worker", Role: "cognitive"},
				},
			},
		},
	}

	configs := buildSensorConfigs(bp)
	if len(configs) != 2 {
		t.Fatalf("Expected 2 sensor configs, got %d", len(configs))
	}
	if _, ok := configs["weather-sensor"]; !ok {
		t.Error("Expected 'weather-sensor' in configs")
	}
	if _, ok := configs["gmail-sensor"]; !ok {
		t.Error("Expected 'gmail-sensor' in configs")
	}
	if _, ok := configs["worker"]; ok {
		t.Error("'worker' should not be in sensor configs")
	}
}

func TestBuildSensorConfigs_Empty(t *testing.T) {
	bp := &protocol.MissionBlueprint{
		Teams: []protocol.BlueprintTeam{
			{
				Name:   "no-sensors",
				Agents: []protocol.AgentManifest{{ID: "worker", Role: "cognitive"}},
			},
		},
	}

	configs := buildSensorConfigs(bp)
	if len(configs) != 0 {
		t.Errorf("Expected 0 sensor configs, got %d", len(configs))
	}
}

func TestExtractBlueprintFromResponse_PlainJSON(t *testing.T) {
	json := `{"mission_id":"m-1","intent":"test","teams":[{"name":"t1","agents":[]}]}`
	bp, err := extractBlueprintFromResponse(json)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp.MissionID != "m-1" {
		t.Errorf("Expected mission_id 'm-1', got %q", bp.MissionID)
	}
}

func TestExtractBlueprintFromResponse_MarkdownFenced(t *testing.T) {
	response := "Here is the blueprint:\n```json\n" +
		`{"mission_id":"m-2","intent":"test","teams":[{"name":"t1","agents":[]}]}` +
		"\n```\nDone."
	bp, err := extractBlueprintFromResponse(response)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if bp.MissionID != "m-2" {
		t.Errorf("Expected mission_id 'm-2', got %q", bp.MissionID)
	}
}

func TestExtractBlueprintFromResponse_NoJSON(t *testing.T) {
	_, err := extractBlueprintFromResponse("no json here")
	if err == nil {
		t.Error("Expected error for response with no JSON")
	}
	if !strings.Contains(err.Error(), "no JSON object found") {
		t.Errorf("Expected 'no JSON object found' error, got: %v", err)
	}
}

func TestExtractBlueprintFromResponse_InvalidBlueprint(t *testing.T) {
	_, err := extractBlueprintFromResponse(`{"foo":"bar"}`)
	if err == nil {
		t.Error("Expected error for invalid blueprint")
	}
}
