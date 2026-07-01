package workers

import "testing"

func TestResultFromMapNormalizesOutputsAndOutputRefs(t *testing.T) {
	result := resultFromMap(map[string]any{
		"summary": "outputs ready",
		"outputs": []any{
			map[string]any{
				"id":           "artifact-1",
				"kind":         "file",
				"title":        "Run package",
				"href":         "/api/v1/workspace/files/view?path=generated%2Frun.md",
				"content_type": "text/markdown",
			},
		},
		"output_refs": []any{
			map[string]any{
				"output_id":   "team-output-1",
				"kind":        "team",
				"label":       "Team brief",
				"storage_ref": "groups/team/brief.md",
			},
			"proof://run-1/output-2",
		},
	})

	if result == nil {
		t.Fatal("resultFromMap returned nil")
	}
	if len(result.Outputs) != 3 {
		t.Fatalf("outputs = %#v, want 3 normalized outputs", result.Outputs)
	}
	if got := result.Outputs[0]; got.ID != "artifact-1" || got.Kind != "file" || got.Name != "Run package" || got.URI == "" || got.ContentType != "text/markdown" {
		t.Fatalf("outputs[0] = %#v", got)
	}
	if got := result.Outputs[1]; got.ID != "team-output-1" || got.Kind != "team" || got.Name != "Team brief" || got.URI != "groups/team/brief.md" {
		t.Fatalf("outputs[1] = %#v", got)
	}
	if got := result.Outputs[2]; got.Kind != "reference" || got.URI != "proof://run-1/output-2" {
		t.Fatalf("outputs[2] = %#v", got)
	}
}

func TestEventFromMapNormalizesTopLevelOutputRefs(t *testing.T) {
	event := eventFromMap(map[string]any{
		"status":  "completed",
		"summary": "team output ready",
		"output_refs": []any{
			map[string]any{
				"output_id":  "output-1",
				"label":      "Playable package",
				"entrypoint": "generated/game/index.html",
			},
		},
	}, "run-1", BackendHermesAPI)

	if event.Kind != EventCompleted {
		t.Fatalf("kind = %s, want %s", event.Kind, EventCompleted)
	}
	if event.Result == nil {
		t.Fatal("expected top-level output_refs to produce a worker result")
	}
	if len(event.Result.Outputs) != 1 {
		t.Fatalf("outputs = %#v, want one normalized output", event.Result.Outputs)
	}
	if got := event.Result.Outputs[0]; got.ID != "output-1" || got.Name != "Playable package" || got.URI != "generated/game/index.html" {
		t.Fatalf("output = %#v", got)
	}
}
