package protocol

import (
	"encoding/json"
	"testing"
)

// TestBlueprintParsing_ObjectConstraints tests the exact JSON payload
// that the LLM generated during the Phase 4 integration test.
// The LLM returned constraints as objects rather than strings.
func TestBlueprintParsing_ObjectConstraints(t *testing.T) {
	raw := `{
  "mission_id": "mission-01",
  "intent": "Monitor server room temperature and alert if above 30C",
  "teams": [
    {
      "name": "Temperature Monitoring Team",
      "role": "Responsible for monitoring the temperature of the server room.",
      "agents": [
        {
          "id": "agent-01",
          "role": "Temperature Sensor Agent",
          "system_prompt": "You are a device that reads the current temperature.",
          "inputs": ["sensor/server-room-temperature"],
          "outputs": ["data/temperature-data"]
        },
        {
          "id": "agent-02",
          "role": "Temperature Data Processor Agent",
          "system_prompt": "Compare the reading with the threshold of 30C.",
          "inputs": ["data/temperature-data"],
          "outputs": ["alert/temperature-alert"]
        }
      ]
    }
  ],
  "constraints": [
    {
      "constraint_id": "c-01",
      "description": "The temperature sensor agent must send readings every 5 minutes."
    },
    {
      "constraint_id": "c-02",
      "description": "Trigger an alert if above 30C for more than 1 minute."
    }
  ]
}`

	var bp MissionBlueprint
	if err := json.Unmarshal([]byte(raw), &bp); err != nil {
		t.Fatalf("Failed to unmarshal blueprint with object constraints: %v", err)
	}

	if bp.MissionID != "mission-01" {
		t.Errorf("MissionID = %q, want %q", bp.MissionID, "mission-01")
	}
	if len(bp.Teams) != 1 {
		t.Fatalf("Teams count = %d, want 1", len(bp.Teams))
	}
	if len(bp.Teams[0].Agents) != 2 {
		t.Errorf("Agents count = %d, want 2", len(bp.Teams[0].Agents))
	}
	if len(bp.Constraints) != 2 {
		t.Fatalf("Constraints count = %d, want 2", len(bp.Constraints))
	}
	if bp.Constraints[0].ID != "c-01" {
		t.Errorf("Constraint[0].ID = %q, want %q", bp.Constraints[0].ID, "c-01")
	}
	if bp.Constraints[0].Description != "The temperature sensor agent must send readings every 5 minutes." {
		t.Errorf("Constraint[0].Description = %q", bp.Constraints[0].Description)
	}
}

// TestBlueprintParsing_StringConstraints tests the prompt-instructed format
// where constraints are plain strings.
func TestBlueprintParsing_StringConstraints(t *testing.T) {
	raw := `{
  "mission_id": "mission-02",
  "intent": "Test string constraints",
  "teams": [],
  "constraints": ["Max 3 retries", "Timeout after 60s"]
}`

	var bp MissionBlueprint
	if err := json.Unmarshal([]byte(raw), &bp); err != nil {
		t.Fatalf("Failed to unmarshal blueprint with string constraints: %v", err)
	}

	if len(bp.Constraints) != 2 {
		t.Fatalf("Constraints count = %d, want 2", len(bp.Constraints))
	}
	if bp.Constraints[0].Description != "Max 3 retries" {
		t.Errorf("Constraint[0].Description = %q, want %q", bp.Constraints[0].Description, "Max 3 retries")
	}
	if bp.Constraints[1].Description != "Timeout after 60s" {
		t.Errorf("Constraint[1].Description = %q, want %q", bp.Constraints[1].Description, "Timeout after 60s")
	}
	// String form should have empty ID
	if bp.Constraints[0].ID != "" {
		t.Errorf("Constraint[0].ID = %q, want empty", bp.Constraints[0].ID)
	}
}

// TestBlueprintParsing_NoConstraints ensures omitted constraints don't break parsing.
func TestBlueprintParsing_NoConstraints(t *testing.T) {
	raw := `{"mission_id": "mission-03", "intent": "No constraints", "teams": []}`

	var bp MissionBlueprint
	if err := json.Unmarshal([]byte(raw), &bp); err != nil {
		t.Fatalf("Failed to unmarshal: %v", err)
	}
	if len(bp.Constraints) != 0 {
		t.Errorf("Constraints count = %d, want 0", len(bp.Constraints))
	}
}

// TestBlueprintRoundTrip ensures marshal/unmarshal round-trips cleanly.
func TestBlueprintRoundTrip(t *testing.T) {
	bp := MissionBlueprint{
		MissionID: "mission-rt",
		Intent:    "Round trip test",
		Teams: []BlueprintTeam{
			{
				Name: "Alpha",
				Role: "Builder",
				Agents: []AgentManifest{
					{ID: "a1", Role: "coder", Inputs: []string{"in"}, Outputs: []string{"out"}},
				},
			},
		},
		Constraints: []Constraint{
			{ID: "c-1", Description: "Must pass tests"},
			{Description: "Budget under 100"},
		},
	}

	data, err := json.Marshal(bp)
	if err != nil {
		t.Fatalf("Marshal failed: %v", err)
	}

	var bp2 MissionBlueprint
	if err := json.Unmarshal(data, &bp2); err != nil {
		t.Fatalf("Unmarshal failed: %v", err)
	}

	if bp2.MissionID != bp.MissionID {
		t.Errorf("MissionID = %q, want %q", bp2.MissionID, bp.MissionID)
	}
	if len(bp2.Constraints) != 2 {
		t.Fatalf("Constraints count = %d, want 2", len(bp2.Constraints))
	}
	if bp2.Constraints[0].ID != "c-1" {
		t.Errorf("Constraint[0].ID = %q, want %q", bp2.Constraints[0].ID, "c-1")
	}
}
