package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/mycelis/core/pkg/pb/swarm"
)

// SensorType distinguishes internal agent acquisition modes.
type SensorType string

const (
	SensorTypeHTTP     SensorType = "http_poll"
	SensorTypeInternal SensorType = "internal"
)

// SensorConfig defines the polling behavior for a sensor agent.
type SensorConfig struct {
	Type     SensorType        `json:"type" yaml:"type"`
	Endpoint string            `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
	Interval time.Duration     `json:"interval,omitempty" yaml:"interval,omitempty"`
	Headers  map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}

// SensorAgent is a poll-based agent that fetches data from HTTP endpoints
// and publishes CTSEnvelopes to NATS. Unlike Agent (LLM inference), SensorAgent
// acquires data via HTTP polling. Runs as a goroutine inside the core binary.
type SensorAgent struct {
	Manifest protocol.AgentManifest
	Config   SensorConfig
	TeamID   string
	nc       *nats.Conn
	client   *http.Client
	ctx      context.Context
	cancel   context.CancelFunc
}

// NewSensorAgent creates a sensor agent from a manifest and sensor config.
func NewSensorAgent(ctx context.Context, manifest protocol.AgentManifest, config SensorConfig, teamID string, nc *nats.Conn) *SensorAgent {
	agentCtx, cancel := context.WithCancel(ctx)

	interval := config.Interval
	if interval == 0 {
		interval = 60 * time.Second
	}
	config.Interval = interval

	return &SensorAgent{
		Manifest: manifest,
		Config:   config,
		TeamID:   teamID,
		nc:       nc,
		client:   &http.Client{Timeout: 10 * time.Second},
		ctx:      agentCtx,
		cancel:   cancel,
	}
}

// Start begins the polling loop and heartbeat. Blocks until context is cancelled.
func (s *SensorAgent) Start() {
	log.Printf("SensorAgent [%s] (%s) Online — polling every %s", s.Manifest.ID, s.Manifest.Role, s.Config.Interval)

	go s.startHeartbeat()

	ticker := time.NewTicker(s.Config.Interval)
	defer ticker.Stop()

	// Initial poll on startup
	s.poll()

	for {
		select {
		case <-s.ctx.Done():
			log.Printf("SensorAgent [%s] shutting down.", s.Manifest.ID)
			return
		case <-ticker.C:
			s.poll()
		}
	}
}

// Stop cancels the sensor agent's context.
func (s *SensorAgent) Stop() {
	s.cancel()
}

// poll executes one data acquisition cycle and publishes results.
func (s *SensorAgent) poll() {
	var data []byte
	var err error

	if s.Config.Endpoint != "" && s.Config.Type == SensorTypeHTTP {
		data, err = s.fetchHTTP()
		if err != nil {
			log.Printf("SensorAgent [%s] poll error: %v", s.Manifest.ID, err)
			// Publish error telemetry instead of dropping
			s.publishTelemetry(protocol.SignalError, json.RawMessage(
				fmt.Sprintf(`{"error":"%s"}`, err.Error()),
			))
			return
		}
	} else {
		// Heartbeat-only mode: no endpoint configured
		data = []byte(`{"status":"online","mode":"heartbeat_only"}`)
	}

	// Publish to telemetry topic (Overseer-compatible CTS envelope)
	s.publishTelemetry(protocol.SignalSensorData, json.RawMessage(data))

	// Publish raw data to agent output topics for downstream consumers
	for _, topic := range s.Manifest.Outputs {
		if err := s.nc.Publish(topic, data); err != nil {
			log.Printf("SensorAgent [%s] publish to %s error: %v", s.Manifest.ID, topic, err)
		}
	}
}

// fetchHTTP performs an HTTP GET to the configured endpoint.
func (s *SensorAgent) fetchHTTP() ([]byte, error) {
	req, err := http.NewRequestWithContext(s.ctx, http.MethodGet, s.Config.Endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	for k, v := range s.Config.Headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http get: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20)) // 1MB limit
	if err != nil {
		return nil, fmt.Errorf("read body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("http %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}

// publishTelemetry wraps data in a CTSEnvelope and publishes to the team telemetry topic.
func (s *SensorAgent) publishTelemetry(sigType protocol.SignalType, payload json.RawMessage) {
	env := protocol.CTSEnvelope{
		Meta: protocol.CTSMeta{
			SourceNode: s.Manifest.ID,
			Timestamp:  time.Now(),
			TraceID:    uuid.New().String(),
		},
		SignalType: sigType,
		TrustScore: protocol.TrustScoreSensory, // 1.0 — sensors are fully trusted
		Payload:    payload,
	}

	data, err := json.Marshal(env)
	if err != nil {
		log.Printf("SensorAgent [%s] marshal CTS error: %v", s.Manifest.ID, err)
		return
	}

	topic := fmt.Sprintf(protocol.TopicTeamTelemetryFmt, s.TeamID)
	if err := s.nc.Publish(topic, data); err != nil {
		log.Printf("SensorAgent [%s] publish telemetry error: %v", s.Manifest.ID, err)
	}
}

// startHeartbeat publishes protobuf heartbeats on the global heartbeat topic.
// Follows the same pattern as Agent.StartHeartbeat (agent.go).
func (s *SensorAgent) startHeartbeat() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			env := &pb.MsgEnvelope{
				Id:            uuid.New().String(),
				Timestamp:     timestamppb.Now(),
				SourceAgentId: s.Manifest.ID,
				Type:          pb.MessageType_MESSAGE_TYPE_EVENT,
				TeamId:        s.TeamID,
				Payload: &pb.MsgEnvelope_Event{
					Event: &pb.EventPayload{
						EventType: "agent.heartbeat",
					},
				},
			}

			data, err := proto.Marshal(env)
			if err != nil {
				continue
			}

			s.nc.Publish(protocol.TopicGlobalHeartbeat, data)
		}
	}
}
