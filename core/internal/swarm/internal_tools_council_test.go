package swarm

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	server "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func TestHandleConsultCouncil_PreservesStructuredArtifacts(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	if _, err := nc.Subscribe("swarm.council.council-creative.request", func(msg *nats.Msg) {
		resp := map[string]any{
			"text": "I prepared a visual direction and a concept note.",
			"artifacts": []map[string]any{
				{
					"id":           "img-1",
					"type":         "image",
					"title":        "Concept Key Art",
					"content_type": "image/png",
					"content":      "cG5n",
					"cached":       true,
				},
				{
					"id":           "doc-1",
					"type":         "document",
					"title":        "Concept Notes",
					"content_type": "text/markdown",
					"content":      "# Notes",
				},
			},
		}
		data, _ := json.Marshal(resp)
		if msg.Reply != "" {
			_ = nc.Publish(msg.Reply, data)
		}
	}); err != nil {
		t.Fatalf("subscribe council request: %v", err)
	}
	nc.Flush()

	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc})

	out, err := reg.handleConsultCouncil(context.Background(), map[string]any{
		"member":   "creative",
		"question": "Create a visual direction for the homepage",
	})
	if err != nil {
		t.Fatalf("handleConsultCouncil: %v", err)
	}

	var parsed struct {
		Message   string `json:"message"`
		Artifacts []struct {
			ID    string `json:"id"`
			Type  string `json:"type"`
			Title string `json:"title"`
		} `json:"artifacts"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("expected structured consult output JSON, got err: %v output: %s", err, out)
	}
	if parsed.Message != "I prepared a visual direction and a concept note." {
		t.Fatalf("message = %q", parsed.Message)
	}
	if len(parsed.Artifacts) != 2 {
		t.Fatalf("artifacts len = %d, want 2", len(parsed.Artifacts))
	}
	if parsed.Artifacts[0].Type != "image" || parsed.Artifacts[1].Type != "document" {
		t.Fatalf("unexpected artifacts: %+v", parsed.Artifacts)
	}
}

func TestHandleConsultCouncil_PreservesPlainTextFallback(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	if _, err := nc.Subscribe("swarm.council.council-architect.request", func(msg *nats.Msg) {
		if msg.Reply != "" {
			_ = nc.Publish(msg.Reply, []byte("Use a simpler bounded architecture."))
		}
	}); err != nil {
		t.Fatalf("subscribe council request: %v", err)
	}
	nc.Flush()

	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc})

	out, err := reg.handleConsultCouncil(context.Background(), map[string]any{
		"member":   "council-architect",
		"question": "What should we simplify?",
	})
	if err != nil {
		t.Fatalf("handleConsultCouncil: %v", err)
	}
	if out != "Use a simpler bounded architecture." {
		t.Fatalf("out = %q", out)
	}
}

func TestHandleConsultCouncil_RequiresCouncilResponse(t *testing.T) {
	opts := server.DefaultTestOptions
	opts.Port = -1
	s := server.RunServer(&opts)
	defer s.Shutdown()

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("connect nats: %v", err)
	}
	defer nc.Close()

	reg := NewInternalToolRegistry(InternalToolDeps{NC: nc})

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err = reg.handleConsultCouncil(ctx, map[string]any{
		"member":   "council-sentry",
		"question": "Review the risk posture",
	})
	if err == nil {
		t.Fatal("expected error when council member does not respond")
	}
}
