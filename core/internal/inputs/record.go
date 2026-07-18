package inputs

import (
	"context"
	"strings"
)

func (s *Service) RecordNATSMessage(ctx context.Context, subject string, data []byte) (Source, BufferEvent, bool, error) {
	return s.RecordBusMessage(ctx, subject, data, nil)
}

func (s *Service) RecordBusMessage(
	ctx context.Context,
	subject string,
	data []byte,
	headers map[string][]string,
) (Source, BufferEvent, bool, error) {
	if s == nil {
		return Source{}, BufferEvent{}, false, ErrUnavailable
	}
	source, ok := s.sourceForSubject(subject)
	if !ok {
		return Source{}, BufferEvent{}, false, nil
	}
	event, err := EventFromBusMessage(source, subject, data, headers)
	if err != nil {
		return source, BufferEvent{}, true, err
	}
	if s.store == nil {
		return source, bufferEventFromIngest(event), true, nil
	}
	recorded, err := s.store.RecordEvent(ctx, source, event)
	return source, recorded, true, err
}

func (s *Service) sourceForSubject(subject string) (Source, bool) {
	subject = strings.TrimSpace(subject)
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, source := range s.sources {
		if source.Status == StatusAvailable && source.AllowedIngressSubject == subject {
			return source, true
		}
	}
	return Source{}, false
}

func bufferEventFromIngest(event IngestEvent) BufferEvent {
	return BufferEvent{
		SourceID:        event.SourceID,
		ChannelKey:      event.ChannelKey,
		Payload:         event.Payload,
		PayloadHash:     event.PayloadHash,
		SourceTimestamp: event.SourceTimestamp,
		RunID:           event.RunID,
		TeamID:          event.TeamID,
		AgentID:         event.AgentID,
		SourceKind:      event.SourceKind,
		SourceChannel:   event.SourceChannel,
		PayloadKind:     event.PayloadKind,
		TenantID:        event.TenantID,
	}
}
