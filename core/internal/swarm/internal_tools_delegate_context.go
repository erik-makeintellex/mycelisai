package swarm

import "github.com/mycelis/core/pkg/protocol"

func askWithDelegateContext(args map[string]any, ask protocol.TeamAsk) protocol.TeamAsk {
	ask = ask.Normalize()
	ctxRaw, ok := args["context"].(map[string]any)
	if !ok || len(ctxRaw) == 0 {
		return ask
	}
	merged := map[string]any{}
	for key, value := range ask.Context {
		merged[key] = value
	}
	for key, value := range ctxRaw {
		merged[key] = value
	}
	ask.Context = merged
	return ask
}
