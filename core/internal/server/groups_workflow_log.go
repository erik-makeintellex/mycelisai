package server

import (
	"context"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

const maxGroupWorkflowLogLimit = 100

type groupWorkflowLogResponse struct {
	Group          CollaborationGroup        `json:"group"`
	Lifecycle      groupLifecycleItem        `json:"lifecycle"`
	GeneratedAt    time.Time                 `json:"generated_at"`
	Limit          int                       `json:"limit"`
	IncludeOutputs bool                      `json:"include_outputs"`
	IncludeAudit   bool                      `json:"include_audit"`
	Timeline       []groupWorkflowLogEntry   `json:"timeline"`
	TeamWork       []protocol.TeamWorkItem   `json:"team_work"`
	Outputs        []artifacts.Artifact      `json:"outputs,omitempty"`
	Degraded       []groupWorkflowLogDegrade `json:"degraded,omitempty"`
}

type groupWorkflowLogEntry struct {
	ID             string                   `json:"id"`
	Kind           string                   `json:"kind"`
	GroupID        string                   `json:"group_id"`
	TeamID         string                   `json:"team_id,omitempty"`
	WorkItemID     string                   `json:"work_item_id,omitempty"`
	RunID          string                   `json:"run_id,omitempty"`
	ArtifactID     string                   `json:"artifact_id,omitempty"`
	Title          string                   `json:"title"`
	Summary        string                   `json:"summary,omitempty"`
	State          string                   `json:"state,omitempty"`
	TrustPosture   string                   `json:"trust_posture,omitempty"`
	StorageRef     string                   `json:"storage_ref,omitempty"`
	Entrypoint     string                   `json:"entrypoint,omitempty"`
	ProofRefs      []string                 `json:"proof_refs,omitempty"`
	AuditRefs      []string                 `json:"audit_refs,omitempty"`
	OutputRefs     []protocol.TeamOutputRef `json:"output_refs,omitempty"`
	Recovery       []string                 `json:"recovery,omitempty"`
	NeedsOperator  bool                     `json:"needs_operator,omitempty"`
	Timestamp      time.Time                `json:"timestamp"`
	SourceKind     string                   `json:"source_kind,omitempty"`
	SourceChannel  string                   `json:"source_channel,omitempty"`
	PayloadKind    string                   `json:"payload_kind,omitempty"`
	CorrelationRef string                   `json:"correlation_ref,omitempty"`
}

type groupWorkflowLogDegrade struct {
	Kind       string    `json:"kind"`
	Message    string    `json:"message"`
	Trusted    bool      `json:"trusted"`
	Timestamp  time.Time `json:"timestamp"`
	NextAction string    `json:"next_action,omitempty"`
}

// GET /api/v1/groups/{id}/workflow-log
func (s *AdminServer) HandleGroupWorkflowLog(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:read"); !ok {
		return
	}

	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "Missing group ID", http.StatusBadRequest)
		return
	}

	group, err := s.getGroupDB(r.Context(), id)
	if err != nil {
		respondAPIError(w, "Failed to load group: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if group == nil {
		respondAPIError(w, "Group not found", http.StatusNotFound)
		return
	}

	log, err := s.buildGroupWorkflowLog(r.Context(), group, groupWorkflowLogOptions{
		Limit:          parseGroupWorkflowLogLimit(r.URL.Query().Get("limit"), 50),
		IncludeOutputs: parseBoolDefault(r.URL.Query().Get("include_outputs"), true),
		IncludeAudit:   parseBoolDefault(r.URL.Query().Get("include_audit"), true),
	})
	if err != nil {
		respondAPIError(w, "Failed to build group workflow log: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(log))
}

type groupWorkflowLogOptions struct {
	Limit          int
	IncludeOutputs bool
	IncludeAudit   bool
}

func (s *AdminServer) buildGroupWorkflowLog(ctx context.Context, group *CollaborationGroup, opts groupWorkflowLogOptions) (groupWorkflowLogResponse, error) {
	if group == nil {
		return groupWorkflowLogResponse{}, fmt.Errorf("group is required")
	}
	opts.Limit = normalizeGroupWorkflowLogLimit(opts.Limit)
	now := time.Now().UTC()

	teamWork, err := s.listGroupTeamWorkItems(ctx, group.TeamIDs, opts.Limit)
	if err != nil {
		return groupWorkflowLogResponse{}, err
	}

	outputs := []artifacts.Artifact{}
	degraded := []groupWorkflowLogDegrade{}
	if opts.IncludeOutputs {
		loadedOutputs, outputErr := s.listGroupOutputs(ctx, group, opts.Limit)
		if outputErr != nil {
			degraded = append(degraded, groupWorkflowLogDegrade{
				Kind:       "outputs_unavailable",
				Message:    outputErr.Error(),
				Trusted:    false,
				Timestamp:  now,
				NextAction: "Review output storage or artifact service health, then retry this workflow log.",
			})
		} else if loadedOutputs != nil {
			outputs = loadedOutputs
		}
	}

	stats := groupTeamWorkStatsFromItems(teamWork)
	lifecycle := buildGroupLifecycleItem(*group, stats, len(outputs), now)
	entries := make([]groupWorkflowLogEntry, 0, 2+len(teamWork)*3+len(outputs)+len(degraded))
	entries = append(entries, groupBriefWorkflowLogEntry(*group))
	entries = append(entries, groupLifecycleWorkflowLogEntry(*group, lifecycle, now))
	entries = append(entries, s.groupBroadcastWorkflowLogEntries(*group)...)

	for _, item := range teamWork {
		entries = append(entries, teamWorkWorkflowLogEntry(*group, item, opts.IncludeAudit))
		if opts.IncludeOutputs {
			entries = append(entries, teamWorkOutputWorkflowLogEntries(*group, item, opts.IncludeAudit)...)
		}
		entries = append(entries, teamWorkProofWorkflowLogEntries(*group, item, opts.IncludeAudit)...)
	}
	if opts.IncludeOutputs {
		for _, artifact := range outputs {
			entries = append(entries, artifactWorkflowLogEntry(*group, artifact, opts.IncludeAudit))
		}
	}
	for _, item := range degraded {
		entries = append(entries, degradedWorkflowLogEntry(*group, item))
	}

	sort.SliceStable(entries, func(i, j int) bool {
		return entries[i].Timestamp.After(entries[j].Timestamp)
	})
	if len(entries) > opts.Limit {
		entries = entries[:opts.Limit]
	}

	return groupWorkflowLogResponse{
		Group:          *group,
		Lifecycle:      lifecycle,
		GeneratedAt:    now,
		Limit:          opts.Limit,
		IncludeOutputs: opts.IncludeOutputs,
		IncludeAudit:   opts.IncludeAudit,
		Timeline:       entries,
		TeamWork:       teamWork,
		Outputs:        outputs,
		Degraded:       degraded,
	}, nil
}

func parseGroupWorkflowLogLimit(raw string, fallback int) int {
	return normalizeGroupWorkflowLogLimit(parseLimit(raw, fallback))
}

func normalizeGroupWorkflowLogLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	if limit > maxGroupWorkflowLogLimit {
		return maxGroupWorkflowLogLimit
	}
	return limit
}

func (s *AdminServer) listGroupTeamWorkItems(ctx context.Context, teamIDs []string, limit int) ([]protocol.TeamWorkItem, error) {
	items := make([]protocol.TeamWorkItem, 0)
	for _, rawTeamID := range teamIDs {
		teamID := strings.TrimSpace(rawTeamID)
		if teamID == "" {
			continue
		}
		teamItems, err := s.listTeamWorkItemsDB(ctx, teamID, limit, false)
		if err != nil {
			return nil, err
		}
		items = append(items, teamItems...)
	}
	sort.SliceStable(items, func(i, j int) bool {
		return items[i].UpdatedAt.After(items[j].UpdatedAt)
	})
	if len(items) > limit {
		items = items[:limit]
	}
	return items, nil
}

func groupTeamWorkStatsFromItems(items []protocol.TeamWorkItem) groupTeamWorkStats {
	var stats groupTeamWorkStats
	for _, item := range items {
		stats.Total++
		switch item.State {
		case protocol.TeamWorkStateOutputReady:
			stats.OutputReady++
		case protocol.TeamWorkStateArchived:
			stats.Archived++
		default:
			stats.ActiveOrBlocked++
		}
		if !item.UpdatedAt.IsZero() && (stats.LatestWorkAt == nil || item.UpdatedAt.After(*stats.LatestWorkAt)) {
			latest := item.UpdatedAt.UTC()
			stats.LatestWorkAt = &latest
		}
	}
	return stats
}
