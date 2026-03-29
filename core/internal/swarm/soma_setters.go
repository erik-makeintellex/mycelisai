package swarm

import (
	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *Soma) SetRunsManager(rm protocol.RunsManager) { s.runsManager = rm }

func (s *Soma) SetEventEmitter(emitter protocol.EventEmitter) { s.eventEmitter = emitter }

func (s *Soma) SetConversationLogger(logger protocol.ConversationLogger) {
	s.conversationLogger = logger
}

func (s *Soma) SetMCPServerNames(names map[uuid.UUID]string) {
	s.mcpServerNames = names
}

func (s *Soma) SetMCPToolDescs(descs map[string]string) {
	s.mcpToolDescs = descs
}

func (s *Soma) SetProviderPolicy(policy ProviderPolicy) {
	s.providerPolicy = policy.Clone()
}
