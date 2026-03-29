package swarm

import (
	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// SetSensorConfigs marks specific agents as sensor agents.
func (t *Team) SetSensorConfigs(configs map[string]SensorConfig) {
	t.sensorConfigs = configs
}

// SetToolDescriptions provides tool name→description pairs for agent prompt injection.
func (t *Team) SetToolDescriptions(descs map[string]string) {
	t.toolDescs = descs
}

// SetInternalTools provides the internal tool registry for runtime context building.
func (t *Team) SetInternalTools(tools *InternalToolRegistry) {
	t.internalTools = tools
}

// SetEventEmitter wires the V7 event emitter + run_id into this team.
func (t *Team) SetEventEmitter(emitter protocol.EventEmitter, runID string) {
	t.eventEmitter = emitter
	t.runID = runID
}

// SetConversationLogger wires the V7 conversation logger into this team.
func (t *Team) SetConversationLogger(logger protocol.ConversationLogger) {
	t.conversationLogger = logger
}

// SetMCPBinding provides the concrete CompositeToolExecutor and MCP lookup data.
func (t *Team) SetMCPBinding(exec *CompositeToolExecutor, serverNames map[uuid.UUID]string, toolDescs map[string]string) {
	t.compositeExec = exec
	t.mcpServerNames = serverNames
	t.mcpToolDescs = toolDescs
}
