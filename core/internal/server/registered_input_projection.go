package server

import (
	"context"
	"fmt"
	"log"

	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

type registeredInputProjection struct {
	server *AdminServer
}

// StartRegisteredInputProjection buffers registered external/service/device
// ingress. It does not route raw traffic to teams; downstream agentry reads
// bounded buffer refs through Soma/governed APIs.
func StartRegisteredInputProjection(ctx context.Context, s *AdminServer) error {
	if s == nil || s.Inputs == nil {
		return fmt.Errorf("registered input projection requires input service")
	}
	if s.getDB() == nil {
		return fmt.Errorf("registered input projection requires database")
	}
	if s.NC == nil || !s.NC.IsConnected() {
		return fmt.Errorf("registered input projection requires connected NATS")
	}
	projection := &registeredInputProjection{server: s}
	sub, err := s.NC.Subscribe(protocol.TopicGlobalInputWild, projection.handleNATSMessage)
	if err != nil {
		return err
	}
	go func() {
		<-ctx.Done()
		_ = sub.Unsubscribe()
	}()
	return nil
}

func (p *registeredInputProjection) handleNATSMessage(msg *nats.Msg) {
	if msg == nil {
		return
	}
	source, event, matched, err := p.server.Inputs.RecordBusMessage(
		context.Background(),
		msg.Subject,
		msg.Data,
		map[string][]string(msg.Header),
	)
	if err != nil {
		log.Printf("registered input projection: failed to buffer %s: %v", msg.Subject, err)
		return
	}
	if matched {
		log.Printf(
			"registered input projection: buffered %s event %s channel %s",
			source.ID,
			event.EventID,
			event.ChannelKey,
		)
	}
}
