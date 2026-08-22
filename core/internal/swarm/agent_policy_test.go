package swarm

import (
	"strings"
	"testing"
)

func TestResponseSuggestsUnexecutedAction(t *testing.T) {
	tests := []struct {
		name string
		text string
		want bool
	}{
		{
			name: "detects step style delegation planning",
			text: "To have the team provide updates, we need to delegate a specific task. Step 1: Delegate Task",
			want: true,
		},
		{
			name: "detects permission-seeking phrasing",
			text: "Would you like me to do that?",
			want: true,
		},
		{
			name: "detects example input narration",
			text: "Example Input:\n{\n  \"operation\": \"consult_council\",\n  \"arguments\": {\"member\": \"council-architect\"}\n}\nThis will route your request to the Architect.",
			want: true,
		},
		{
			name: "ignores normal result response",
			text: "Task delegated to team admin-core.",
			want: false,
		},
		{
			name: "ignores empty",
			text: "",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := responseSuggestsUnexecutedAction(tt.text)
			if got != tt.want {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAutofillToolArguments_InfersNaturalApplicationPackage(t *testing.T) {
	call := &toolCallPayload{
		Name: "write_file",
		Arguments: map[string]any{
			"path":    "groups/app-team/generated/client-portal/play.html",
			"content": "<!doctype html><main>Ready</main>",
		},
	}

	autofillToolArguments(call, "Build and return a playable browser app with a direct entrypoint.")

	if call.Arguments["package_kind"] != "project_package" {
		t.Fatalf("package_kind = %v, want project_package", call.Arguments["package_kind"])
	}
	if files := stringSlice(call.Arguments["package_files"]); len(files) != 0 {
		t.Fatalf("package_files = %v, want no files before successful writes", files)
	}
}

func TestParseToolCall_FallbackOperationPayload(t *testing.T) {
	got := parseToolCall("{\"operation\":\"consult_council\",\"arguments\":{\"member\":\"council-architect\",\"question\":\"What API should we use?\"}}")
	if got == nil {
		t.Fatal("expected parsed tool call")
	}
	if got.Name != "consult_council" {
		t.Fatalf("name = %q, want consult_council", got.Name)
	}
	if got.Arguments["member"] != "council-architect" {
		t.Fatalf("member = %v, want council-architect", got.Arguments["member"])
	}
}

func TestParseToolCall_PrefersToolCallPayload(t *testing.T) {
	text := "{\"tool_call\":{\"name\":\"delegate_task\",\"arguments\":{\"team_id\":\"admin-core\",\"task\":\"x\"}}}\n{\"operation\":\"consult_council\",\"arguments\":{\"member\":\"council-architect\"}}"
	got := parseToolCall(text)
	if got == nil {
		t.Fatal("expected parsed tool call")
	}
	if got.Name != "delegate_task" {
		t.Fatalf("name = %q, want delegate_task", got.Name)
	}
}

func TestAutofillToolArguments_ConsultCouncilQuestion(t *testing.T) {
	call := &toolCallPayload{
		Name:      "consult_council",
		Arguments: map[string]any{"member": "council-architect"},
	}
	autofillToolArguments(call, "Use image API recommendations for this plan.")
	if call.Arguments["question"] != "Use image API recommendations for this plan." {
		t.Fatalf("question = %v", call.Arguments["question"])
	}
}

func TestAutofillToolArguments_DoesNotOverrideQuestion(t *testing.T) {
	call := &toolCallPayload{
		Name: "consult_council",
		Arguments: map[string]any{
			"member":   "council-architect",
			"question": "existing",
		},
	}
	autofillToolArguments(call, "new input")
	if call.Arguments["question"] != "existing" {
		t.Fatalf("question overwritten: %v", call.Arguments["question"])
	}
}

func TestAutofillToolArguments_ReadSignalsTopicPatternAlias(t *testing.T) {
	call := &toolCallPayload{
		Name:      "read_signals",
		Arguments: map[string]any{"topic_pattern": "swarm.team.admin-core.signal.status"},
	}
	autofillToolArguments(call, "check signals")
	if call.Arguments["subject"] != "swarm.team.admin-core.signal.status" {
		t.Fatalf("subject = %v", call.Arguments["subject"])
	}
}

func TestAutofillToolArguments_NormalizesFileAndArtifactAliases(t *testing.T) {
	writeCall := &toolCallPayload{
		Name:      "write_file",
		Arguments: map[string]any{"file_path": "groups/team/generated/index.html", "body": "<html></html>"},
	}
	autofillToolArguments(writeCall, "build the output")
	if writeCall.Arguments["path"] != "groups/team/generated/index.html" || writeCall.Arguments["content"] != "<html></html>" {
		t.Fatalf("write aliases = %#v", writeCall.Arguments)
	}

	artifactCall := &toolCallPayload{
		Name:      "store_artifact",
		Arguments: map[string]any{"kind": "document", "name": "Proof", "data": "validated"},
	}
	autofillToolArguments(artifactCall, "retain the proof")
	if artifactCall.Arguments["type"] != "document" || artifactCall.Arguments["title"] != "Proof" || artifactCall.Arguments["content"] != "validated" {
		t.Fatalf("artifact aliases = %#v", artifactCall.Arguments)
	}
}

func TestAutofillToolArguments_InfersDeclaredProjectPackageWrite(t *testing.T) {
	call := &toolCallPayload{
		Name: "write_file",
		Arguments: map[string]any{
			"file_path": "groups/team/generated/app/index.html",
			"content":   "<html></html>",
		},
	}
	autofillToolArguments(call, "Return a retained project_package. Use the package title Team App.")

	if call.Arguments["package_kind"] != "project_package" {
		t.Fatalf("package_kind = %#v", call.Arguments["package_kind"])
	}
	if call.Arguments["package_entrypoint"] != "groups/team/generated/app/index.html" {
		t.Fatalf("package_entrypoint = %#v", call.Arguments["package_entrypoint"])
	}
	if call.Arguments["package_folder"] != "groups/team/generated/app" {
		t.Fatalf("package_folder = %#v", call.Arguments["package_folder"])
	}
	if call.Arguments["package_title"] != "Team App" {
		t.Fatalf("package_title = %#v", call.Arguments["package_title"])
	}
	if files := stringSlice(call.Arguments["package_files"]); len(files) != 0 {
		t.Fatalf("package_files = %#v, want no synthesized support files", files)
	}
}

func TestAutofillToolArgumentsNormalizesProjectPackageContentMetadata(t *testing.T) {
	call := &toolCallPayload{Name: "store_artifact", Arguments: map[string]any{
		"key":  "project-package-a1b2c",
		"type": "project_package",
		"content": map[string]any{
			"entrypoint": "index.html",
			"folder":     "groups/app-team/deliveries/project-package-a1b2c",
			"files":      []any{"index.html", "README.md"},
			"validation": map[string]any{"pass": []any{"launch"}},
		},
	}}

	autofillToolArguments(call, "Retain the application package.")

	if call.Arguments["title"] != "project-package-a1b2c" {
		t.Fatalf("title = %#v", call.Arguments["title"])
	}
	if _, ok := call.Arguments["content"].(string); !ok {
		t.Fatalf("content = %#v, want JSON string", call.Arguments["content"])
	}
	metadata, ok := call.Arguments["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata = %#v", call.Arguments["metadata"])
	}
	if metadata["entrypoint"] != "index.html" || metadata["folder"] != "groups/app-team/deliveries/project-package-a1b2c" {
		t.Fatalf("metadata = %#v", metadata)
	}
	if files := stringSlice(metadata["files"]); len(files) != 2 {
		t.Fatalf("files = %#v", files)
	}
	if validation, ok := metadata["validation"].(string); !ok || !strings.Contains(validation, "launch") {
		t.Fatalf("validation = %#v", metadata["validation"])
	}
}

func TestNormalizeAgentToolCallArgumentsScopesProjectPackageMetadataToTeam(t *testing.T) {
	call := &toolCallPayload{Name: "store_artifact", Arguments: map[string]any{
		"type":    "project_package",
		"title":   "Application package",
		"content": `{"kind":"project_package"}`,
		"metadata": map[string]any{
			"folder":     "application-delivery-team-3ff8d/project-package-698b9fa7",
			"entrypoint": "index.html",
			"files":      []any{"index.html", "README.md"},
		},
	}}

	normalizeAgentToolCallArguments(call, "application-delivery-team-3ff8d", "Develop a playable browser application.")

	metadata, ok := call.Arguments["metadata"].(map[string]any)
	if !ok {
		t.Fatalf("metadata = %#v", call.Arguments["metadata"])
	}
	if metadata["folder"] != "groups/application-delivery-team-3ff8d/generated/project-package-698b9fa7" {
		t.Fatalf("folder = %#v", metadata["folder"])
	}
	if metadata["entrypoint"] != "index.html" {
		t.Fatalf("entrypoint = %#v", metadata["entrypoint"])
	}
}

func TestNormalizeCouncilMember(t *testing.T) {
	tests := map[string]string{
		"Architect":         "council-architect",
		"council architect": "council-architect",
		"council-coder":     "council-coder",
		"Sentry":            "council-sentry",
	}
	for in, want := range tests {
		if got := normalizeCouncilMember(in); got != want {
			t.Fatalf("normalizeCouncilMember(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAutofillToolArguments_ReadSignalsExtractsSubjectFromInput(t *testing.T) {
	call := &toolCallPayload{
		Name:      "read_signals",
		Arguments: map[string]any{},
	}
	autofillToolArguments(call, "read swarm.team.OpenClawSearchTeam.signal.status and report back")
	if call.Arguments["subject"] != "swarm.team.OpenClawSearchTeam.signal.status" {
		t.Fatalf("subject = %v", call.Arguments["subject"])
	}
}

func TestExtractNATSSubject(t *testing.T) {
	got := extractNATSSubject("please inspect (swarm.team.admin-core.signal.status). now")
	if got != "swarm.team.admin-core.signal.status" {
		t.Fatalf("got %q", got)
	}
}

func TestParseLooseToolCall(t *testing.T) {
	text := `{"tool_call":{"name":"consult_council","arguments":{"member":"council-architect"}}`
	got := parseLooseToolCall(text)
	if got == nil {
		t.Fatal("expected loose tool call parse")
	}
	if got.Name != "consult_council" {
		t.Fatalf("name = %q", got.Name)
	}
}

func TestAutofillToolArguments_ConsultCouncilInfersMember(t *testing.T) {
	call := &toolCallPayload{
		Name:      "consult_council",
		Arguments: map[string]any{},
	}
	autofillToolArguments(call, "Consult the architect about API strategy")
	if call.Arguments["member"] != "council-architect" {
		t.Fatalf("member = %v", call.Arguments["member"])
	}
	if call.Arguments["question"] == nil {
		t.Fatal("expected inferred question")
	}
}

func TestAutofillToolArguments_DelegateTaskPromotesCreateTeam(t *testing.T) {
	call := &toolCallPayload{
		Name: "delegate_task",
		Arguments: map[string]any{
			"team_name":  "OpenClawSearchTeam",
			"agent_type": "coder",
		},
	}
	autofillToolArguments(call, "create team")
	if call.Name != "create_team" {
		t.Fatalf("name = %q, want create_team", call.Name)
	}
	if call.Arguments["team_id"] != "OpenClawSearchTeam" {
		t.Fatalf("team_id = %v", call.Arguments["team_id"])
	}
	if call.Arguments["role"] != "coder" {
		t.Fatalf("role = %v", call.Arguments["role"])
	}
}
