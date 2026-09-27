package server

import (
	"net/http"
	"time"

	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

type teamHandoffChainStep struct {
	Step     string    `json:"step"`
	At       time.Time `json:"at"`
	Evidence string    `json:"evidence,omitempty"`
}

// teamHandoffView is one row of the Exchange "Team handoffs" list. Status is
// derived from the receiving team's work item and interactions, never from
// the handoff record alone.
type teamHandoffView struct {
	HandoffID      string                  `json:"handoff_id"`
	RunID          string                  `json:"run_id"`
	SourceTeamID   string                  `json:"source_team_id"`
	SourceAgentID  string                  `json:"source_agent_id"`
	TargetTeamID   string                  `json:"target_team_id"`
	Note           string                  `json:"note"`
	ExpectedAction string                  `json:"expected_action"`
	Inputs         []swarm.HandoffInputRef `json:"inputs"`
	WorkItemID     string                  `json:"work_item_id"`
	WorkState      protocol.TeamWorkState  `json:"work_state,omitempty"`
	Status         string                  `json:"status"`
	Chain          []teamHandoffChainStep  `json:"chain"`
	CreatedAt      time.Time               `json:"created_at"`
}

// teamHandoffStatus maps the receiving work item to the operator statuses:
// queued, accepted, read, completed, needs_attention (or closed).
func teamHandoffStatus(item protocol.TeamWorkItem, interactions []protocol.TeamInteraction) string {
	switch item.State {
	case protocol.TeamWorkStateOutputReady:
		return "completed"
	case protocol.TeamWorkStateDegraded, protocol.TeamWorkStateNeedsOperator, protocol.TeamWorkStatePaused:
		return "needs_attention"
	case protocol.TeamWorkStateArchived:
		return "closed"
	}
	if hasInteractionVerb(interactions, teamHandoffVerbInputRead) {
		return "read"
	}
	if item.State == protocol.TeamWorkStateQueued || item.State == "" {
		if hasInteractionVerb(interactions, teamHandoffVerbFailed) {
			return "needs_attention"
		}
		return "queued"
	}
	return "accepted"
}

func hasInteractionVerb(interactions []protocol.TeamInteraction, verb string) bool {
	_, ok := firstInteraction(interactions, verb)
	return ok
}

func firstInteraction(interactions []protocol.TeamInteraction, verb string) (protocol.TeamInteraction, bool) {
	for _, interaction := range interactions {
		if interaction.Verb == verb {
			return interaction, true
		}
	}
	return protocol.TeamInteraction{}, false
}

// teamHandoffChain lists only steps with evidence: created (the HandoffNote),
// queued (handoff_queued at commit), handed (handoff_sent, recorded only after
// the notice was published) and acknowledged (the receiving team
// read an input or accepted the work).
func teamHandoffChain(record teamHandoffRecord, item protocol.TeamWorkItem, interactions []protocol.TeamInteraction) []teamHandoffChainStep {
	chain := []teamHandoffChainStep{{Step: "created", At: record.CreatedAt, Evidence: "exchange:" + teamHandoffChannel}}
	if queued, ok := firstInteraction(interactions, teamHandoffVerbQueued); ok {
		chain = append(chain, teamHandoffChainStep{Step: "queued", At: queued.Timestamp, Evidence: teamHandoffVerbQueued})
	}
	if sent, ok := firstInteraction(interactions, teamHandoffVerbSent); ok {
		chain = append(chain, teamHandoffChainStep{Step: "handed", At: sent.Timestamp, Evidence: teamHandoffVerbSent})
	}
	if read, ok := firstInteraction(interactions, teamHandoffVerbInputRead); ok {
		chain = append(chain, teamHandoffChainStep{Step: "acknowledged", At: read.Timestamp, Evidence: teamHandoffVerbInputRead})
	} else if item.State != "" && item.State != protocol.TeamWorkStateQueued {
		chain = append(chain, teamHandoffChainStep{Step: "acknowledged", At: item.UpdatedAt, Evidence: "team_accepted"})
	}
	return chain
}

// listTeamHandoffs serves GET /api/v1/exchange/handoffs.
func listTeamHandoffs(store teamHandoffStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		records, err := store.ListHandoffs(r.Context(), parsePositiveInt(r.URL.Query().Get("limit"), 50))
		if err != nil {
			respondAPIError(w, "Failed to list team handoffs: "+err.Error(), http.StatusInternalServerError)
			return
		}
		views := make([]teamHandoffView, 0, len(records))
		for _, record := range records {
			view := teamHandoffView{
				HandoffID: record.HandoffID, RunID: record.RunID, SourceTeamID: record.SourceTeamID, SourceAgentID: record.SourceAgentID,
				TargetTeamID: record.TargetTeamID, Note: record.Note, ExpectedAction: record.ExpectedAction, Inputs: record.Inputs,
				WorkItemID: record.WorkItemID, CreatedAt: record.CreatedAt, Status: "needs_attention",
			}
			item, itemErr := store.WorkItem(r.Context(), record.TargetTeamID, record.WorkItemID)
			interactions, interactionErr := store.Interactions(r.Context(), record.TargetTeamID, record.WorkItemID)
			if itemErr == nil && interactionErr == nil {
				view.WorkState = item.State
				view.Status = teamHandoffStatus(item, interactions)
				view.Chain = teamHandoffChain(record, item, interactions)
			} else {
				// The receiving work cannot be read back: show the record only.
				view.Chain = teamHandoffChain(record, protocol.TeamWorkItem{}, nil)
			}
			views = append(views, view)
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(views))
	}
}

// handleListTeamHandoffs uses the Work-surface gate: root admin + groups:read.
func (s *AdminServer) handleListTeamHandoffs(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:read"); !ok {
		return
	}
	listTeamHandoffs(&sqlTeamHandoffStore{server: s})(w, r)
}
