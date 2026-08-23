package swarm

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

type resultContractProvider struct {
	responses        []string
	errors           map[int]error
	calls            int
	rejectBlankTurns bool
}

func (provider *resultContractProvider) Infer(_ context.Context, _ string, options cognitive.InferOptions) (*cognitive.InferResponse, error) {
	if provider.rejectBlankTurns {
		for _, message := range options.Messages {
			if strings.TrimSpace(message.Content) == "" {
				return nil, errors.New("provider rejected blank conversation turn")
			}
		}
	}
	index := provider.calls
	provider.calls++
	if provider.errors != nil {
		if err := provider.errors[index]; err != nil {
			return nil, err
		}
	}
	if index >= len(provider.responses) {
		index = len(provider.responses) - 1
	}
	return &cognitive.InferResponse{Text: provider.responses[index], Provider: "mock", ModelUsed: "contract-test"}, nil
}

func (provider *resultContractProvider) Probe(context.Context) (bool, error) { return true, nil }

type resultContractToolExecutor struct {
	calls        []string
	fail         bool
	failRead     bool
	files        map[string]string
	readOverride map[string]string
}

func (executor *resultContractToolExecutor) FindToolByName(_ context.Context, name string) (uuid.UUID, string, error) {
	return InternalServerID, name, nil
}

func (executor *resultContractToolExecutor) CallTool(_ context.Context, _ uuid.UUID, name string, args map[string]any) (string, error) {
	executor.calls = append(executor.calls, name)
	if executor.fail || (executor.failRead && name == "read_file") {
		return "", errors.New("write unavailable")
	}
	path := cleanEvidencePath(stringValue(args["path"]))
	switch name {
	case "write_file":
		return executor.writeFile(path, name, args)
	case "read_file", "read_text_file":
		if content, ok := executor.readOverride[path]; ok {
			return content, nil
		}
		return executor.files[path], nil
	}
	return "completed:" + name, nil
}

func (executor *resultContractToolExecutor) writeFile(path string, name string, args map[string]any) (string, error) {
	if executor.files == nil {
		executor.files = map[string]string{}
	}
	executor.files[path] = stringValue(args["content"])
	if !strings.EqualFold(strings.TrimSpace(stringValue(args["package_kind"])), "project_package") {
		return "completed:" + name, nil
	}
	folder := strings.TrimSpace(stringValue(args["package_folder"]))
	if folder == "" {
		folder = path
		if index := strings.LastIndex(folder, "/"); index >= 0 {
			folder = folder[:index]
		}
	}
	entrypoint := strings.TrimSpace(stringValue(args["package_entrypoint"]))
	if entrypoint == "" {
		entrypoint = path
	}
	title := strings.TrimSpace(stringValue(args["package_title"]))
	if title == "" {
		title = "Generated project package"
	}
	return mustJSON(map[string]any{
		"message": "completed:" + name,
		"artifact": protocol.ChatArtifactRef{
			Type: "project_package", Title: title, ContentType: "application/vnd.mycelis.project+json",
			SavedPath: folder, Entrypoint: entrypoint, Folder: folder, Files: projectPackageSupportFileNames(args),
		},
	}), nil
}

func resultContractTestAgent(provider cognitive.LLMProvider, executor MCPToolExecutor) *Agent {
	router := &cognitive.Router{
		Config: &cognitive.BrainConfig{
			Profiles:  map[string]string{"chat": "mock"},
			Providers: map[string]cognitive.ProviderConfig{"mock": {Type: "mock", Enabled: true, ModelID: "contract-test"}},
		},
		Adapters: map[string]cognitive.LLMProvider{"mock": provider},
	}
	agent := NewAgent(context.Background(), protocol.AgentManifest{
		ID: "worker", Role: "implementer", Provider: "mock", Tools: []string{"write_file", "read_file"}, MaxIterations: 6,
	}, "delivery-team", nil, router, executor)
	agent.SetToolDescriptions(map[string]string{"write_file": "Write retained output.", "read_file": "Read retained output."})
	return agent
}

func testStringSliceContains(items []string, want string) bool {
	for _, item := range items {
		if item == want {
			return true
		}
	}
	return false
}
